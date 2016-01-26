package model

import (
	"fmt"
	"log"
	"strings"

	"github.com/deis/router/utils"
	modelerUtility "github.com/deis/router/utils/modeler"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

// RouterConfig is the primary type used to encapsulate all router configuration.
type RouterConfig struct {
	WorkerProcesses          string      `router:"workerProcesses"`
	MaxWorkerConnections     int         `router:"maxWorkerConnections"`
	DefaultTimeout           int         `router:"defaultTimeout"`
	ServerNameHashMaxSize    int         `router:"serverNameHashMaxSize"`
	ServerNameHashBucketSize int         `router:"serverNameHashBucketSize"`
	GzipConfig               *GzipConfig `router:"gzip"`
	BodySize                 int         `router:"bodySize"`
	ProxyRealIPCIDR          string      `router:"proxyRealIpCidr"`
	ErrorLogLevel            string      `router:"errorLogLevel"`
	DefaultDomain            string      `router:"defaultDomain"`
	UseProxyProtocol         bool        `router:"useProxyProtocol"`
	EnforceWhitelists        bool        `router:"enforceWhitelists"`
	SSLConfig                *SSLConfig  `router:"ssl"`
	AppConfigs               []*AppConfig
	BuilderConfig            *BuilderConfig
	DefaultCertificate       *Certificate
}

func newRouterConfig() *RouterConfig {
	return &RouterConfig{
		WorkerProcesses:          "auto",
		MaxWorkerConnections:     768,
		DefaultTimeout:           1300,
		ServerNameHashMaxSize:    512,
		ServerNameHashBucketSize: 64,
		GzipConfig:               newGzipConfig(),
		BodySize:                 1,
		ProxyRealIPCIDR:          "10.0.0.0/8",
		ErrorLogLevel:            "error",
		UseProxyProtocol:         false,
		EnforceWhitelists:        false,
		SSLConfig:                newSSLConfig(),
	}
}

// GzipConfig encapsulates gzip configuration.
type GzipConfig struct {
	Enabled     bool   `router:"enabled"`
	CompLevel   int    `router:"compLevel"`
	Disable     string `router:"disable"`
	HTTPVersion string `router:"httpVersion"`
	MinLength   int    `router:"minLength"`
	Proxied     string `router:"proxied"`
	Types       string `router:"types"`
	Vary        string `router:"vary"`
}

func newGzipConfig() *GzipConfig {
	return &GzipConfig{
		Enabled:     true,
		CompLevel:   5,
		Disable:     "msie6",
		HTTPVersion: "1.1",
		MinLength:   256,
		Proxied:     "any",
		Types:       "application/atom+xml application/javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component",
		Vary:        "on",
	}
}

// AppConfig encapsulates the configuration for all routes to a single back end.
type AppConfig struct {
	Domains        []string `router:"domains"`
	Whitelist      []string `router:"whitelist"`
	ConnectTimeout int      `router:"connectTimeout"`
	TCPTimeout     int      `router:"tcpTimeout"`
	ServiceIP      string
	Certificates   map[string]*Certificate
}

func newAppConfig(routerConfig *RouterConfig) *AppConfig {
	return &AppConfig{
		ConnectTimeout: 30,
		TCPTimeout:     routerConfig.DefaultTimeout,
		Certificates:   make(map[string]*Certificate, 0),
	}
}

// BuilderConfig encapsulates the configuration of the deis-builder-- if it's in use.
type BuilderConfig struct {
	ConnectTimeout int `router:"connectTimeout"`
	TCPTimeout     int `router:"tcpTimeout"`
	ServiceIP      string
}

func newBuilderConfig() *BuilderConfig {
	return &BuilderConfig{
		ConnectTimeout: 10,
		TCPTimeout:     1200,
	}
}

// Certificate represents an SSL certificate for use in securing routable applications.
type Certificate struct {
	Cert string
	Key  string
}

func newCertificate(cert string, key string) *Certificate {
	return &Certificate{
		Cert: cert,
		Key:  key,
	}
}

// SSLConfig represents SSL-related configuration options.
type SSLConfig struct {
	Enforce        bool        `router:"enforce"`
	Protocols      string      `router:"protocols"`
	Ciphers        string      `router:"ciphers"`
	SessionCache   string      `router:"sessionCache"`
	SessionTimeout int         `router:"sessionTimeout"`
	SessionTickets string      `router:"sessionTickets"`
	BufferSize     int         `router:"bufferSize"`
	HSTSConfig     *HSTSConfig `router:"hsts"`
	DHParam        string
}

func newSSLConfig() *SSLConfig {
	return &SSLConfig{
		Enforce:        false,
		Protocols:      "TLSv1 TLSv1.1 TLSv1.2",
		SessionTimeout: 10,
		SessionTickets: "on",
		BufferSize:     4,
		HSTSConfig:     newHSTSConfig(),
	}
}

// HSTSConfig represents configuration options having to do with HTTP Strict Transport Security.
type HSTSConfig struct {
	Enabled           bool `router:"enabled"`
	MaxAge            int  `router:"maxAge"`
	IncludeSubDomains bool `router:"includeSubDomains"`
	Preload           bool `router:"preload"`
}

func newHSTSConfig() *HSTSConfig {
	return &HSTSConfig{
		Enabled:           false,
		MaxAge:            15552000, // 180 days
		IncludeSubDomains: false,
		Preload:           false,
	}
}

var (
	namespace = utils.GetOpt("POD_NAMESPACE", "default")
	modeler   = modelerUtility.NewModeler("router.deis.io", "router", "constraint", true)
)

// Build creates a RouterConfig configuration object by querying the k8s API for
// relevant metadata concerning itself and all routable services.
func Build(kubeClient *client.Client) (*RouterConfig, error) {
	// Get all relevant information from k8s:
	//   deis-router rc
	//   All services with label "routable=true"
	//   deis-builder service, if it exists
	// These are used to construct a model...
	routerRC, err := getRC(kubeClient)
	if err != nil {
		return nil, err
	}
	appServices, err := getAppServices(kubeClient)
	if err != nil {
		return nil, err
	}
	// builderService might be nil if it's not found and that's ok.
	builderService, err := getBuilderService(kubeClient)
	if err != nil {
		return nil, err
	}
	defaultCertSecret, err := getSecret(kubeClient, "deis-router-default-cert", namespace)
	if err != nil {
		return nil, err
	}
	dhParamSecret, err := getSecret(kubeClient, "deis-router-dhparam", namespace)
	if err != nil {
		return nil, err
	}
	// Build the model...
	routerConfig, err := build(kubeClient, routerRC, defaultCertSecret, dhParamSecret, appServices, builderService)
	if err != nil {
		return nil, err
	}
	return routerConfig, nil
}

func getRC(kubeClient *client.Client) (*api.ReplicationController, error) {
	rcClient := kubeClient.ReplicationControllers(namespace)
	rc, err := rcClient.Get("deis-router")
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func getAppServices(kubeClient *client.Client) (*api.ServiceList, error) {
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

// getBuilderService will return the service named "deis-builder" from the same namespace as
// the router, but will return nil (without error) if no such service exists.
func getBuilderService(kubeClient *client.Client) (*api.Service, error) {
	serviceClient := kubeClient.Services(namespace)
	service, err := serviceClient.Get("deis-builder")
	if err != nil {
		statusErr, ok := err.(*errors.StatusError)
		// If the issue is just that no deis-builder was found, that's ok.
		if ok && statusErr.Status().Code == 404 {
			// We'll just return nil instead of a found *api.Service.
			return nil, nil
		}
		return nil, err
	}
	return service, nil
}

func getSecret(kubeClient *client.Client, name string, ns string) (*api.Secret, error) {
	secretClient := kubeClient.Secrets(ns)
	secret, err := secretClient.Get(name)
	if err != nil {
		statusErr, ok := err.(*errors.StatusError)
		// If the issue is just that no such secret was found, that's ok.
		if ok && statusErr.Status().Code == 404 {
			// We'll just return nil instead of a found *api.Secret
			return nil, nil
		}
		return nil, err
	}
	return secret, nil
}

func build(kubeClient *client.Client, routerRC *api.ReplicationController, defaultCertSecret *api.Secret, dhParamSecret *api.Secret, appServices *api.ServiceList, builderService *api.Service) (*RouterConfig, error) {
	routerConfig, err := buildRouterConfig(routerRC, defaultCertSecret, dhParamSecret)
	if err != nil {
		return nil, err
	}
	for _, appService := range appServices.Items {
		appConfig, err := buildAppConfig(kubeClient, appService, routerConfig)
		if err != nil {
			return nil, err
		}
		if appConfig != nil {
			routerConfig.AppConfigs = append(routerConfig.AppConfigs, appConfig)
		}
	}
	if builderService != nil {
		builderConfig, err := buildBuilderConfig(builderService)
		if err != nil {
			return nil, err
		}
		if builderConfig != nil {
			routerConfig.BuilderConfig = builderConfig
		}
	}
	return routerConfig, nil
}

func buildRouterConfig(rc *api.ReplicationController, defaultCertSecret *api.Secret, dhParamSecret *api.Secret) (*RouterConfig, error) {
	routerConfig := newRouterConfig()
	err := modeler.MapToModel(rc.Annotations, routerConfig)
	if err != nil {
		return nil, err
	}
	if defaultCertSecret != nil {
		defaultCertificate, err := buildCertificate(defaultCertSecret, "default")
		if err != nil {
			return nil, err
		}
		routerConfig.DefaultCertificate = defaultCertificate
	}
	if dhParamSecret != nil {
		dhParam, err := buildDHParam(dhParamSecret)
		if err != nil {
			return nil, err
		}
		routerConfig.SSLConfig.DHParam = dhParam
	}
	return routerConfig, nil
}

func buildAppConfig(kubeClient *client.Client, service api.Service, routerConfig *RouterConfig) (*AppConfig, error) {
	appConfig := newAppConfig(routerConfig)
	err := modeler.MapToModel(service.Annotations, appConfig)
	if err != nil {
		return nil, err
	}
	// If no domains are found, we don't have the information we need to build routes
	// to this application.  Abort.
	if len(appConfig.Domains) == 0 {
		return nil, nil
	}
	// Step through the domains, and decide which cert, if any, will be used for securing each.
	// For each that is a FQDN, we'll look to see if a corresponding cert-bearing secret also
	// exists.  If so, that will be used.  If a domain isn't an FQDN OR a it is, but a corresponding
	// cert-bearing secret does not exist, we will use the default cert-- even if that is nil.
	for _, domain := range appConfig.Domains {
		if strings.Contains(domain, ".") {
			// Look for a cert-bearing secret for this domain.
			var secretName string
			if strings.HasPrefix(domain, "*.") {
				secretName = fmt.Sprintf("%s-wildcard-cert", strings.TrimPrefix(domain, "*."))
			} else {
				secretName = fmt.Sprintf("%s-cert", domain)
			}
			certSecret, err := getSecret(kubeClient, secretName, service.Namespace)
			if err != nil {
				return nil, err
			}
			if certSecret == nil {
				appConfig.Certificates[domain] = routerConfig.DefaultCertificate
			} else {
				certificate, err := buildCertificate(certSecret, domain)
				if err != nil {
					return nil, err
				}
				appConfig.Certificates[domain] = certificate
			}
		} else {
			appConfig.Certificates[domain] = routerConfig.DefaultCertificate
		}
	}
	appConfig.ServiceIP = service.Spec.ClusterIP
	return appConfig, nil
}

func buildBuilderConfig(service *api.Service) (*BuilderConfig, error) {
	builderConfig := newBuilderConfig()
	builderConfig.ServiceIP = service.Spec.ClusterIP
	err := modeler.MapToModel(service.Annotations, builderConfig)
	if err != nil {
		return nil, err
	}
	return builderConfig, nil
}

func buildCertificate(certSecret *api.Secret, context string) (*Certificate, error) {
	cert, ok := certSecret.Data["cert"]
	// If no cert is found in the secret, warn and return nil
	if !ok {
		log.Printf("WARN: The k8s secret intended to convey the %s certificate contained no entry \"cert\".\n", context)
		return nil, nil
	}
	key, ok := certSecret.Data["key"]
	// If no key is found in the secret, warn and return nil
	if !ok {
		log.Printf("WARN: The k8s secret intended to convey the %s certificate key contained no entry \"key\".\n", context)
		return nil, nil
	}
	certStr := string(cert[:])
	keyStr := string(key[:])
	return newCertificate(certStr, keyStr), nil
}

func buildDHParam(dhParamSecret *api.Secret) (string, error) {
	dhParam, ok := dhParamSecret.Data["dhparam"]
	// If no dhparam is found in the secret, warn and return ""
	if !ok {
		log.Println("WARN: The k8s secret intended to convey the dhparam contained no entry \"dhparam\".")
		return "", nil
	}
	return string(dhParam), nil
}
