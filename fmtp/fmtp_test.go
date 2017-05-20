package fmtp_test

import (
	"context"
	"testing"

	"github.com/aabizri/aero/fmtp"
)

var defaultClient *fmtp.Client

func init() {
	TestMain(&testing.T{})
}

func TestMain(t *testing.T) {
	c, err := fmtp.NewClient("localID")
	if err != nil {
		t.Fatalf("error while creating new client: %s", err)
	}
	defaultClient = c
}

/*func TestListen(t *testing.T) {
	//ctx := context.Background()
	server := defaultClient.NewServer("127.0.0.1:9050", nil)
	t.Fatal(server.ListenAndServe())
}*/

func TestConnect(t *testing.T) {
	ctx := context.Background()
	_, err := defaultClient.Connect(ctx, "127.0.0.1:9050", "remoteID")
	if err != nil {
		t.Fatalf("connection failed: %s", err)
	}
}
