package nginx

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/deis/router/model"
)

func TestWriteCerts(t *testing.T) {
	sslPath, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(sslPath)

	// Create an extra crt/key pair to ensure they are correctly removed.
	certPath := filepath.Join(sslPath, "extra.crt")
	keyPath := filepath.Join(sslPath, "extra.key")
	err = ioutil.WriteFile(certPath, []byte("foo"), 0644)
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile(keyPath, []byte("bar"), 0600)
	if err != nil {
		t.Error(err)
	}

	expectedPlatformCrt := "platform-biz"
	expectedPlatformKey := "platform-baz"
	expectedExampleCrt := "examplecom-crt"
	expectedExampleKey := "examplecom-key"
	routerConfig := model.RouterConfig{
		PlatformCertificate: &model.Certificate{
			Cert: expectedPlatformCrt,
			Key:  expectedPlatformKey,
		},
		AppConfigs: []*model.AppConfig{
			&model.AppConfig{
				Certificates: map[string]*model.Certificate{
					"example.com": &model.Certificate{
						Cert: expectedExampleCrt,
						Key:  expectedExampleKey,
					},
				},
			},
		},
	}

	WriteCerts(&routerConfig, sslPath)

	// Any extra crt/key files should be removed.
	if _, err := os.Stat(certPath); err == nil {
		t.Errorf("Expected extra.crt to be removed, but the file was found.")
	}
	if _, err := os.Stat(keyPath); err == nil {
		t.Errorf("Expected extra.key to be removed, but the file was found.")
	}

	// platform.crt and platform.key should exist with correct permissions and contents.
	platformCrtPath := filepath.Join(sslPath, "platform.crt")
	platformKeyPath := filepath.Join(sslPath, "platform.key")
	err = checkCertAndKey(platformCrtPath, platformKeyPath, expectedPlatformCrt, expectedPlatformKey)
	if err != nil {
		t.Error(err)
	}

	// example application crt and key should exist with correct permissions and contents.
	exampleCrtPath := filepath.Join(sslPath, "example.com.crt")
	exampleKeyPath := filepath.Join(sslPath, "example.com.key")
	err = checkCertAndKey(exampleCrtPath, exampleKeyPath, expectedExampleCrt, expectedExampleKey)
	if err != nil {
		t.Error(err)
	}
}

func TestWriteCert(t *testing.T) {
	// Ensure cert/key are written with correct contents and correct permissions.
	expectedCertContents := "foo"
	expectedKeyContents := "bar"
	certificate := model.Certificate{
		Cert: expectedCertContents,
		Key:  expectedKeyContents,
	}

	sslPath, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(sslPath)
	crtPath := filepath.Join(sslPath, "test.crt")
	keyPath := filepath.Join(sslPath, "test.key")

	err = writeCert("test", &certificate, sslPath)
	if err != nil {
		t.Error(err)
	}

	err = checkCertAndKey(crtPath, keyPath, expectedCertContents, expectedKeyContents)
	if err != nil {
		t.Error(err)
	}
}

func TestWriteDHParam(t *testing.T) {
	// Ensure sslPath/dhparam.pem exists with the contents of routerConfig.SSLConfig.DHParam and is 0644
	sslPath, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(sslPath)
	dhParamPath := filepath.Join(sslPath, "dhparam.pem")

	expectedDHParam := "bizbar"
	routerConfig := model.RouterConfig{
		SSLConfig: &model.SSLConfig{
			DHParam: expectedDHParam,
		},
	}

	err = WriteDHParam(&routerConfig, sslPath)
	if err != nil {
		t.Error(err)
	}

	actualDHParam, err := ioutil.ReadFile(dhParamPath)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expectedDHParam, string(actualDHParam)) {
		t.Errorf("Expected dhparam.pem contents, %s, does not match actual contents, %s.", expectedDHParam, string(actualDHParam))
	}

	expectedPerm := "-rw-r--r--" // 0644

	info, _ := os.Stat(dhParamPath)
	actualPerm := info.Mode().String()
	if !reflect.DeepEqual(expectedPerm, actualPerm) {
		t.Errorf("Expected permission on dhparam.pem, %s, does not match actual, %s.", expectedPerm, actualPerm)
	}

	// Ensure dhparam.pem is erased when routerConfig.SSLConfig.DHParam is empty
	sslPath, err = ioutil.TempDir("", "test-empty")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(sslPath)
	dhParamPath = filepath.Join(sslPath, "dhparam.pem")

	routerConfig = model.RouterConfig{
		SSLConfig: &model.SSLConfig{
			DHParam: "",
		},
	}
	err = WriteDHParam(&routerConfig, sslPath)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(dhParamPath); err == nil {
		t.Errorf("Expected dhparam.pem to be erased when DHParam was an empty string, but the file was found.")
	}
}

