package access

import (
	"net"
	"github.com/go-log/log"
	"os"
	"os/signal"
	"syscall"
)

type tcpServer struct {
	addr string
	l net.Listener
	exit chan chan error
}

func (s *tcpServer) run() error {
	if err := s.start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	log.Logf("Received signal %s", <-ch)

	return s.stop()
}

func (s *tcpServer) start() error {
	addr, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.l = l

	log.Log("listen success. addr:", addr)

	go s.accept()

	go func() {
		ch := <-s.exit

		ch <- s.l.Close() //关闭socket
	}()

	return nil
}

func (s *tcpServer) accept() error {

	return nil
}

func (s *tcpServer) stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}
