//go:build unix

package eventstore

import (
	"os"
	"syscall"
)

// lockDirectory 对 dir 下的 lockfile 获取非阻塞排他锁。
// 锁随 fd 关闭自动释放，进程崩溃后不会残留，因此重启仍可恢复。
func lockDirectory(dir, lockName string) (*os.File, error) {
	path := dir + string(os.PathSeparator) + lockName
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o640)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}