func TestWriteConfig(t *testing.T) {
	routerConfig := model.RouterConfig{}

	tmpFile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpFile.Name())

	WriteConfig(&routerConfig, tmpFile.Name())

	if _, err := os.Stat(tmpFile.Name()); os.IsNotExist(err) {
		t.Errorf("Expected to find nginx config file. No file found.")
	}
}

func checkCertAndKey(crtPath string, keyPath string, expectedCertContents string, expectedKeyContents string) error {
	actualCertContents, err := ioutil.ReadFile(crtPath)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(expectedCertContents, string(actualCertContents)) {
		return fmt.Errorf("Expected test.crt contents, %s, does not match actual contents, %s.", expectedCertContents, string(actualCertContents))
	}

	actualKeyContents, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expectedKeyContents, string(actualKeyContents)) {
		return fmt.Errorf("Expected test.key contents, %s, does not match actual contents, %s.", expectedKeyContents, string(actualKeyContents))
	}

	expectedCertPerm := "-rw-r--r--" // 0644
	expectedKeyPerm := "-rw-------"  // 0600

	crtInfo, _ := os.Stat(crtPath)
	actualCertPerm := crtInfo.Mode().String()
	if !reflect.DeepEqual(expectedCertPerm, actualCertPerm) {
		return fmt.Errorf("Expected permission on test.crt, %s, does not match actual, %s.", expectedCertPerm, actualCertPerm)
	}

	keyInfo, _ := os.Stat(keyPath)
	actualKeyPerm := keyInfo.Mode().String()
	if !reflect.DeepEqual(expectedKeyPerm, actualKeyPerm) {
		return fmt.Errorf("Expected permission on test.key, %s, does not match actual, %s.", expectedKeyPerm, actualKeyPerm)
	}

	return nil
}

func TestDisableServerTokens(t *testing.T) {
	routerConfig := &model.RouterConfig{
		WorkerProcesses:          "auto",
		MaxWorkerConnections:     "768",
		TrafficStatusZoneSize:    "1m",
		DefaultTimeout:           "1300s",
		ServerNameHashMaxSize:    "512",
		ServerNameHashBucketSize: "64",
		GzipConfig: &model.GzipConfig{
			Enabled:     true,
			CompLevel:   "5",
			Disable:     "msie6",
			HTTPVersion: "1.1",
			MinLength:   "256",
			Proxied:     "any",
			Types:       "application/atom+xml application/javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component",
			Vary:        "on",
		},
		BodySize:          "1m",
		ProxyRealIPCIDRs:  []string{"10.0.0.0/8"},
		ErrorLogLevel:     "error",
		UseProxyProtocol:  false,
		EnforceWhitelists: false,
		WhitelistMode:     "extend",
		SSLConfig: &model.SSLConfig{
			Enforce:           false,
			Protocols:         "TLSv1 TLSv1.1 TLSv1.2",
			SessionTimeout:    "10m",
			UseSessionTickets: true,
			BufferSize:        "4k",
			HSTSConfig: &model.HSTSConfig{
				Enabled:           false,
				MaxAge:            15552000, // 180 days
				IncludeSubDomains: false,
				Preload:           false,
			},
		},

		DisableServerTokens: true,
	}

	var b bytes.Buffer

	tmpl, err := template.New("nginx").Funcs(sprig.TxtFuncMap()).Parse(confTemplate)

	if err != nil {
		t.Fatalf("Encountered an error: %v", err)
	}

	err = tmpl.Execute(&b, routerConfig)

	if err != nil {
		t.Fatalf("Encountered an error: %v", err)
	}

	validDirective := regexp.MustCompile(`(?m)^(\s*)server_tokens off;$`)

	if !validDirective.Match(b.Bytes()) {
		t.Errorf("Expected: 'server_tokens off' in the configuration. Actual: no match")
	}

}
