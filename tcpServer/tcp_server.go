package tcpServer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/dayan-be/access-service/proto"
	"github.com/dayan-be/access-service/tcpServer/socket"
	"github.com/gogo/protobuf/proto"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	MSG_READ_SIZE   = 4096
	MSG_BUFFER_SIZE = 10240
)

type Options struct {
	writeTimeout    int
	readTimeout     int
	deadlineTimeout int
}

func SetDeadlineTime(t int)Option{
	return func(o *Options){
		o.deadlineTimeout = t
	}
}


type Option func(*Options)

type TcpServer struct {
	uid2Fid   sync.Map
	opt       Options
	listener  *net.TCPListener
	wg        sync.WaitGroup
	socketHub *socket.SocketHub
	scnt      uint32 //当前socket 计数
	Id        uint32 //当前服务实例id
}

func NewTcpServer(l *net.TCPListener, op ...Option) *TcpServer {
	s := &TcpServer{
		listener:  l,
		socketHub: socket.NewSocketHub(),
	}

	for _, o := range op {
		o(&(s.opt))
	}
	return s
}


func (srv *TcpServer) Init() {

}

func (srv *TcpServer) Run() {

	for {
		s, err := srv.listener.Accept()
		if err != nil {

		}
		base := uint64(srv.Id) << 32
		tmp := srv.scnt + 1
		ss := socket.NewSocket(s)
		ss.SetFid(base + uint64(tmp))

		go srv.MsgHandle(ss)
	}

}

func (s *TcpServer) Stop() {

}

//向指定uid发送消息
func (srv *TcpServer) SendMsgByUid(uid uint64, msg []byte) error {
	fid, loaded := srv.uid2Fid.Load(uid)
	if loaded {
		return srv.SendMsgByfid(fid.(uint64), msg)
	}
	return errors.New("not find uid :" + strconv.FormatUint(uid, 10))
}

//向指定链路id发送消息
func (srv *TcpServer) SendMsgByfid(fid uint64, msg []byte) error {
	_socket, loaded := srv.socketHub.Get(fid)
	if loaded {
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
		_, err = _socket.Write(totalMsg)
		return err
	}
	return errors.New("not find fid" + strconv.FormatUint(fid, 10))
}

//处理链接消息
func (srv *TcpServer) MsgHandle(socket socket.Socket) {
	defer srv.socketHub.Delete(socket.GetFid())
	//初始化socket
	srv.socketHub.Add(socket)
	socket.SetWriteDeadline(time.Now().Add(time.Duration(srv.opt.deadlineTimeout)))
	socket.SetWriteDeadline(time.Now().Add(time.Duration(srv.opt.writeTimeout)))
	socket.SetReadDeadline(time.Now().Add(time.Duration(srv.opt.readTimeout)))

	// 消息缓冲
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
		n, err := socket.Read(databuf)
		if err == io.EOF {
			fmt.Printf("Client exit: %s\n", socket.RemoteAddr())
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
					srv.socketHub.Delete(socket.GetFid())
					return
				}
				binary.Read(msgbuf, binary.LittleEndian, &ulength)
				length = int(ulength)
				// 检查超长消息
				if length > MSG_BUFFER_SIZE {
					fmt.Printf("Message too length: %d\n", length)
					srv.socketHub.Delete(socket.GetFid())
					return
				}
			}
			// 消息体
			if length > 0 && msgbuf.Len() >= length {
				go srv.ProcMsg(socket.GetFid(),msgbuf.Next(length))
				length = 0
			} else {
				break
			}
		}
	}
}



func (srv *TcpServer)ProcMsg(fid uint64, msg []byte){
		reqPkg := new(access.PkgReq)
		err := proto.Unmarshal(msg, reqPkg)
		if err != nil{
			return
		}

		//todo:调用后端服务

}
