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

// IdempotencyPayloadConflictError 表示同一案卷复用幂等键提交了不同的操作、资源、期望版本或业务载荷。
type IdempotencyPayloadConflictError struct {
	Key string
}

func (e *IdempotencyPayloadConflictError) Error() string {
	return "idempotencyKey 已用于不同的请求内容，无法复用: " + e.Key
}
