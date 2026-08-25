package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/eventstore"
	"heritage-tree-relocation-clearance/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(os.Args[1:], logger); err != nil {
		logger.Error("服务退出", "error", err)
		os.Exit(1)
	}
}

func run(arguments []string, logger *slog.Logger) error {
	configuration, err := parseConfig(arguments)
	if err != nil {
		return fmt.Errorf("解析配置: %w", err)
	}
	dataDir := configuration.DataDir
	if configuration.SelfCheck {
		dataDir, err = os.MkdirTemp("", "heritage-relocation-selfcheck-*")
		if err != nil {
			return fmt.Errorf("创建自检存储: %w", err)
		}
		defer os.RemoveAll(dataDir)
	}
	store, err := eventstore.Open(dataDir)
	if err != nil {
		return fmt.Errorf("打开事件存储: %w", err)
	}
	defer store.Close()
	service := application.NewService(store)
	api := httpapi.New(service, logger)
	listener, err := net.Listen("tcp", configuration.Address)
	if err != nil {
		return fmt.Errorf("监听 %s: %w", configuration.Address, err)
	}
	server := &http.Server{Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	serveErrors := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			serveErrors <- err
			return
		}
		serveErrors <- nil
	}()
	logger.Info("古树移栽审查服务已启动", "address", listener.Addr().String(), "dataDir", dataDir, "selfcheck", configuration.SelfCheck)
	if configuration.SelfCheck {
		return runBoundedSelfCheck(server, listener, serveErrors)
	}
	return waitForShutdown(server, serveErrors, logger)
}

func runBoundedSelfCheck(server *http.Server, listener net.Listener, serveErrors <-chan error) error {
	checkContext, cancelCheck := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelCheck()
	baseURL := "http://" + listener.Addr().String()
	checkErr := httpapi.RunSelfCheck(checkContext, baseURL)
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 4*time.Second)
	shutdownErr := server.Shutdown(shutdownContext)
	cancelShutdown()
	serveErr := <-serveErrors
	if checkErr != nil {
		return fmt.Errorf("selfcheck 失败: %w", checkErr)
	}
	if shutdownErr != nil {
		return fmt.Errorf("selfcheck 关闭服务: %w", shutdownErr)
	}
	if serveErr != nil {
		return fmt.Errorf("selfcheck HTTP 服务: %w", serveErr)
	}
	fmt.Println("selfcheck: 古树移栽阻断整改与凭据签发流程通过")
	return nil
}

func waitForShutdown(server *http.Server, serveErrors <-chan error, logger *slog.Logger) error {
	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serveErrors:
		if err != nil {
			return fmt.Errorf("HTTP 服务异常: %w", err)
		}
		return nil
	case <-signalContext.Done():
		logger.Info("收到关闭信号，开始排空请求")
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("优雅关闭超时: %w", err)
	}
	if err := <-serveErrors; err != nil {
		return fmt.Errorf("HTTP 服务关闭: %w", err)
	}
	return nil
}
