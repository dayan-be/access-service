package main

import (
	"flag"
	"fmt"
	_ "github.com/dayan-be/golibs/log"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)


var (
	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"
)

func main() {

	//显示版本号信息　
	version := flag.Bool("v", false, "version")
	flag.Parse()

	if *version {
		fmt.Println("Git Tag: " + GitTag)
		fmt.Println("Build Time: " + BuildTime)
		return
	}

	//2.log
	logrus.SetLevel(logrus.DebugLevel)

	//1.load configer
	cfg := Config()
	cfg.Load()

	h := NewHandle()
	h.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	<-c
	os.Exit(0)
}
