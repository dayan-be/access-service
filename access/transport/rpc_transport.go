package transport

import (
	"io"
	"net"
	mnet "github.com/micro/util/go/lib/net"
	"time"
	"bufio"
	//"net/http"
	"github.com/micro/go-log"
	"sync"
	"io/ioutil"
	"bytes"
	"errors"
	"net/http"
)

type buffer struct {
	io.ReadWriter
}



type transportSocket struct {
	t   *transport
	//r    chan *http.Request
	conn net.Conn
	once sync.Once

	sync.Mutex
	buff *bufio.Reader
}

func (h *transportSocket) Recv(m *Message) error {


	return nil
}

func (h *transportSocket) Send(m *Message) error {
	b := bytes.NewBuffer(m.Body)
	defer b.Reset()

	return nil
}

func (h *transportSocket) error(m *Message) error {
	return nil
}

func (h *transportSocket) Close() error {
	err := h.conn.Close()
	h.once.Do(func() {
		h.Lock()
		h.buff.Reset(nil)
		h.buff = nil
		h.Unlock()
	})
	return err
}

type transportListener struct {
	t *transport
	listener net.Listener
}


func (h *transportListener) Addr() string {
	return h.listener.Addr().String()
}

func (h *transportListener) Close() error {
	return h.listener.Close()
}

func (h *transportListener) Accept(fn func(Socket)) error {
	var tempDelay time.Duration

	for {
		c, err := h.listener.Accept()
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
				log.Logf("http: Accept error: %v; retrying in %v\n", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		sock := &transportSocket{
			t:   h.t,
			conn: c,
			buff: bufio.NewReader(c),
			//r:    make(chan *http.Request, 1),
		}

		go func() {
			// TODO: think of a better error response strategy
			defer func() {
				if r := recover(); r != nil {
					log.Log("panic recovered: ", r)
					sock.Close()
				}
			}()

			fn(sock)
		}()
	}
}

type transport struct {
	opts Options
}

func (h *transport) Listen(addr string, opts ...ListenOptions) (Listener, error) {
	var l net.Listener
	var err error

	fn := func(addr string) (net.Listener, error) {
		return net.Listen("tcp", addr)
	}

	l, err = mnet.Listen(addr, fn)
	if err != nil {
		return nil, err
	}

	return &transportListener{
		t:       h,
		listener: l,
	}, nil
}

