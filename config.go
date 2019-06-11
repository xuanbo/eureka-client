package eureka_client

import (
	"fmt"
	"time"
)

// Config eureka客户端配置
type Config struct {
	// eureka服务端地址
	DefaultZone string
	// 心跳间隔
	RenewalIntervalInSecs time.Duration
	// 获取服务列表间隔
	RegistryFetchIntervalSeconds time.Duration
	// 应用名称
	App string
	// 端口
	Port int

	// 服务实例信息
	Instance *Instance
}

// Applications eureka服务端注册的apps
type Applications struct {
	VersionsDelta string        `xml:"versions__delta,omitempty" json:"versions__delta,omitempty"`
	AppsHashcode  string        `xml:"apps__hashcode,omitempty" json:"apps__hashcode,omitempty"`
	Applications  []Application `xml:"application,omitempty" json:"application,omitempty"`
}

// Application eureka服务端注册的app
type Application struct {
	Name      string     `xml:"name" json:"name"`
	Instances []Instance `xml:"instance" json:"instance"`
}

// Instance 服务实例
type Instance struct {
	HostName                      string          `xml:"hostName" json:"hostName"`
	HomePageURL                   string          `xml:"homePageUrl,omitempty" json:"homePageUrl,omitempty"`
	StatusPageURL                 string          `xml:"statusPageUrl" json:"statusPageUrl"`
	HealthCheckURL                string          `xml:"healthCheckUrl,omitempty" json:"healthCheckUrl,omitempty"`
	App                           string          `xml:"app" json:"app"`
	IPAddr                        string          `xml:"ipAddr" json:"ipAddr"`
	VipAddress                    string          `xml:"vipAddress" json:"vipAddress"`
	SecureVipAddress              string          `xml:"secureVipAddress,omitempty" json:"secureVipAddress,omitempty"`
	Status                        string          `xml:"status" json:"status"`
	Port                          *Port           `xml:"port,omitempty" json:"port,omitempty"`
	SecurePort                    *Port           `xml:"securePort,omitempty" json:"securePort,omitempty"`
	DataCenterInfo                *DataCenterInfo `xml:"dataCenterInfo" json:"dataCenterInfo"`
	LeaseInfo                     *LeaseInfo      `xml:"leaseInfo,omitempty" json:"leaseInfo,omitempty"`
	MetaData                      *MetaData       `xml:"metadata,omitempty" json:"metadata,omitempty"`
	IsCoordinatingDiscoveryServer string          `xml:"isCoordinatingDiscoveryServer,omitempty" json:"isCoordinatingDiscoveryServer,omitempty"`
	LastUpdatedTimestamp          string          `xml:"lastUpdatedTimestamp,omitempty" json:"lastUpdatedTimestamp,omitempty"`
	LastDirtyTimestamp            string          `xml:"lastDirtyTimestamp,omitempty" json:"lastDirtyTimestamp,omitempty"`
	ActionType                    string          `xml:"actionType,omitempty" json:"actionType,omitempty"`
	OverriddenStatus              string          `xml:"overriddenstatus,omitempty" json:"overriddenstatus,omitempty"`
	CountryID                     int             `xml:"countryId,omitempty" json:"countryId,omitempty"`
	InstanceID                    string          `xml:"instanceId,omitempty" json:"instanceId,omitempty"`
}

// Port 端口
type Port struct {
	Port    int    `xml:",chardata" json:"$"`
	Enabled string `xml:"enabled,attr" json:"@enabled"`
}

// DataCenterInfo 数据中心信息
type DataCenterInfo struct {
	Name     string              `xml:"name" json:"name"`
	Class    string              `xml:"class,attr" json:"@class"`
	MetaData *DataCenterMetaData `xml:"metadata,omitempty" json:"metadata,omitempty"`
}

// DataCenterMetaData 数据中心信息元数据
type DataCenterMetaData struct {
	AmiLaunchIndex   string `xml:"ami-launch-index,omitempty" json:"ami-launch-index,omitempty"`
	LocalHostname    string `xml:"local-hostname,omitempty" json:"local-hostname,omitempty"`
	AvailabilityZone string `xml:"availability-zone,omitempty" json:"availability-zone,omitempty"`
	InstanceID       string `xml:"instance-id,omitempty" json:"instance-id,omitempty"`
	PublicIpv4       string `xml:"public-ipv4,omitempty" json:"public-ipv4,omitempty"`
	PublicHostname   string `xml:"public-hostname,omitempty" json:"public-hostname,omitempty"`
	AmiManifestPath  string `xml:"ami-manifest-path,omitempty" json:"ami-manifest-path,omitempty"`
	LocalIpv4        string `xml:"local-ipv4,omitempty" json:"local-ipv4,omitempty"`
	Hostname         string `xml:"hostname,omitempty" json:"hostname,omitempty"`
	AmiID            string `xml:"ami-id,omitempty" json:"ami-id,omitempty"`
	InstanceType     string `xml:"instance-type,omitempty" json:"instance-type,omitempty"`
}

// LeaseInfo 续约信息
type LeaseInfo struct {
	RenewalIntervalInSecs int `xml:"renewalIntervalInSecs,omitempty" json:"renewalIntervalInSecs,omitempty"`
	DurationInSecs        int `xml:"durationInSecs,omitempty" json:"durationInSecs,omitempty"`
}

// MetaData 元数据
type MetaData map[string]interface{}

// NewInstance 创建服务实例
func NewInstance(app, ip string, port int) *Instance {
	instance := &Instance{
		InstanceID: fmt.Sprintf("%s:%s:%d", ip, app, port),
		HostName:   ip,
		App:        app,
		IPAddr:     ip,
		Port: &Port{
			Port:    port,
			Enabled: "true",
		},
		Status:           "UP",
		OverriddenStatus: "UNKNOWN",
		DataCenterInfo: &DataCenterInfo{
			Name:  "MyOwn",
			Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
		},
		MetaData: nil,
	}
	instance.HomePageURL = fmt.Sprintf("http://%s:%d", ip, port)
	instance.StatusPageURL = fmt.Sprintf("http://%s:%d/info", ip, port)
	return instance
}
