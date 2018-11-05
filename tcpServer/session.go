package tcpServer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/dayan-be/access-service/tcpServer/socket"
	"github.com/gogo/protobuf/proto"
	"github.com/dayan-be/access-service/proto"
	"golang.org/x/net/context"
	"io"
	"net"
	"sync"
)


const (
	MSG_READ_SIZE   = 4096
	MSG_BUFFER_SIZE = 10240
)

type Session struct{
	socket socket.Socket
	uid  uint64
	authed bool
	colseNotifyCh chan struct{}
	status int32
	srv *TcpServer
}

func NewSession(srv *TcpServer,id uint64, con net.Conn)*Session{
	ss := &Session{
		socket:socket.NewSocket(con),
		uid:0,
		authed:false,
		srv: srv,
	}

	ss.socket.SetFid(id)
	return ss
}

func (ss *Session)Id()uint64{
	return ss.socket.GetFid()
}



func (ss *Session)Close()error{
	ss.srv.sessionHub.Delete(ss.Id())
	return ss.socket.Close()
}

func (ss *Session)StartReadAndHandle(){
	ctx := context.Background()
	msgbuf := bytes.NewBuffer(make([]byte, 0, MSG_BUFFER_SIZE))
	// 数据缓冲
	databuf := make([]byte, MSG_READ_SIZE)
	// 消息长度
	length := 0
	// 消息长度uint32
	ulength := uint32(0)
	msgFlag := ""

	for {
		// 读取数据
		n, err := ss.socket.Read(databuf)
		if err == io.EOF {
			fmt.Printf("Client exit: %s\n", ss.socket.RemoteAddr())
		}
		if err != nil {
			fmt.Printf("Read error: %s\n", err)
			return
		}
		fmt.Println(databuf[:n])
		// 数据添加到消息缓冲
		n, err = msgbuf.Write(databuf[:n])
		if err != nil {
			fmt.Printf("Buffer write error: %s\n", err)
			return
		}

		// 消息分割循环
		for {
			// 消息头
			if length == 0 && msgbuf.Len() >= 6 {
				msgFlag = string(msgbuf.Next(2))
				if msgFlag != "DY" {
					fmt.Printf("invalid message")
					ss.srv.sessionHub.Delete(ss.socket.GetFid())
					return
				}
				binary.Read(msgbuf, binary.LittleEndian, &ulength)
				length = int(ulength)
				// 检查超长消息
				if length > MSG_BUFFER_SIZE {
					fmt.Printf("Message too length: %d\n", length)
					ss.srv.sessionHub.Delete(ss.socket.GetFid())
					return
				}
			}
			// 消息体
			if length > 0 && msgbuf.Len() >= length {
				length = 0
				go ss.HandleMsg(ctx,msgbuf.Next(length))
			} else {
				break
			}
		}
	}

}


func (ss *Session)HandleMsg(ctx context.Context, msg []byte){
		reqPkg := new(access.PkgReq)
		err := proto.Unmarshal(msg, reqPkg)
		if err != nil{
			return
		}

		//todo:调用后端服务
		//1.认证socket

		//2.

}

func (ss *Session)WriteMsg(msg []byte)error{
	length := len(msg)
	buf := bytes.NewBuffer(make([]byte,0,4))
	err := binary.Write(buf,binary.LittleEndian,length)
	if err != nil {
		return err
	}
	var totalMsg []byte
	totalMsg = append(totalMsg, []byte("DY")...)
	totalMsg = append(totalMsg, buf.Bytes()...)
	totalMsg = append(totalMsg, msg...)
	return nil
}


type SessionHub struct{
	sessions sync.Map
}


func NewSessionHub() *SessionHub {
	return &SessionHub{}
}

//添加一个socket
func (sh *SessionHub) Add(ss *Session) {
	_session, loaded := sh.sessions.LoadOrStore(ss.Id(), ss)
	if !loaded {
		return
	}
	sh.sessions.Store(ss.Id(), ss)
	if oldSession := _session.(*Session); ss != oldSession{
		oldSession.Close()
	}

}

func (sh *SessionHub) Delete(id uint64) {
	_ss, loaded := sh.sessions.Load(id)
	if !loaded {
		return
	} else {
		_ss.(Session).Close()
		sh.sessions.Delete(id)
	}
}

func (sh *SessionHub) Get(id uint64) (*Session, bool) {
	_ss, loaded := sh.sessions.Load(id)
	if !loaded {
		return nil, false
	}
	return _ss.(*Session), true
}

func (sh *SessionHub) Range(f func(ss *Session) bool) {
	sh.sessions.Range(func(key, value interface{}) bool {
		return f(value.(*Session))
	})
}