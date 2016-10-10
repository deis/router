package main

import (
	"log"
	"reflect"

	"github.com/deis/router/model"
	"github.com/deis/router/nginx"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/util/flowcontrol"
	"k8s.io/client-go/1.4/rest"
)

func main() {
	nginx.Start()
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
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
		err = nginx.Reload()
		if err != nil {
			log.Printf("Failed to reload nginx; continuing with existing configuration: %v", err)
			continue
		}
		known = routerConfig
	}
}
