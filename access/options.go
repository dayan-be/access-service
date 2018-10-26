package access


import (
	"context"
	"github.com/dayan-be/access-service/access/net"
)

type Options struct {
	Transport    anet.Transport
	Address      string

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func newOptions(opt ...Option) Options {
	opts := Options{}
	for _, o := range opt {
		o(&opts)
	}

	if opts.Transport == nil {
		opts.Transport = anet.DefaultTransport
	}

	if len(opts.Address) == 0 {
		opts.Address = DefaultAddress
	}

	return opts
}

// Address to bind to - host:port
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// Transport mechanism for communication e.g http, rabbitmq, etc
func Transport(t anet.Transport) Option {
	return func(o *Options) {
		o.Transport = t
	}
}


