package tcpServer

import (
	"net"
	"testing"
)

func TestTcpServer_Run(t *testing.T) {
	err,listener := net.ListenTCP("tcp",)
	s := NewTcpServer()
}