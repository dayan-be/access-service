package tcp_server

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	writeTimeout    int
	readTimeout     int
	deadlineTimeout int
	srvId           uint32 //当前服务实例id
	msgProFunc      func(context.Context, []byte)
	port            int
}

type Option func(*Options)

func SetDeadlineTimeOut(t int) Option {
	return func(o *Options) {
		o.deadlineTimeout = t
	}
}

func SetReadTimeout(t int) Option {
	return func(o *Options) {
		o.readTimeout = t
	}
}

func SetWriteTimeout(t int) Option {
	return func(o *Options) {
		o.writeTimeout = t
	}
}

func SetSrvId(id uint32) Option {
	return func(o *Options) {
		o.srvId = id
	}
}

func SetMsgProcFunc(f func(context.Context, []byte)) Option {
	return func(o *Options) {
		o.msgProFunc = f
	}
}

func SetListenPort(p int)Option{
	return func(o *Options){
		 o.port = p
	}
}

type TcpServer struct {
	uid2Sid    sync.Map
	opt        Options
	listener   net.Listener
	wg         sync.WaitGroup
	sessionHub SessionHub
	scnt       uint32 //当前socket 计数
	baseValue  uint64 //会话id的基础值
	closeCh    chan struct{}
}

func NewTcpServer(op ...Option) *TcpServer {
	s := &TcpServer{
		scnt:     0,
	}

	for _, o := range op {
		o(&(s.opt))
	}
	s.baseValue = uint64(s.opt.srvId) << 32
	return s
}

// ErrListenClosed listener is closed error.
var ErrListenClosed = errors.New("listener is closed")

func (srv *TcpServer) Run() error {
	var err error
	port := strconv.FormatInt(int64(srv.opt.port),10)
	addr := "0.0.0.0:" + port
	srv.listener ,err = net.Listen("tcp",addr)
	if err != nil {
		log.Errorf("listen failed:%v",err)
	}
	var (
		tempDelay time.Duration // how long to sleep on accept failure
		closeCh   = srv.closeCh
	)
	for {
		conn, e := srv.listener.Accept()
		if e != nil {
			select {
			case <-closeCh:
				return ErrListenClosed
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0

		go func(con net.Conn, srv *TcpServer) {
			id := atomic.AddUint32(&srv.scnt, 1)
			tmp := srv.baseValue + uint64(id)
			var sess = NewSession(srv, tmp, conn)
			srv.sessionHub.Add(sess)
			sess.StartReadAndHandle()
		}(conn, srv)
	}

	return nil
}

func (srv *TcpServer) Stop() error {
	var (
		count int
	)
	srv.sessionHub.Range(func(sess *Session) bool {
		count++
		sess.Close()
		return true
	})
	return nil
}

//向指定uid发送消息
func (srv *TcpServer) SendMsgByUid(uid uint64, msg []byte) error {
	sid, loaded := srv.uid2Sid.Load(uid)
	if loaded {
		return srv.SendMsgBySid(sid.(uint64), msg)
	}
	return errors.New("not find uid :" + strconv.FormatUint(uid, 10))
}

//向指定会话id发送消息
func (srv *TcpServer) SendMsgBySid(sid uint64, msg []byte) error {
	_ss, loaded := srv.sessionHub.Get(sid)
	if loaded {
		return _ss.WriteMsg(msg)
	}
	return errors.New("not find fid" + strconv.FormatUint(sid, 10))
}
