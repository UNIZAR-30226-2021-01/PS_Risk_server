package test

import (
	"PS_Risk_server/server"
	"testing"
)

func TestHello(t *testing.T) {
	want := "Hello, world."
	if got := server.Hello(); got != want {
		t.Errorf("Hello() = %q, want %q", got, want)
	}
}
