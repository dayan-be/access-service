package access

import (

	"github.com/micro/go-log"

	"github.com/dayan-be/access-service/access/net"
	"sync"
	"runtime/debug"
)

type tcpServer struct {
	exit chan chan error

	sync.RWMutex
	opts        Options
	// used for first registration
	registered bool
	// graceful exit
	wg sync.WaitGroup
}

func newTcpServer(opts ...Option) Server {
	options := newOptions(opts...)
	return &tcpServer{
		opts: options,
		exit:        make(chan chan error),
	}
}

func (s *tcpServer) accept(sock anet.Socket) {
	defer func() {
		// close socket
		sock.Close()

		if r := recover(); r != nil {
			log.Log("panic recovered: ", r)
			log.Log(string(debug.Stack()))
		}
	}()

	for {
		var msg anet.Message
		if err := sock.Recv(&msg); err != nil {
			return
		}

		// add to wait group
		s.wg.Add(1)
		//
		//// we use this Timeout header to set a server deadline
		//to := msg.Header["Timeout"]
		//// we use this Content-Type header to identify the codec needed
		//ct := msg.Header["Content-Type"]
		//
		//cf, err := s.newCodec(ct)
		//// TODO: needs better error handling
		//if err != nil {
		//	sock.Send(&anet.Message{
		//		Header: map[string]string{
		//			"Content-Type": "text/plain",
		//		},
		//		Body: []byte(err.Error()),
		//	})
		//	s.wg.Done()
		//	return
		//}
		//
		//codec := newTcpPlusCodec(&msg, sock, cf)
		//
		//// strip our headers
		//hdr := make(map[string]string)
		//for k, v := range msg.Header {
		//	hdr[k] = v
		//}
		//delete(hdr, "Content-Type")
		//delete(hdr, "Timeout")
		//
		//ctx := metadata.NewContext(context.Background(), hdr)
		//
		//// set the timeout if we have it
		//if len(to) > 0 {
		//	if n, err := strconv.ParseUint(to, 10, 64); err == nil {
		//		ctx, _ = context.WithTimeout(ctx, time.Duration(n))
		//	}
		//}
		//
		//// TODO: needs better error handling
		//if err := s.rpc.serveRequest(ctx, codec, ct); err != nil {
		//	s.wg.Done()
		//	log.Logf("Unexpected error serving request, closing socket: %v", err)
		//	return
		//}
		s.wg.Done()
	}
}

func (s *tcpServer) Options() Options {
	s.RLock()
	opts := s.opts
	s.RUnlock()
	return opts
}

func (s *tcpServer) Init(opts ...Option) error {
	s.Lock()
	for _, opt := range opts {
		opt(&s.opts)
	}
	s.Unlock()
	return nil
}

func (s *tcpServer) Start() error {
	config := s.Options()

	ts, err := config.Transport.Listen(config.Address)
	if err != nil {
		return err
	}

	log.Logf("Listening on %s", ts.Addr())
	s.Lock()
	// swap address
	addr := s.opts.Address
	s.opts.Address = ts.Addr()
	s.Unlock()

	go ts.Accept(s.accept)

	go func() {
		// wait for exit
		ch := <-s.exit

		s.wg.Wait()
		// close anet listener
		ch <- ts.Close()

		s.Lock()
		// swap back address
		s.opts.Address = addr
		s.Unlock()
	}()

	return nil
}

func (s *tcpServer) Stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}
