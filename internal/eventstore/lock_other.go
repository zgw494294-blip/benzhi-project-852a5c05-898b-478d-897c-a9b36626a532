//go:build !unix

package eventstore

import (
	"errors"
	"os"
)

// lockDirectory 在不支持 flock 的平台上退化为独占创建。
// 这些平台不是项目目标运行环境；恢复能力在 unix 上有保证。
func lockDirectory(dir, lockName string) (*os.File, error) {
	path := dir + string(os.PathSeparator) + lockName
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o640)
	if err != nil {
		if os.IsExist(err) {
			return nil, errors.New("事件存储目录已被其他实例锁定")
		}
		return nil, err
	}
	return file, nil
}
