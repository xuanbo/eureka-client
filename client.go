package eureka_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// eureka client
type Client struct {
	mutex              sync.RWMutex
	Running            bool
	EurekaClientConfig *EurekaClientConfig
	// services from eureka server
	Services []Instance
}

func (c *Client) Start() {
	c.mutex.Lock()
	c.Running = true
	c.mutex.Unlock()

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

	body, err := json.Marshal(c.EurekaClientConfig.instanceInfo)
	if err != nil {
		return err
	}

	url := c.EurekaClientConfig.DefaultZone + "apps/" + c.EurekaClientConfig.App
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("registing failed %s\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("registing success")
	} else {
		return errors.New(fmt.Sprintf("registing failed, status: %d\n", resp.StatusCode))

	}

	return nil
}

func (c *Client) doUnRegister() error {
	instance := c.EurekaClientConfig.instanceInfo.Instance
	url := fmt.Sprintf("%sapps/%s/%s",
		c.EurekaClientConfig.DefaultZone, instance.App, instance.InstanceId)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("unregisting failed %s\n", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("unregisting success")
	} else {
		return errors.New(fmt.Sprintf("unregisting failed, status: %d\n", resp.StatusCode))
	}

	return nil
}

func (c *Client) doHeartbeat() error {
	instance := c.EurekaClientConfig.instanceInfo.Instance
	url := fmt.Sprintf("%sapps/%s/%s?status=UP&lastDirtyTimestamp=%d",
		c.EurekaClientConfig.DefaultZone, instance.App, instance.InstanceId, time.Now().Nanosecond())

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("heartbeat failed %s\n", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("heartbeat success")
	} else {
		return errors.New(fmt.Sprintf("heartbeat failed, status: %d\n", resp.StatusCode))
	}

	return nil
}

func (c *Client) doRefresh() error {
	// todo If the delta is disabled or if it is the first time, get all applications

	// get all applications
	url := c.EurekaClientConfig.DefaultZone + "apps"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("refresh failed %s\n", err)
	}
	req.Header.Add("Accept", " application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("refresh failed %s\n", err)
		}
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return err
	}

	instances, err := ParseApplications(result)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("refresh success")

		// set applications
		c.mutex.Lock()
		c.Services = instances
		c.mutex.Unlock()
	} else {
		return errors.New(fmt.Sprintf("refresh failed, status: %d\n", resp.StatusCode))
	}

	return nil
}

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
