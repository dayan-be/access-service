package logic

import (
	"net"
	"sync"
	"github.com/go-log/log"
	"time"
	"runtime/debug"
	"github.com/dayan-be/access-service/proto"
	"context"
)

type Server struct {
	l net.Listener

	sync.RWMutex
	opts Options

	ss map[uint64]Session
}

func newServer() *Server {
	return &Server{

	}
}

func (s *Server) Options() Options {
	s.RLock()
	opts := s.opts
	s.RUnlock()
	return opts
}

func (s *Server) Start() error {
	opt := s.Options()

	l, err := net.ListenTCP("tcp", opt.Addr)
	if err != nil {
		return err
	}

	log.Logf("tcp: listening on %s", l.Addr())

	// swap address
	s.Lock()
	addr := s.opts.Addr
	s.opts.Addr = l.Addr()
	s.Unlock()

	go s.Accept()

	//TODO

}

func (s *Server) Accept() error {
	var tempDelay time.Duration

	for {
		c, err := s.l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Logf("tcp: Accept error: %v; retrying in %v\n", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		ses := newSession(c)
		go ses.readLoop()
	}
}

func (h *Server) Push(ctx context.Context, req *access.PushReq, rsp *access.PushRsp) error {

}