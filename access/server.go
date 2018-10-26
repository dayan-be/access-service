package access


import (
	"os"
	"os/signal"
	"syscall"

	"github.com/micro/go-log"
)

type Server interface {
	Options() Options
	Init(...Option) error
	Start() error
	Stop() error
}


type Option func(*Options)

var (
	DefaultAddress        = ":0"
	DefaultServer  Server = newRpcServer()
)

// DefaultOptions returns config options for the default service
func DefaultOptions() Options {
	return DefaultServer.Options()
}

// Init initialises the default server with options passed in
func Init(opt ...Option) {
	if DefaultServer == nil {
		DefaultServer = newTcpServer(opt...)
	}
	DefaultServer.Init(opt...)
}

// NewServer returns a new server with options passed in
func NewServer(opt ...Option) Server {
	return newTcpServer(opt...)
}

// Run starts the default server and waits for a kill
// signal before exiting. Also registers/deregisters the server
func Run() error {
	if err := Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	log.Logf("Received signal %s", <-ch)

	return Stop()
}

// Start starts the default server
func Start() error {
	config := DefaultServer.Options()
	log.Logf("Starting server %s id %s", config.Name, config.Id)
	return DefaultServer.Start()
}

// Stop stops the default server
func Stop() error {
	log.Logf("Stopping server")
	return DefaultServer.Stop()
}

