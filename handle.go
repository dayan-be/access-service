package main

import (
	"errors"
	"github.com/dayan-be/access-service/proto"
	"github.com/dayan-be/access-service/server"
	"github.com/dayan-be/golibs/micro-codec/byterpc"
	"github.com/golang/protobuf/proto"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

type Handle struct {
	tcpSrv   *server.TcpServer
	microSrv micro.Service
}

func NewHandle() *Handle {
	h := &Handle{}

	h.tcpSrv = server.NewTcpServer(
		server.Addr(Config().Srv.Addr),
		server.SrvId(Config().Srv.SrvId),
		server.HandleRequest(h.HandleRequest),
	)
	h.microSrv = micro.NewService(
		micro.Name(Config().Srv.SrvName),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
		micro.Version(Config().Srv.Version),
		micro.Metadata(map[string]string{"ID": strconv.FormatUint(uint64(Config().Srv.SrvId), 10)}),
	)

	h.microSrv.Init()
	access.RegisterAccessHandler(h.microSrv.Server(), h)
	return h
}

func (h *Handle) Start() {
	go func() {
		if err := h.microSrv.Run(); err != nil {
			logrus.Fatalf("run micro service failed(err:%v)", err)
		}
	}()

	go func() {
		if err := h.tcpSrv.Run(); err != nil {
			logrus.Fatalf("run tcp server failed(err:%v)", err)
		}
	}()
}

func (h *Handle) Push(ctx context.Context, req *access.PushReq, rsp *access.PushRsp) error {
	rsp.Code = 0
	// TODO: check uid:session is valid first?

	msg := &access.PkgRsp{
		Head: &access.PkgRspHead{Seq: req.Seq},
		Body: &access.PkgRspBody{
			Head: &access.RspHead{
				Uid:  req.Uid,
				Code: 0,
			},
			Bodys: []*access.RspBody{
				&access.RspBody{
					Service: req.Service,
					Method:  req.Method,
					Content: req.Content,
				},
			},
		},
	}

	err := h.PushMsg(req.Uid, msg)
	if err != nil {
		rsp.Code = 1
		return err
	}
	return nil
}

func (h *Handle) HandleRequest(ctx context.Context, ses *server.Session, body []byte) error {
	req := &access.PkgReq{}
	err := proto.Unmarshal(body, req)
	if err != nil {
		logrus.Errorf("PkgReq Unmarshal failed(err:%v)", err)
		return err
	}

	// TODO: need close session?
	if req.Head == nil || req.Body == nil {
		logrus.Errorf("invalid request")
		return errors.New("invalid request")
	}

	rsp := &access.PkgRsp{
		Head: &access.PkgRspHead{Seq: req.Head.Seq},
		Body: &access.PkgRspBody{
			Head: &access.RspHead{},
		},
	}
	defer h.Response(ctx, ses, rsp)

	if ses.Authed {
		// session has authenticated

		for _, subReq := range req.Body.Bodys {
			out, err := h.RawCallMicroService(ctx, subReq.Service, subReq.Method, subReq.Content)
			rspBody := &access.RspBody{
				Service:              subReq.Service,
				Method:               subReq.Method,
				Content:              out,
				Code: 0,
			}
			if err != nil {
				rspBody.Code = 1
			}
			rsp.Body.Bodys = append(rsp.Body.Bodys, rspBody)
		}

		return nil
	} else {
		if req.Body.Head.Account != nil {
			// TODO: authenticate account

		} else {
			rsp.Body.Head.Code = 1 // failed
			return nil
		}
	}
	return nil
}

func (h *Handle) Response(ctx context.Context, ses *server.Session, msg *access.PkgRsp) error {
	byt, err := proto.Marshal(msg)
	if err != nil {
		logrus.Errorf("PkgRsp Marshal failed(err:%v)", err)
		return err
	}

	return ses.WriteMsg(byt)
}

func (h *Handle) PushMsg(uid uint64, msg *access.PkgRsp) error {
	byt, err := proto.Marshal(msg)
	if err != nil {
		logrus.Errorf("PkgRsp Marshal failed(err:%v)", err)
		return err
	}

	return h.tcpSrv.SendMsgByUid(uid, byt)
}

func (h *Handle) RawCallMicroService(ctx context.Context, service, method string, in []byte, opts ...client.CallOption) (out []byte, err error) {
	c := client.NewClient(client.Codec("byte-rpc", byterpc.NewCodec))
	req := c.NewRequest(service, method, in, client.WithContentType("byte-rpc"))
	out =[]byte{}

	err = c.Call(ctx, req, &out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
