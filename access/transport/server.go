package transport

type Server interface {
	Options() Options
	Init(...Option) error
}
