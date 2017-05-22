package fmtp_test

import (
	"context"
	"testing"
)

func TestConnect_Online(t *testing.T) {
	ctx := context.Background()
	_, err := defaultClient.Connect(ctx, "127.0.0.1:9050", "remoteID")
	if err != nil {
		t.Fatalf("connection failed: %s", err)
	}
}
