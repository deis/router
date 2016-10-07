package model

import (
	"reflect"
	"testing"

	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/util/intstr"
)

const (
	routerName       = "deis-router"
	routerNamespace  = "deis"
	dhParamName      = "deis-router-dhparam"
	platformCertName = "deis-router-platform-cert"
)

func TestBuildRouterConfig(t *testing.T) {
	replicas := int32(1)
	routerDeployment := v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      routerName,
			Namespace: routerNamespace,
			Annotations: map[string]string{
				"router.deis.io/nginx.defaultTimeout":             "1500s",
				"router.deis.io/nginx.ssl.bufferSize":             "6k",
				"router.deis.io/nginx.ssl.hsts.maxAge":            "1234",
				"router.deis.io/nginx.ssl.hsts.includeSubDomains": "true",
			},
			Labels: map[string]string{
				"heritage": "deis",
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Strategy: v1beta1.DeploymentStrategy{
				Type:          v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{},
			},
			Replicas: &replicas,
			Selector: &v1beta1.LabelSelector{MatchLabels: map[string]string{"app": routerName}},
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": routerName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "deis/router",
						},
					},
				},
			},
		},
	}

	platformCertSecret := v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      platformCertName,
			Namespace: routerNamespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"tls.crt": []byte("foo"),
			"tls.key": []byte("bar"),
		},
	}

	dhParamSecret := v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      dhParamName,
			Namespace: routerNamespace,
			Labels: map[string]string{
				"heritage": "deis",
			},
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"dhparam": []byte("bizbaz"),
		},
	}

	expectedConfig := newRouterConfig()
	sslConfig := newSSLConfig()
	hstsConfig := newHSTSConfig()
	platformCert := newCertificate("foo", "bar")

	// A value not set in the deployment annotations (should be default value).
	expectedConfig.MaxWorkerConnections = "768"

	// A simple string value.
	expectedConfig.DefaultTimeout = "1500s"

	// A nested value.
	sslConfig.BufferSize = "6k"
	sslConfig.DHParam = "bizbaz"

	// A nested+nested value.
	hstsConfig.MaxAge = 1234
	hstsConfig.IncludeSubDomains = true

	sslConfig.HSTSConfig = hstsConfig
	expectedConfig.SSLConfig = sslConfig

	expectedConfig.PlatformCertificate = platformCert

	actualConfig, err := buildRouterConfig(&routerDeployment, &platformCertSecret, &dhParamSecret)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expectedConfig, actualConfig) {
		t.Errorf("Expected routerConfig does not match actual.")

		t.Errorf("Expected:\n")
		t.Errorf("%+v\n", expectedConfig)
		t.Errorf("Actual:\n")
		t.Errorf("%+v\n", actualConfig)
	}
}

func TestBuildBuilderConfig(t *testing.T) {
	builderService := v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "deis-builder",
			Namespace: routerNamespace,
			Labels: map[string]string{
				"heritage": "deis",
			},
			Annotations: map[string]string{
				"router.deis.io/nginx.connectTimeout": "20s",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "ssh",
					Port: int32(2222),
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 2223,
					},
				},
			},
			Selector: map[string]string{
				"app": "deis-builder",
			},
			ClusterIP: "1.2.3.4",
		},
	}

	expectedConfig := BuilderConfig{
		// A value  set in the service annotations.
		ConnectTimeout: "20s",
		// A value not set in the service annotations (should be default value).
		TCPTimeout: "1200s",
		// A value determined by the service.spec.ClusterIP
		ServiceIP: "1.2.3.4",
	}

	actualConfig, err := buildBuilderConfig(&builderService)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(&expectedConfig, actualConfig) {
		t.Errorf("Expected builderConfig does not match actual.")

		t.Errorf("Expected:\n")
		t.Errorf("%+v\n", &expectedConfig)
		t.Errorf("Actual:\n")
		t.Errorf("%+v\n", actualConfig)
	}

}
