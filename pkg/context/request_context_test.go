package context

import (
"context"
"testing"
"time"
)

func TestNewRequestContext(t *testing.T) {
reqCtx := NewRequestContext()

if reqCtx == nil {
t.Fatal("NewRequestContext returned nil")
}

if reqCtx.RequestID == "" {
t.Error("RequestID should not be empty")
}

if reqCtx.StartTime.IsZero() {
t.Error("StartTime should not be zero")
}

// Verify UUID format (basic check)
if len(reqCtx.RequestID) != 36 {
t.Errorf("RequestID format unexpected: %s", reqCtx.RequestID)
}
}

func TestWithRequestContext(t *testing.T) {
reqCtx := NewRequestContext()
ctx := context.Background()

ctxWithReq := WithRequestContext(ctx, reqCtx)

retrieved := FromContext(ctxWithReq)
if retrieved == nil {
t.Fatal("FromContext returned nil")
}

if retrieved.RequestID != reqCtx.RequestID {
t.Errorf("RequestID mismatch: got %s, want %s", retrieved.RequestID, reqCtx.RequestID)
}
}

func TestGetRequestID(t *testing.T) {
tests := []struct {
name string
ctx  context.Context
want string
}{
{
name: "with request context",
ctx:  WithRequestContext(context.Background(), NewRequestContext()),
want: "", // will be non-empty
},
{
name: "without request context",
ctx:  context.Background(),
want: "",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := GetRequestID(tt.ctx)
if tt.name == "with request context" {
if got == "" {
t.Error("GetRequestID returned empty string with request context")
}
} else {
if got != "" {
t.Errorf("GetRequestID returned %s, want empty string", got)
}
}
})
}
}

func TestGetElapsedTime(t *testing.T) {
reqCtx := NewRequestContext()
ctx := WithRequestContext(context.Background(), reqCtx)

// Sleep a bit to ensure measurable time
time.Sleep(10 * time.Millisecond)

elapsed := GetElapsedTime(ctx)
if elapsed < 10*time.Millisecond {
t.Errorf("Elapsed time too small: %v", elapsed)
}

// Test with context without request context
elapsed = GetElapsedTime(context.Background())
if elapsed != 0 {
t.Errorf("GetElapsedTime returned %v for context without request context", elapsed)
}
}
