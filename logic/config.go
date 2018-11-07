package logic

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"sync"
)

type Configer struct {
	Srv struct {
		SrvName string `yaml:"srvName"`
		SrvId   uint32 `yaml:"srvId"`
		Port    int  `yaml:"port"`
	}

	Log struct {
		LogLevel    int  `yaml:"logLevel"`
		LogFileSize int  `yaml:"logFileSize"`
		JsonFile    bool `yaml:"jsonFile"`
	}


}

var cfgIns *Configer
var once sync.Once

func Config() *Configer {

	once.Do(func() {
		cfgIns = &Configer{}
	})
	return cfgIns
}

func (c *Configer) Load() {

	buff, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		goto FAILED
	}

	err = yaml.Unmarshal(buff, c)
	if err != nil {
		goto FAILED
	}
	return

FAILED:
	fmt.Printf("failed:%v",err)
	os.Exit(1)
}
