package eureka_client

import "time"

// eureka client configuration
type EurekaClientConfig struct {
	// eureka server address
	DefaultZone              string
	HeartbeatIntervalSeconds time.Duration
	RefreshIntervalSeconds   time.Duration

	App        string
	Port       int
	SecurePort int

	instanceInfo *InstanceInfo
}

type InstanceInfo struct {
	Instance *Instance `json:"instance"`
}

// eureka client instance configuration
type Instance struct {
	InstanceId       string          `json:"instanceId"`
	HostName         string          `json:"hostName"`
	IpAddr           string          `json:"ipAddr"`
	App              string          `json:"app"`
	Port             *PortWrapper    `json:"port"`
	SecurePort       *PortWrapper    `json:"securePort"`
	Status           string          `json:"status"`
	OverriddenStatus string          `json:"overriddenStatus"`
	DataCenterInfo   *DataCenterInfo `json:"dataCenterInfo"`
}

type PortWrapper struct {
	Enabled string `json:"@enabled"`
	Port    int    `json:"$"`
}

type DataCenterInfo struct {
	Name  string `json:"name"`
	Class string `json:"@class"`
}
