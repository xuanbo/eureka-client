# Eureka-Client

> eureka client by golang

## Features

* Heartbeat
* Refresh（Only all applications）

## TODO

* Refresh by delta

If the delta is disabled or if it is the first time, get all applications

## Example

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	client "github.com/xuanbo/eureka-client"
)

func main() {
	// create eureka client
	c := client.NewClient(&client.EurekaClientConfig{
		DefaultZone: "http://127.0.0.1:8080/eureka/",
		App:         "golang-example",
		Port:        10000,
	})
	// start client, register、heartbeat、refresh
	c.Start()

	// Go signal notification works by sending `os.Signal`
	// values on a channel. We'll create a channel to
	// receive these notifications (we'll also make one to
	// notify us when the program can exit).
	sigs := make(chan os.Signal)
	exit := make(chan bool, 1)
	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// http server
	http.HandleFunc("/services", func(writer http.ResponseWriter, request *http.Request) {
		// full applications from eureka server
		services := c.Services
		
		b, _ := json.Marshal(services)
		_, _ = writer.Write(b)
	})
	server := &http.Server{
		Addr:    ":10000",
		Handler: http.DefaultServeMux,
	}

	// start http server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()

	// shutdown
	// This goroutine executes a blocking receive for
	// signals. When it gets one it'll print it out
	// and then notify the program that it can finish.
	go func() {
		// receive for signals
		fmt.Println(<-sigs)

		// shutdown http server
		if err := server.Close(); err != nil {
			panic(err)
		}

		// shutdown eureka client, unregister
		c.Shutdown()

		// notify the program that it can finish.
		exit <- true
	}()

	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `exit`) and then exit.
	<-exit
}
```

[examples](./examples/main.go)