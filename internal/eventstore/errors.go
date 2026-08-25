package eventstore

import "fmt"

type VersionConflictError struct{ Expected, Actual uint64 }

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("expectedVersion=%d 与当前版本 %d 不一致", e.Expected, e.Actual)
}

type CorruptionError struct {
	Offset int64
	Reason string
}

func (e *CorruptionError) Error() string {
	return fmt.Sprintf("事件日志在偏移 %d 处损坏：%s", e.Offset, e.Reason)
}

type IdempotencyConflictError struct{ Key string }

func (e *IdempotencyConflictError) Error() string {
	return "idempotencyKey 已被其他案卷使用: " + e.Key
}
