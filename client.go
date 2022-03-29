package eureka_client

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Client eureka客户端
type Client struct {
	logger Logger

	// for monitor system signal
	signalChan chan os.Signal
	mutex      sync.RWMutex
	running    bool

	Config   *Config
	Instance *Instance

	// eureka服务中注册的应用
	Applications *Applications
}

// Option 自定义
type Option func(instance *Instance)

// SetLogger 设置日志实现
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
}

// Start 启动时注册客户端，并后台刷新服务列表，以及心跳
func (c *Client) Start() {
	c.mutex.Lock()
	c.running = true
	c.mutex.Unlock()
	// 刷新服务列表
	go c.refresh()
	// 心跳
	go c.heartbeat()
	// 监听退出信号，自动删除注册信息
	go c.handleSignal()
}

// refresh 刷新服务列表
func (c *Client) refresh() {
	timer := time.NewTimer(0)
	interval := time.Duration(c.Config.RenewalIntervalInSecs) * time.Second
	for c.running {
		select {
		case <-timer.C:
			if err := c.doRefresh(); err != nil {
				c.logger.Error("refresh application instance failed", err)
			} else {
				c.logger.Debug("refresh application instance successful")
			}
		}
		// reset interval
		timer.Reset(interval)
	}
	// stop
	timer.Stop()
}

// heartbeat 心跳
func (c *Client) heartbeat() {
	timer := time.NewTimer(0)
	interval := time.Duration(c.Config.RegistryFetchIntervalSeconds) * time.Second
	for c.running {
		select {
		case <-timer.C:
			err := c.doHeartbeat()
			if err == nil {
				c.logger.Debug("heartbeat application instance successful")
			} else if err == ErrNotFound {
				// heartbeat not found, need register
				err = c.doRegister()
				if err == nil {
					c.logger.Info("register application instance successful")
				} else {
					c.logger.Error("register application instance failed", err)
				}
			} else {
				c.logger.Error("heartbeat application instance failed", err)
			}
		}
		// reset interval
		timer.Reset(interval)
	}
	// stop
	timer.Stop()
}

func (c *Client) doRegister() error {
	return Register(c.Config.DefaultZone, c.Config.App, c.Instance)
}

func (c *Client) doUnRegister() error {
	return UnRegister(c.Config.DefaultZone, c.Instance.App, c.Instance.InstanceID)
}

func (c *Client) doHeartbeat() error {
	return Heartbeat(c.Config.DefaultZone, c.Instance.App, c.Instance.InstanceID)
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
			c.logger.Info("receive exit signal, client instance going to de-register")
			err := c.doUnRegister()
			if err != nil {
				c.logger.Error("de-register application instance failed", err)
			} else {
				c.logger.Info("de-register application instance successful")
			}
			os.Exit(0)
		}
	}
}

// NewClient 创建客户端
func NewClient(config *Config, opts ...Option) *Client {
	defaultConfig(config)
	instance := NewInstance(config)
	client := &Client{
		logger:   NewLogger(),
		Config:   config,
		Instance: instance,
	}
	for _, opt := range opts {
		opt(client.Instance)
	}
	return client
}

func defaultConfig(config *Config) {
	if config.DefaultZone == "" {
		config.DefaultZone = "http://localhost:8761/eureka/"
	}
	if config.RenewalIntervalInSecs == 0 {
		config.RenewalIntervalInSecs = 30
	}
	if config.RegistryFetchIntervalSeconds == 0 {
		config.RegistryFetchIntervalSeconds = 15
	}
	if config.DurationInSecs == 0 {
		config.DurationInSecs = 90
	}
	if config.App == "" {
		config.App = "unknown"
	} else {
		config.App = strings.ToLower(config.App)
	}
	if config.IP == "" {
		config.IP = GetLocalIP()
	}
	if config.HostName == "" {
		config.HostName = config.IP
	}
	if config.Port == 0 {
		config.Port = 80
	}
	if config.InstanceID == "" {
		config.InstanceID = fmt.Sprintf("%s:%s:%d", config.IP, config.App, config.Port)
	}
}

// 根据服务名获取注册的服务实例列表
func (c *Client) GetApplicationInstance(name string) []Instance {
	instances := make([]Instance, 0)
	c.mutex.Lock()
	if c.Applications != nil {
		for _, app := range c.Applications.Applications {
			if app.Name == name {
				instances = append(instances, app.Instances...)
			}
		}
	}
	c.mutex.Unlock()

	return instances
}
