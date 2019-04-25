package eureka_client

import "time"

// EurekaClientConfig eureka client configuration
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

// InstanceInfo the app instance info
type InstanceInfo struct {
	Instance *Instance `json:"instance"`
}

// Instance eureka client instance configuration
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

// PortWrapper wrap port
type PortWrapper struct {
	Enabled string `json:"@enabled"`
	Port    int    `json:"$"`
}

// DataCenterInfo the date center info
type DataCenterInfo struct {
	Name  string `json:"name"`
	Class string `json:"@class"`
}
