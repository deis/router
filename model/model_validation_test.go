package model

import (
	"reflect"
	"testing"

	modelerUtility "github.com/deis/router/utils/modeler"
)

var (
	// We'll test the declarative constraints on model attributes using this modeler instead of the
	// one defined in the non-test half of this package.  Why?  We want one that generates errors
	// instead of merely outputting warnings.  That allows us to easily assert validation failures
	// and also prevents us from clutter STDOUT with useless noise.
	testModeler = modelerUtility.NewModeler("", modelerFieldTag, modelerConstraintTag, false)
)

func TestInvalidWorkerProcesses(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "WorkerProcesses", "workerProcesses", []string{"0", "-1", "foobar"})
}

func TestValidWorkerProcesses(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "WorkerProcesses", "workerProcesses", []string{"auto", "2", "10"})
}

func TestInvalidMaxWorkerConnections(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "MaxWorkerConnections", "maxWorkerConnections", []string{"0", "-1", "foobar"})
}

func TestValidMaxWorkerConnections(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "MaxWorkerConnections", "maxWorkerConnections", []string{"1", "2", "10"})
}

func TestInvalidTrafficStatusZoneSize(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "TrafficStatusZoneSize", "trafficStatusZoneSize", []string{"0", "-1", "foobar"})
}

func TestValidTrafficStatusZoneSize(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "TrafficStatusZoneSize", "trafficStatusZoneSize", []string{"1", "2", "20", "1k", "2k", "10m", "10M"})
}

func TestInvalidDefaultTimeout(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "DefaultTimeout", "defaultTimeout", []string{"0", "-1", "foobar"})
}

func TestValidDefaultTimeout(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "DefaultTimeout", "defaultTimeout", []string{"1", "2", "10", "1ms", "2s", "10m"})
}

func TestInvalidServerNameHashMaxSize(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "ServerNameHashMaxSize", "serverNameHashMaxSize", []string{"0", "-1", "foobar"})
}

func TestValidServerNameHashMaxSize(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "ServerNameHashMaxSize", "serverNameHashMaxSize", []string{"1", "2", "20", "1k", "2k", "10m", "10M"})
}

func TestInvalidServerNameHashBucketSize(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "ServerNameHashBucketSize", "serverNameHashBucketSize", []string{"0", "-1", "foobar"})
}

func TestValidServerNameHashBucketSize(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "ServerNameHashBucketSize", "serverNameHashBucketSize", []string{"1", "2", "20", "1k", "2k", "10m", "10M"})
}

func TestInvalidBodySize(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "BodySize", "bodySize", []string{"-1", "foobar"})
}

func TestValidBodySize(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "BodySize", "bodySize", []string{"1", "2", "20", "1k", "2k", "10m", "10M"})
}

func TestInvalidProxyRealIPCIDRs(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "ProxyRealIPCIDRs", "proxyRealIpCidrs", []string{"0", "-1", "foobar"})
}

func TestValidProxyRealIPCIDRs(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "ProxyRealIPCIDRs", "proxyRealIpCidrs", []string{"0.0.0.0/0", "10.0.0.0/16", "10.0.0.0/16,192.168.0.0/16", "10.0.0.0/16, 192.168.0.0/16", "10.0.0.0/16 ,192.168.0.0/16", "10.0.0.0/16 , 192.168.0.0/16"})
}

func TestInvalidErrorLogLevel(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "ErrorLogLevel", "errorLogLevel", []string{"0", "-1", "foobar"})
}

func TestValidErrorLogLevel(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "ErrorLogLevel", "errorLogLevel", []string{"info", "notice", "warn"})
}

func TestInvalidPlatformDomain(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "PlatformDomain", "platformDomain", []string{"0", "-1", "foobar", "foo_bar.com", "foobar.c"})
}

func TestValidPlatformDomain(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "PlatformDomain", "platformDomain", []string{"foobar.com", "foo-bar.io"})
}

func TestInvalidUseProxyProtocol(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "UseProxyProtocol", "useProxyProtocol", []string{"0", "-1", "foobar"})
}

func TestValidUseProxyProtocol(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "UseProxyProtocol", "useProxyProtocol", []string{"true", "false", "TRUE", "FALSE"})
}

func TestValidServerTokens(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "DisableServerTokens", "disableServerTokens", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidServerTokens(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "DisableServerTokens", "disableServerTokens", []string{"0", "-1", "foobar"})
}

