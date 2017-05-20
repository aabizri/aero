package fmtp_test

import (
	"bytes"
	"testing"

	"io/ioutil"

	"github.com/aabizri/aero/fmtp"
)

func TestMessageOperatorFromString_WriteTo(t *testing.T) {
	msg, err := fmtp.NewOperatorMessageFromString("hyper-lol")
	if err != nil {
		t.Fatalf("error creating operator msg: %v", err)
	}
	buf := &bytes.Buffer{}
	n, err := msg.WriteTo(buf)
	if err != nil {
		t.Fatalf("error in WriteTo (wrote %d bytes): %s", n, err)
	}
	b, _ := ioutil.ReadAll(buf)
	t.Logf("Wrote %d bytes:\n%v", n, b)
}
