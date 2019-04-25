package eureka_client

import (
	"encoding/json"
	"errors"
	"net"
)

// GetIpAddrs returns local ip address
func GetIpAddrs() ([]string, error) {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// check ip addr
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	return ips, nil
}

// ParseApplications parse Applications from eureka server
func ParseApplications(result map[string]interface{}) ([]Instance, error) {
	var instances []Instance

	applications, ok := result["applications"].(map[string]interface{})
	if !ok {
		return nil, errors.New("refresh failed, It's not ok for type applications")
	}
	applicationArr, ok := applications["application"].([]interface{})
	if !ok {
		return nil, errors.New("refresh failed, It's not ok for type application arr")
	}
	for _, application := range applicationArr {
		data, ok := application.(map[string]interface{})
		if !ok {
			return nil, errors.New("refresh failed, It's not ok for type application")
		}

		instanceArr, ok := data["instance"].([]interface{})
		if !ok {
			return nil, errors.New("refresh failed, It's not ok for type instance arr")
		}
		for _, v := range instanceArr {
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}

			var instance Instance
			err = json.Unmarshal(b, &instance)
			if err != nil {
				return nil, err
			}

			// filter applications for instances with only UP states
			if instance.Status == "UP" {
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}
