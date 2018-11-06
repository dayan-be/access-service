package tcp_server

import (
	"net"
	"testing"
)

func TestTcpServer_Run(t *testing.T) {
	err, listener := net.ListenTCP("tcp")
	s := NewTcpServer()
}
