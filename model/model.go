package model

import (
	"encoding/json"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

type RouterConfig struct {
	UseProxyProtocol bool `json:"useProxyProtocol"`
	AppConfigs       []*AppConfig
}

func newRouterConfig() *RouterConfig {
	return &RouterConfig{
		UseProxyProtocol: false,
	}
}

type AppConfig struct {
	Domains   []string `json:"domains"`
	ServiceIP string
	Available bool
}

func newAppConfig() *AppConfig {
	return &AppConfig{}
}

func Build(kubeClient *client.Client) (*RouterConfig, error) {
	rc, err := getRC(kubeClient)
	if err != nil {
		return nil, err
	}
	services, err := getServices(kubeClient)
	if err != nil {
		return nil, err
	}
	routerConfig, err := build(kubeClient, rc, services)
	if err != nil {
		return nil, err
	}
	return routerConfig, nil
}

func getRC(kubeClient *client.Client) (*api.ReplicationController, error) {
	rcClient := kubeClient.ReplicationControllers("deis")
	rc, err := rcClient.Get("deis-router")
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func getServices(kubeClient *client.Client) (*api.ServiceList, error) {
	serviceClient := kubeClient.Services(api.NamespaceAll)
	servicesSelector, err := labels.Parse("routable==true")
	if err != nil {
		return nil, err
	}
	services, err := serviceClient.List(servicesSelector)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func build(kubeClient *client.Client, rc *api.ReplicationController, services *api.ServiceList) (*RouterConfig, error) {
	routerConfig, err := buildRouterConfig(rc)
	if err != nil {
		return nil, err
	}
	for _, service := range services.Items {
		appConfig, err := buildAppConfig(kubeClient, service)
		if err != nil {
			return nil, err
		}
		routerConfig.AppConfigs = append(routerConfig.AppConfigs, appConfig)
	}
	return routerConfig, nil
}

func buildRouterConfig(rc *api.ReplicationController) (*RouterConfig, error) {
	routerConfig := newRouterConfig()
	routerConfigStr := rc.Annotations["routerConfig"]
	if routerConfigStr == "" {
		routerConfigStr = "{}"
	}
	err := json.Unmarshal([]byte(routerConfigStr), routerConfig)
	if err != nil {
		return nil, err
	}
	return routerConfig, nil
}

func buildAppConfig(kubeClient *client.Client, service api.Service) (*AppConfig, error) {
	appConfig := newAppConfig()
	err := json.Unmarshal([]byte(service.Annotations["routerConfig"]), appConfig)
	if err != nil {
		return nil, err
	}
	appConfig.ServiceIP = service.Spec.ClusterIP
	endpointsClient := kubeClient.Endpoints(service.Namespace)
	endpointsSelector, err := labels.Parse(labels.FormatLabels(service.Spec.Selector))
	if err != nil {
		return nil, err
	}
	endpoints, err := endpointsClient.List(endpointsSelector)
	if err != nil {
		return nil, err
	}
	appConfig.Available = len(endpoints.Items[0].Subsets) > 0
	return appConfig, nil
}
