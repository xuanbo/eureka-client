package eureka_client

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/xuanbo/requests"
)

var lastDirtyTimestamp = time.Now().UnixNano() / 1e6

// 与eureka服务端rest交互

// Register 注册实例
// POST /eureka/v2/apps/appID
func Register(zone, app string, instance *Instance) error {
	// Instance 服务实例
	type InstanceInfo struct {
		Instance *Instance `json:"instance"`
	}
	var info = &InstanceInfo{
		Instance: instance,
	}

	url := zone + "apps/" + app

	// status: http.StatusNoContent
	result := requests.Post(url).Json(info).Send().Status2xx()
	if result.Err != nil {
		return fmt.Errorf("Register application instance failed, error: %s", result.Err)
	}
	return nil
}

// UnRegister 删除实例
// DELETE /eureka/v2/apps/appID/instanceID
func UnRegister(zone, app, instanceID string) error {
	url := zone + "apps/" + app + "/" + instanceID
	// status: http.StatusNoContent
	result := requests.Delete(url).Send().StatusOk()
	if result.Err != nil {
		return fmt.Errorf("UnRegister application instance failed, error: %s", result.Err)
	}
	return nil
}

// Refresh 查询所有服务实例
// GET /eureka/v2/apps
func Refresh(zone string) (*Applications, error) {
	type Result struct {
		Applications *Applications `json:"applications"`
	}
	apps := new(Applications)
	res := &Result{
		Applications: apps,
	}
	url := zone + "apps"
	err := requests.Get(url).Header("Accept", " application/json").Send().StatusOk().Json(res)
	if err != nil {
		return nil, fmt.Errorf("Refresh failed, error: %s", err)
	}
	return apps, nil
}

// Heartbeat 发送心跳
// PUT /eureka/v2/apps/appID/instanceID
func Heartbeat(zone, app, instanceID string) error {
	u := zone + "apps/" + app + "/" + instanceID
	params := url.Values{
		"status": {"UP"},
		// 数据未变化，一直使用之前的时间戳
		"lastDirtyTimestamp": {strconv.FormatInt(lastDirtyTimestamp, 10)},
	}
	result := requests.Put(u).Params(params).Send().StatusOk()
	if result.Err != nil {
		return fmt.Errorf("Heartbeat failed, error: %s", result.Err)
	}
	return nil
}
