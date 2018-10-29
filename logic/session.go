package logic

import (
	"net"
	"runtime/debug"
	"io"
	"errors"
	"github.com/go-log/log"
	"encoding/binary"
	"github.com/dayan-be/access-service/proto"
	"github.com/gogo/protobuf/proto"
	"bytes"
)

const (
	MAX_PACKAGE_SIZE = 64*1024*1024
)

type session struct {
	net.Conn
	uid uint64
	auth bool
	s *server

	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func newSession(conn net.Conn) Session {
	return &session{
		Conn:conn,
		readBuf:bytes.NewBuffer(make([]byte, 0, MAX_PACKAGE_SIZE)),
		writeBuf:bytes.NewBuffer(make([]byte, 0, MAX_PACKAGE_SIZE)),
	}
}

type Session interface {
	net.Conn
	ReadLoop(conn net.Conn) error
}

type Message struct {
	Flag [2]byte // 业务标识
	Len uint32 // 业务包长度
	Body []byte // 业务包
}

func (s *session) readAtLeast(buf []byte, size int) (err error) {
	if len(data) < size {
		return errors.New("buf is small")
	}

	n := 0
	for n < size && err == nil {
		var nn int
		nn, err = s.Read(buf[n:size])
		if err != nil {
			return err
		}
		n += nn
	}
	if n == size {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (s *session) ReadLoop(conn net.Conn) error {
	defer func() {
		// close socket
		conn.Close()

		if r := recover(); r != nil {
			log.Log("panic recovered: ", r)
			log.Log(string(debug.Stack()))
		}
	}()

	for {


		flag := make([]byte, 2, 2)

		if err := s.readAtLeast(flag, 2); err != nil {
			// TODO: tell peer
			return err
		}

		if flag[0] != 'D' && flag[1] != 'Y' {
			// TODO: invalid business flag
			return errors.New("invalid flag")
		}

		lenByte := make([]byte, 4, 4)
		if err := s.readAtLeast(lenByte, 4); err != nil {
			// TODO: tell peer
			return err
		}
		bodyLen := binary.BigEndian.Uint32(lenByte)

		if bodyLen > MAX_PACKAGE_SIZE {
			return errors.New("package size exceeds max limit")
		}

		body := make([]byte, bodyLen, bodyLen)
		if err := s.readAtLeast(body, bodyLen); err != nil {
			// TODO
			return err
		}

		go s.handleRequest(body)
	}

	return nil
}

func (s *session) Send(msg *access.PkgRsp) error {
	buf := make([]byte, 4, )
}

func (s *session) handleRequest(body []byte) error {
	req := &access.PkgReq{}
	err := proto.Unmarshal(body, req)
	if err != nil {
		log.Log("unmarshal PkgReq failed")
		return err
	}

	if s.auth {
		// TODO: transfer to backend service
	} else {
		// ignore unauthenticated request
		if req.Body.Head.Token == "" && (req.Body.Head.Uid == 0 || req.Body.Head.Password == "") {
			log.Log("session is unauthenticated")
			return errors.New("session is unauthenticated")
		}

		// TODO: login
		s.login(req)

		s.auth = true
	}

	return nil
}

func (s *session) login(msg *access.PkgReq) error {
	return nil
}