func TestInvalidEnforceWhitelists(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "EnforceWhitelists", "enforceWhitelists", []string{"0", "-1", "foobar"})
}

func TestValidEnforceWhitelists(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "EnforceWhitelists", "enforceWhitelists", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidDefaultWhitelist(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "DefaultWhitelist", "defaultWhitelist", []string{"0", "-1", "foobar"})
}

func TestValidDefaultWhitelist(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "DefaultWhitelist", "defaultWhitelist", []string{"1.2.3.4", "0.0.0.0/0", "1.2.3.4,0.0.0.0/0", "1.2.3.4, 0.0.0.0/0"})
}

func TestInvalidWhitelistMode(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "WhitelistMode", "whitelistMode", []string{"0", "-1", "foobar"})
}

func TestValidWhitelistMode(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "WhitelistMode", "whitelistMode", []string{"extend", "override"})
}

func TestValidHTTP2Enabled(t *testing.T) {
	testValidValues(t, newTestRouterConfig, "HttpEnabled", "http2Enabled", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidHTTP2Enabled(t *testing.T) {
	testInvalidValues(t, newTestRouterConfig, "HTTP2Enabled", "http2Enabled", []string{"0", "-1", "foobar"})
}

func TestInvalidGzipEnabled(t *testing.T) {
	testInvalidValues(t, newTestGzipConfig, "Enabled", "enabled", []string{"0", "-1", "foobar"})
}

func TestValidGzipEnabled(t *testing.T) {
	testValidValues(t, newTestGzipConfig, "Enabled", "enabled", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidGzipCompLevel(t *testing.T) {
	testInvalidValues(t, newTestGzipConfig, "CompLevel", "compLevel", []string{"0", "-1", "foobar"})
}

func TestValidGzipCompLevel(t *testing.T) {
	testValidValues(t, newTestGzipConfig, "CompLevel", "compLevel", []string{"1", "2", "3", "4"})
}

func TestInvalidGzipHTTPVersion(t *testing.T) {
	testInvalidValues(t, newTestGzipConfig, "HTTPVersion", "httpVersion", []string{"0", "-1", "foobar"})
}

func TestValidGzipHTTPVersion(t *testing.T) {
	testValidValues(t, newTestGzipConfig, "HTTPVersion", "httpVersion", []string{"1.0", "1.1"})
}

func TestInvalidGzipMinLength(t *testing.T) {
	testInvalidValues(t, newTestGzipConfig, "MinLength", "minLength", []string{"-1", "foobar"})
}

func TestValidGzipMinLength(t *testing.T) {
	testValidValues(t, newTestGzipConfig, "MinLength", "minLength", []string{"0", "1", "2", "20"})
}

func TestInvalidGzipProxied(t *testing.T) {
	testInvalidValues(t, newTestGzipConfig, "Proxied", "proxied", []string{"0", "-1", "foobar"})
}

func TestValidGzipProxied(t *testing.T) {
	testValidValues(t, newTestGzipConfig, "Proxied", "proxied", []string{"off", "expired", "no-cache", "no-store private no_etag"})
}

func TestInvalidGzipTypes(t *testing.T) {
	testInvalidValues(t, newTestGzipConfig, "Types", "types", []string{"0", "-1", "foobar"})
}

func TestValidGzipTypes(t *testing.T) {
	testValidValues(t, newTestGzipConfig, "Types", "types", []string{"application/json", "application/json application/text"})
}

func TestInvalidGzipVary(t *testing.T) {
	testInvalidValues(t, newTestGzipConfig, "Vary", "vary", []string{"0", "-1", "foobar"})
}

func TestValidGzipVary(t *testing.T) {
	testValidValues(t, newTestGzipConfig, "Vary", "vary", []string{"on", "off"})
}

func TestInvalidAppDomains(t *testing.T) {
	testInvalidValues(t, newTestAppConfig, "Domains", "domains", []string{"-1", "foo_bar", "foobar.c", "foo bar"})
}

func TestValidAppDomains(t *testing.T) {
	testValidValues(t, newTestAppConfig, "Domains", "domains", []string{"foobar", "foo-bar", "foobar.com", "foobar,foobar.com", "foobar, foobar.com", "*.foobar.com", "xn--eckwd4c7c.xn--zckzah", "xn--80ahd1agd.ru", "xn--tst-qla.xn--knigsgsschen-lcb0w.de"})
}

func TestInvalidAppWhitelist(t *testing.T) {
	testInvalidValues(t, newTestAppConfig, "Whitelist", "whitelist", []string{"0", "-1", "foobar"})
}

func TestValidAppWhitelist(t *testing.T) {
	testValidValues(t, newTestAppConfig, "Whitelist", "whitelist", []string{"1.2.3.4", "0.0.0.0/0", "1.2.3.4,0.0.0.0/0", "1.2.3.4, 0.0.0.0/0"})
}

func TestInvalidAppConnectTimeout(t *testing.T) {
	testInvalidValues(t, newTestAppConfig, "ConnectTimeout", "connectTimeout", []string{"0", "-1", "foobar"})
}

func TestValidAppConnectTimeout(t *testing.T) {
	testValidValues(t, newTestAppConfig, "ConnectTimeout", "connectTimeout", []string{"1", "2", "10", "1ms", "2s", "10m"})
}

func TestInvalidAppTCPTimeout(t *testing.T) {
	testInvalidValues(t, newTestAppConfig, "TCPTimeout", "tcpTimeout", []string{"0", "-1", "foobar"})
}

func TestValidAppTCPTimeout(t *testing.T) {
	testValidValues(t, newTestAppConfig, "TCPTimeout", "tcpTimeout", []string{"1", "2", "10", "1ms", "2s", "10m"})
}

func TestInvalidCertMappings(t *testing.T) {
	testInvalidValues(t, newTestAppConfig, "CertMappings", "certificates", []string{"0", "-1", "foobar"})
}

func TestValidCertMappings(t *testing.T) {
	testValidValues(t, newTestAppConfig, "CertMappings", "certificates", []string{"foobar.com:foobar,*.foobar.deis.ninja:foobar-deis-ninja"})
}

func TestInvalidBuilderConnectTimeout(t *testing.T) {
	testInvalidValues(t, newTestBuilderConfig, "ConnectTimeout", "connectTimeout", []string{"0", "-1", "foobar"})
}

func TestValidBuilderConnectTimeout(t *testing.T) {
	testValidValues(t, newTestBuilderConfig, "ConnectTimeout", "connectTimeout", []string{"1", "2", "10", "1ms", "2s", "10m"})
}

func TestInvalidBuilderTCPTimeout(t *testing.T) {
	testInvalidValues(t, newTestBuilderConfig, "TCPTimeout", "tcpTimeout", []string{"0", "-1", "foobar"})
}

func TestValidBuilderTCPTimeout(t *testing.T) {
	testValidValues(t, newTestBuilderConfig, "TCPTimeout", "tcpTimeout", []string{"1", "2", "10", "1ms", "2s", "10m"})
}

func TestInvalidSSLEnforce(t *testing.T) {
	testInvalidValues(t, newTestSSLConfig, "Enforce", "enforce", []string{"0", "-1", "foobar"})
}

func TestValidSSLEnforce(t *testing.T) {
	testValidValues(t, newTestSSLConfig, "Enforce", "enforce", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidSSLProtocols(t *testing.T) {
	testInvalidValues(t, newTestSSLConfig, "Protocols", "protocols", []string{"0", "-1", "foobar"})
}

func TestValidSSLProtocols(t *testing.T) {
	testValidValues(t, newTestSSLConfig, "Protocols", "protocols", []string{"SSLv3", "TLSv1", "TLSv1 TLSv1.1"})
}

func TestInvalidSSLCiphers(t *testing.T) {
	testInvalidValues(t, newTestSSLConfig, "Ciphers", "ciphers", []string{"0", "-1", "foobar"})
}

func TestValidSSLCiphers(t *testing.T) {
	testValidValues(t, newTestSSLConfig, "Ciphers", "ciphers", []string{"DHE-RSA-AES256-SHA", "DHE-RSA-AES256-SHA:DHE-DSS-AES256-SHA:AES256-SHA", "EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5"})
}

func TestInvalidSSLSessionCache(t *testing.T) {
	testInvalidValues(t, newTestSSLConfig, "SessionCache", "sessionCache", []string{"0", "-1", "foobar"})
}

func TestValidSSLSessionCache(t *testing.T) {
	testValidValues(t, newTestSSLConfig, "SessionCache", "sessionCache", []string{"off", "none", "builtin", "builtin:1000", "builtin:1000 shared:SSL:16k"})
}

func TestInvalidSSLSessionTimeout(t *testing.T) {
	testInvalidValues(t, newTestSSLConfig, "SessionTimeout", "sessionTimeout", []string{"0", "-1", "foobar"})
}

func TestValidSSLSessionTimeout(t *testing.T) {
	testValidValues(t, newTestSSLConfig, "SessionTimeout", "sessionTimeout", []string{"1", "2", "10", "1ms", "2s", "10m"})
}

func TestInvalidSSLUseSessionTickets(t *testing.T) {
	testInvalidValues(t, newTestSSLConfig, "UseSessionTickets", "useSessionTickets", []string{"0", "-1", "foobar"})
}

func TestValidSSLUseSessionTickets(t *testing.T) {
	testValidValues(t, newTestSSLConfig, "UseSessionTickets", "useSessionTickets", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidSSLBufferSize(t *testing.T) {
	testInvalidValues(t, newTestSSLConfig, "BufferSize", "bufferSize", []string{"0", "-1", "foobar"})
}

func TestValidSSLBufferSize(t *testing.T) {
	testValidValues(t, newTestSSLConfig, "BufferSize", "bufferSize", []string{"1", "2", "20", "1k", "2k", "10m", "10M"})
}

func TestInvalidHSTSEnabled(t *testing.T) {
	testInvalidValues(t, newTestHSTSConfig, "Enabled", "enabled", []string{"0", "-1", "foobar"})
}

func TestValidHSTSEnabled(t *testing.T) {
	testValidValues(t, newTestHSTSConfig, "Enabled", "enabled", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidHSTSMaxAge(t *testing.T) {
	testInvalidValues(t, newTestHSTSConfig, "MaxAge", "maxAge", []string{"0", "-1", "foobar"})
}

func TestValidHSTSMaxAge(t *testing.T) {
	testValidValues(t, newTestHSTSConfig, "MaxAge", "maxAge", []string{"1", "2", "15552000"})
}

func TestInvalidHSTSIncludeSubDomains(t *testing.T) {
	testInvalidValues(t, newTestHSTSConfig, "IncludeSubDomains", "includeSubDomains", []string{"0", "-1", "foobar"})
}

func TestValidHSTSIncludeSubDomains(t *testing.T) {
	testValidValues(t, newTestHSTSConfig, "IncludeSubDomains", "includeSubDomains", []string{"true", "false", "TRUE", "FALSE"})
}

func TestInvalidHSTSPreload(t *testing.T) {
	testInvalidValues(t, newTestHSTSConfig, "Preload", "preload", []string{"0", "-1", "foobar"})
}

func TestValidHSTSPreload(t *testing.T) {
	testValidValues(t, newTestHSTSConfig, "Preload", "preload", []string{"true", "false", "TRUE", "FALSE"})
}

func testInvalidValues(t *testing.T, builder func() interface{}, fieldName string, key string, badValues []string) {
	badMap := make(map[string]string, 1)
	for _, badValue := range badValues {
		badMap[key] = badValue
		model := builder()
		err := testModeler.MapToModel(badMap, "", model)
		checkError(t, badValue, err)
	}
}

func testValidValues(t *testing.T, builder func() interface{}, fieldName string, key string, goodValues []string) {
	goodMap := make(map[string]string, 1)
	for _, goodValue := range goodValues {
		goodMap[key] = goodValue
		model := builder()
		err := testModeler.MapToModel(goodMap, "", model)
		if err != nil {
			t.Errorf("Using value \"%s\", received an unexpected error: %s", goodValue, err)
			t.FailNow()
		}
	}
}

func newTestRouterConfig() interface{} {
	return newRouterConfig()
}

func newTestGzipConfig() interface{} {
	return newGzipConfig()
}

func newTestAppConfig() interface{} {
	return newAppConfig(newRouterConfig())
}

func newTestBuilderConfig() interface{} {
	return newBuilderConfig()
}

func newTestSSLConfig() interface{} {
	return newSSLConfig()
}

func newTestHSTSConfig() interface{} {
	return newHSTSConfig()
}

func checkError(t *testing.T, value string, err error) {
	want := "modeler.ModelValidationError"
	if err == nil {
		t.Errorf("Using value \"%s\", expected a %s, but did not receive any error", value, want)
		t.FailNow()
	}
	if got := reflect.TypeOf(err).String(); want != got {
		t.Errorf("Using value \"%s\", expected a %s, but got a %s", value, want, got)
	}
}
