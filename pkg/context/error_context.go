package context

import (
"context"
"fmt"
"time"
)

// ErrorContext contains structured metadata for error messages
type ErrorContext struct {
RequestID    string
Operation    string
Path         string
SecretName   string
DurationMs   int64
RetryCount   int
StartedAt    time.Time
FailedAt     time.Time
}

// ErrorBuilder helps construct errors with rich context
type ErrorBuilder struct {
ctx       context.Context
operation string
path      string
secretName string
retryCount int
startTime time.Time
}

// NewErrorBuilder creates a new error builder with context
func NewErrorBuilder(ctx context.Context, operation string) *ErrorBuilder {
return &ErrorBuilder{
ctx:       ctx,
operation: operation,
startTime: time.Now(),
}
}

// WithPath sets the path field
func (b *ErrorBuilder) WithPath(path string) *ErrorBuilder {
b.path = path
return b
}

// WithSecretName sets the secret name field
func (b *ErrorBuilder) WithSecretName(name string) *ErrorBuilder {
b.secretName = name
return b
}

// WithRetryCount sets the retry count field
func (b *ErrorBuilder) WithRetryCount(count int) *ErrorBuilder {
b.retryCount = count
return b
}

// Build creates an error with full context
func (b *ErrorBuilder) Build(message string, err error) error {
duration := time.Since(b.startTime)
requestID := GetRequestID(b.ctx)

// Build error message with available context
var errMsg string
if requestID != "" {
errMsg = fmt.Sprintf("[req=%s]", requestID)
}

if b.operation != "" {
errMsg += fmt.Sprintf(" operation=%s", b.operation)
}

if b.path != "" {
errMsg += fmt.Sprintf(" path=%q", b.path)
}

if b.secretName != "" {
errMsg += fmt.Sprintf(" secret=%q", b.secretName)
}

if b.retryCount > 0 {
errMsg += fmt.Sprintf(" retries=%d", b.retryCount)
}

errMsg += fmt.Sprintf(" duration=%dms", duration.Milliseconds())

if errMsg != "" {
errMsg += ": "
}

errMsg += message

if err != nil {
return fmt.Errorf("%s: %w", errMsg, err)
}

return fmt.Errorf("%s", errMsg)
}

// Errorf creates an error with formatted message
func (b *ErrorBuilder) Errorf(format string, args ...interface{}) error {
message := fmt.Sprintf(format, args...)
return b.Build(message, nil)
}

// Wrap wraps an existing error with context
func (b *ErrorBuilder) Wrap(err error, message string) error {
return b.Build(message, err)
}

// GetContext extracts ErrorContext from builder state
func (b *ErrorBuilder) GetContext() ErrorContext {
duration := time.Since(b.startTime)
return ErrorContext{
RequestID:   GetRequestID(b.ctx),
Operation:   b.operation,
Path:        b.path,
SecretName:  b.secretName,
DurationMs:  duration.Milliseconds(),
RetryCount:  b.retryCount,
StartedAt:   b.startTime,
FailedAt:    time.Now(),
}
}
