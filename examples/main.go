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
