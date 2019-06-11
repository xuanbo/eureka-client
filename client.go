package eureka_client

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Client eureka客户端
type Client struct {
	// for monitor system signal
	signalChan chan os.Signal
	mutex      sync.RWMutex
	Running    bool
	Config     *Config
	// eureka服务中注册的应用
	Applications *Applications
}

// Start 启动时注册客户端，并后台刷新服务列表，以及心跳
func (c *Client) Start() {
	c.mutex.Lock()
	c.Running = true
	c.mutex.Unlock()
	// 注册
	if err := c.doRegister(); err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("Register application instance successful")
	// 刷新服务列表
	go c.refresh()
	// 心跳
	go c.heartbeat()
	// 监听退出信号，自动删除注册信息
	go c.handleSignal()
}

// refresh 刷新服务列表
func (c *Client) refresh() {
	for {
		if c.Running {
			if err := c.doRefresh(); err != nil {
				fmt.Println(err)
			} else {
				log.Println("Refresh application instance successful")
			}
		} else {
			break
		}
		time.Sleep(c.Config.RegistryFetchIntervalSeconds)
	}
}

// heartbeat 心跳
func (c *Client) heartbeat() {
	for {
		time.Sleep(c.Config.RenewalIntervalInSecs)
		if c.Running {
			if err := c.doHeartbeat(); err != nil {
				fmt.Println(err)
			} else {
				log.Println("Heartbeat application instance successful")
			}
		} else {
			break
		}
	}
}

func (c *Client) doRegister() error {
	instance := c.Config.Instance
	return Register(c.Config.DefaultZone, c.Config.App, instance)
}

func (c *Client) doUnRegister() error {
	instance := c.Config.Instance
	return UnRegister(c.Config.DefaultZone, instance.App, instance.InstanceID)
}

func (c *Client) doHeartbeat() error {
	instance := c.Config.Instance
	return Heartbeat(c.Config.DefaultZone, instance.App, instance.InstanceID)
}

func (c *Client) doRefresh() error {
	// todo If the delta is disabled or if it is the first time, get all applications

	// get all applications
	applications, err := Refresh(c.Config.DefaultZone)
	if err != nil {
		return err
	}

	// set applications
	c.mutex.Lock()
	c.Applications = applications
	c.mutex.Unlock()
	return nil
}

// handleSignal 监听退出信号，删除注册的实例
func (c *Client) handleSignal() {
	if c.signalChan == nil {
		c.signalChan = make(chan os.Signal)
	}
	signal.Notify(c.signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		switch <-c.signalChan {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			log.Println("Receive exit signal, client instance going to de-egister")
			err := c.doUnRegister()
			if err != nil {
				log.Println(err.Error())
			} else {
				log.Println("UnRegister application instance successful")
			}
			os.Exit(0)
		}
	}
}

// NewClient 创建客户端
func NewClient(config *Config) *Client {
	defaultConfig(config)
	config.Instance = NewInstance(config.App, getLocalIP(), config.Port)
	return &Client{Config: config}
}

func defaultConfig(config *Config) {
	if config.DefaultZone == "" {
		config.DefaultZone = "http://localhost:8761/eureka/"
	}
	if config.RenewalIntervalInSecs == 0 {
		config.RenewalIntervalInSecs = 30 * time.Second
	}
	if config.RegistryFetchIntervalSeconds == 0 {
		config.RegistryFetchIntervalSeconds = 90 * time.Second
	}
	if config.App == "" {
		config.App = "SERVER"
	}
	if config.Port == 0 {
		config.Port = 80
	}
}
