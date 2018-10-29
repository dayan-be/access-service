package access


import "github.com/dayan-be/access-service/access/net"

type conn struct {
	sock anet.Socket

	auth bool

	uid uint64
}


type Handler struct {
	connMap map[uint64]*conn
}

func newHandler() *Handler {
	return &Handler{
		connMap:map[uint64]*conn{},
	}
}

func (h *Handler) Push(ctx context.Context, req *access.PushReq, rsp *access.PushRsp) error {

}

func (h *Handler) handle(sock anet.Socket, msg *anet.Message) error {

	return nil
}