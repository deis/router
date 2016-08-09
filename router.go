package main

import (
	"log"
	"reflect"

	"github.com/deis/router/model"
	"github.com/deis/router/nginx"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/flowcontrol"
)

func main() {
	nginx.Start()
	kubeClient, err := client.NewInCluster()
	if err != nil {
		log.Fatalf("Failed to create client: %v.", err)
	}
	rateLimiter := flowcontrol.NewTokenBucketRateLimiter(0.1, 1)
	known := &model.RouterConfig{}
	// Main loop
	for {
		rateLimiter.Accept()
		routerConfig, err := model.Build(kubeClient)
		if err != nil {
			log.Printf("Error building model; not modifying certs or configuration: %v.", err)
			continue
		}
		if reflect.DeepEqual(routerConfig, known) {
			continue
		}
		log.Println("INFO: Router configuration has changed in k8s.")
		err = nginx.WriteCerts(routerConfig, "/opt/router/ssl")
		if err != nil {
			log.Printf("Failed to write certs; continuing with existing certs, dhparam, and configuration: %v", err)
			continue
		}
		err = nginx.WriteDHParam(routerConfig, "/opt/router/ssl")
		if err != nil {
			log.Printf("Failed to write dhparam; continuing with existing dhparam and configuration: %v", err)
			continue
		}
		err = nginx.WriteConfig(routerConfig, "/opt/router/conf/nginx.conf")
		if err != nil {
			log.Printf("Failed to write new nginx configuration; continuing with existing configuration: %v", err)
			continue
		}
		nginx.Reload()
		known = routerConfig
	}
}
