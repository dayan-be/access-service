package logic

type Options struct {
	Addr string
}

type Option func(*Options)

func newOptions(opt ...Option) Options {
	opts := Options{}
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

func Addr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

