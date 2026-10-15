//go:build linux || windows

package tcpinfo

import (
	"net"
	"testing"
)

func TestGetRTT(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	rtt, err := GetRTT(conn.(*net.TCPConn))
	if err != nil {
		t.Fatalf("GetRTT failed: %s", err)
	}
	if rtt < 0 {
		t.Fatalf("expected non-negative rtt, got %s", rtt)
	}
	t.Logf("rtt=%s", rtt)
}

func TestGetRTTNilConn(t *testing.T) {
	if _, err := GetRTT(nil); err == nil {
		t.Fatal("expected error for nil conn")
	}
}
