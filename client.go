package eureka_client

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/xuanbo/requests"
)

// Client eureka client
type Client struct {
	mutex              sync.RWMutex
	Running            bool
	EurekaClientConfig *EurekaClientConfig
	// services from eureka server
	Services []Instance
}

// Start start eureka client
// do refresh and heartbeat
func (c *Client) Start() {
	c.mutex.Lock()
	c.Running = true
	c.mutex.Unlock()

	// refresh、heartbeat
	refreshTicker := time.NewTicker(c.EurekaClientConfig.RefreshIntervalSeconds)
	heartbeatTicker := time.NewTicker(c.EurekaClientConfig.HeartbeatIntervalSeconds)

	go func() {
		for range refreshTicker.C {
			if c.Running {
				if err := c.doRefresh(); err != nil {
					fmt.Println(err)
				}
			} else {
				break
			}
		}
	}()

	go func() {
		if err := c.doRegister(); err != nil {
			fmt.Println(err)
		}
		for range heartbeatTicker.C {
			if c.Running {
				if err := c.doHeartbeat(); err != nil {
					fmt.Println(err)
				}
			} else {
				break
			}
		}
	}()
}

// Shutdown close eureka client
// delete info from eureka server
func (c *Client) Shutdown() {
	c.mutex.Lock()
	c.Running = false
	c.mutex.Unlock()

	if err := c.doUnRegister(); err != nil {
		fmt.Println(err)
	}
}

func (c *Client) doRegister() error {
	c.mutex.Lock()
	c.EurekaClientConfig.instanceInfo.Instance.Status = "UP"
	c.mutex.Unlock()

	u := c.EurekaClientConfig.DefaultZone + "apps/" + c.EurekaClientConfig.App
	info := c.EurekaClientConfig.instanceInfo

	// status: http.StatusNoContent
	result := requests.Post(u).Json(info).Send().Status2xx()
	if result.Err != nil {
		return fmt.Errorf("registing failed, error: %s", result.Err)
	} else {
		fmt.Println("registing success")
	}

	return nil
}

func (c *Client) doUnRegister() error {
	instance := c.EurekaClientConfig.instanceInfo.Instance
	u := fmt.Sprintf("%sapps/%s/%s",
		c.EurekaClientConfig.DefaultZone, instance.App, instance.InstanceId)

	result := requests.Delete(u).Send().StatusOk()
	if result.Err != nil {
		return fmt.Errorf("unregisting failed, error: %s", result.Err)
	} else {
		fmt.Println("unregisting success")
	}

	return nil
}

func (c *Client) doHeartbeat() error {
	instance := c.EurekaClientConfig.instanceInfo.Instance
	u := fmt.Sprintf("%sapps/%s/%s", c.EurekaClientConfig.DefaultZone, instance.App, instance.InstanceId)
	params := url.Values{
		"status":             {"UP"},
		"lastDirtyTimestamp": {strconv.Itoa(time.Now().Nanosecond())},
	}

	result := requests.Put(u).Params(params).Send().StatusOk()
	if result.Err != nil {
		return fmt.Errorf("heartbeat failed, error: %s", result.Err)
	} else {
		fmt.Println("heartbeat success")
	}

	return nil
}

func (c *Client) doRefresh() error {
	// todo If the delta is disabled or if it is the first time, get all applications

	// get all applications
	u := c.EurekaClientConfig.DefaultZone + "apps"

	r := requests.Get(u).Header("Accept", " application/json").Send().StatusOk()
	if r.Err != nil {
		return fmt.Errorf("refresh failed, error: %s", r.Err)
	} else {
		fmt.Println("refresh success")

		// parse applications
		var result map[string]interface{}
		err := r.Json(&result)
		if err != nil {
			return err
		}
		instances, err := ParseApplications(result)
		if err != nil {
			return err
		}

		// set applications
		c.mutex.Lock()
		c.Services = instances
		c.mutex.Unlock()
	}

	return nil
}

// NewClient returns a new eureka client
func NewClient(config *EurekaClientConfig) *Client {
	setDefault(config)

	ipAddrs, err := GetIpAddrs()
	if err != nil {
		panic(err)
	}
	ipAddr := ipAddrs[0]

	config.instanceInfo = &InstanceInfo{
		Instance: &Instance{
			InstanceId:       fmt.Sprintf("%s:%s:%d", ipAddr, config.App, config.Port),
			HostName:         ipAddr,
			IpAddr:           ipAddr,
			App:              config.App,
			Port:             &PortWrapper{Enabled: "true", Port: config.Port},
			SecurePort:       &PortWrapper{Enabled: "true", Port: config.SecurePort},
			Status:           "DOWN",
			OverriddenStatus: "UNKNOWN",
			DataCenterInfo: &DataCenterInfo{
				Name:  "MyOwn",
				Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
			},
		},
	}

	c := &Client{EurekaClientConfig: config}
	return c
}

func setDefault(config *EurekaClientConfig) {
	if config.DefaultZone == "" {
		config.DefaultZone = "http://localhost:8761/eureka/"
	}
	if config.HeartbeatIntervalSeconds == 0 {
		config.HeartbeatIntervalSeconds = 30 * time.Second
	}
	if config.RefreshIntervalSeconds == 0 {
		config.RefreshIntervalSeconds = 30 * time.Second
	}
	if config.App == "" {
		config.App = "SERVER"
	}
	if config.Port == 0 {
		config.Port = 80
	}
	if config.SecurePort == 0 {
		config.SecurePort = 443
	}
}
