package poll_test

import (
	"context"
	"testing"
	"time"

	"github.com/fun7257/xai-sdk-go/internal/poll"
)

func TestWaitSuccess(t *testing.T) {
	n := 0
	err := poll.Wait(context.Background(), poll.Config{Timeout: time.Second, Interval: time.Millisecond}, func(context.Context) (bool, error) {
		n++
		return n >= 2, nil
	})
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestWaitTimeout(t *testing.T) {
	err := poll.Wait(context.Background(), poll.Config{Timeout: 15 * time.Millisecond, Interval: 5 * time.Millisecond, Context: "x"}, func(context.Context) (bool, error) {
		return false, nil
	})
	if err == nil {
		t.Fatal("expected timeout")
	}
}
