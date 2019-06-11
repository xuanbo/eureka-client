# Eureka-Client

> eureka client by golang

![ui](./doc/eureka-server.jpg)

## Features

* Heartbeat
* Refresh（Only all applications）

## Todo

* Refresh by delta

If the delta is disabled or if it is the first time, get all applications

## Example

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	eureka "github.com/xuanbo/eureka-client"
)

func main() {
	// create eureka client
	client := eureka.NewClient(&eureka.Config{
		DefaultZone: "http://eureka.didispace.com/eureka/",
		App:         "golang-example",
		Port:        10000,
	})
	// start client, register、heartbeat、refresh
	client.Start()

	// http server
	http.HandleFunc("/services", func(writer http.ResponseWriter, request *http.Request) {
		// full applications from eureka server
		apps := client.Applications

		b, _ := json.Marshal(apps)
		_, _ = writer.Write(b)
	})

	// start http server
	if err := http.ListenAndServe(":10000", nil); err != nil {
		fmt.Println(err)
	}
}
```

[examples](./examples/main.go)

## Test

I use `spring-cloud-starter-netflix-eureka-server` in Java.

```xml
<dependency>
    <groupId>org.springframework.cloud</groupId>
    <artifactId>spring-cloud-starter-netflix-eureka-server</artifactId>
    <version>2.1.0.RELEASE</version>
</dependency>
```

The code is as follows:

[spring-cloud-v2](https://github.com/xuanbo/spring-cloud-v2)