package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const defaultAddress = "127.0.0.1:19081"

type config struct {
	Address   string
	DataDir   string
	SelfCheck bool
}

func parseConfig(arguments []string) (config, error) {
	flags := flag.NewFlagSet("heritage-tree-relocation-clearance", flag.ContinueOnError)
	address := flags.String("addr", "", "监听地址，仅允许显式回环地址")
	dataDir := flags.String("data-dir", "./data", "事件存储目录")
	selfcheck := flags.Bool("selfcheck", false, "执行完整 HTTP 自检后退出")
	if err := flags.Parse(arguments); err != nil {
		return config{}, err
	}
	if flags.NArg() != 0 {
		return config{}, errors.New("存在无法识别的位置参数")
	}
	addressSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "addr" {
			addressSet = true
		}
	})
	resolved := defaultAddress
	if addressSet {
		if strings.TrimSpace(*address) == "" {
			return config{}, errors.New("-addr 不得为空")
		}
		resolved = *address
	} else if portText := strings.TrimSpace(os.Getenv("PORT")); portText != "" {
		port, err := strconv.Atoi(portText)
		if err != nil || port < 1024 || port > 65535 {
			return config{}, errors.New("PORT 必须为 1024 至 65535 的十进制端口号")
		}
		resolved = net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	}
	if err := validateLoopbackAddress(resolved); err != nil {
		return config{}, err
	}
	if strings.TrimSpace(*dataDir) == "" {
		return config{}, errors.New("-data-dir 不得为空")
	}
	return config{Address: resolved, DataDir: *dataDir, SelfCheck: *selfcheck}, nil
}

func validateLoopbackAddress(address string) error {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("监听地址必须为 host:port: %w", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1024 || port > 65535 {
		return errors.New("监听端口必须处于 1024 至 65535 之间")
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("监听地址必须是明确的回环地址，拒绝非回环绑定")
	}
	return nil
}
