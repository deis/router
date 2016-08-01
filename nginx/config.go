package nginx

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/deis/router/model"
)

const (
	confTemplate = `{{ $routerConfig := . }}daemon off;
pid /tmp/nginx.pid;
worker_processes {{ $routerConfig.WorkerProcesses }};

events {
	worker_connections {{ $routerConfig.MaxWorkerConnections }};
	# multi_accept on;
}

http {
	# basic settings
	sendfile on;
	tcp_nopush on;
	tcp_nodelay on;

	vhost_traffic_status_zone shared:vhost_traffic_status:{{ $routerConfig.TrafficStatusZoneSize }};

	# The timeout value must be greater than the front facing load balancers timeout value.
	# Default is the deis recommended timeout value for ELB - 1200 seconds + 100s extra.
	keepalive_timeout {{ $routerConfig.DefaultTimeout }};

	types_hash_max_size 2048;
	server_names_hash_max_size {{ $routerConfig.ServerNameHashMaxSize }};
	server_names_hash_bucket_size {{ $routerConfig.ServerNameHashBucketSize }};

	{{ $gzipConfig := $routerConfig.GzipConfig }}{{ if $gzipConfig.Enabled }}gzip on;
	gzip_comp_level {{ $gzipConfig.CompLevel }};
	gzip_disable {{ $gzipConfig.Disable }};
	gzip_http_version {{ $gzipConfig.HTTPVersion }};
	gzip_min_length {{ $gzipConfig.MinLength }};
	gzip_types {{ $gzipConfig.Types }};
	gzip_proxied {{ $gzipConfig.Proxied }};
	gzip_vary {{ $gzipConfig.Vary }};{{ end }}

	client_max_body_size {{ $routerConfig.BodySize }};

	{{ range $realIPCIDR := $routerConfig.ProxyRealIPCIDRs -}}
	set_real_ip_from {{ $realIPCIDR }};
	{{ end -}}
	real_ip_recursive on;
	{{ if $routerConfig.UseProxyProtocol -}}
	real_ip_header proxy_protocol;
	{{- else -}}
	real_ip_header X-Forwarded-For;
	{{- end }}

	log_format upstreaminfo '[$time_iso8601] - $app_name - $remote_addr - $remote_user - $status - "$request" - $bytes_sent - "$http_referer" - "$http_user_agent" - "$server_name" - $upstream_addr - $http_host - $upstream_response_time - $request_time';

	access_log /tmp/logpipe upstreaminfo;
	error_log  /tmp/logpipe {{ $routerConfig.ErrorLogLevel }};

	map $http_upgrade $connection_upgrade {
		default upgrade;
		'' close;
	}

	# The next two maps work together to determine the $access_scheme:
	# 1. Determine if SSL may have been offloaded by the load balancer, in such cases, an HTTP request should be
	# treated as if it were HTTPs.
	map $http_x_forwarded_proto $tmp_access_scheme {
		default $scheme;               # if X-Forwarded-Proto header is empty, $tmp_access_scheme will be the actual protocol used
		"~^(.*, ?)?http$" "http";      # account for the possibility of a comma-delimited X-Forwarded-Proto header value
		"~^(.*, ?)?https$" "https";    # account for the possibility of a comma-delimited X-Forwarded-Proto header value
	}
	# 2. If the request is an HTTPS request, upgrade $access_scheme to https, regardless of what the X-Forwarded-Proto
	# header might say.
	map $scheme $access_scheme {
		default $tmp_access_scheme;
		"https" "https";
	}

	# Determine the forwarded port:
	# 1. First map the unprivileged ports that Nginx (as a non-root user) actually listen on to the
	# familiar, equivalent privileged ports. (These would be the ports the k8s service listens on.)
	map $server_port $standard_server_port {
		default $server_port;
		8080 80;
		6443 443;
	}
	# 2. If the X-Forwarded-Port header has been set already (e.g. by a load balancer), use its
	# value, otherwise, the port we're forwarding for is the $standard_server_port we determined
	# above.
	map $http_x_forwarded_proto $forwarded_port {
		default $http_x_forwarded_port;
		'' $standard_server_port;
	}

	{{ $sslConfig := $routerConfig.SSLConfig }}
	{{ $hstsConfig := $sslConfig.HSTSConfig }}{{ if $hstsConfig.Enabled }}
	# HSTS instructs the browser to replace all HTTP links with HTTPS links for this domain until maxAge seconds from now.
	# The $sts variable is used later in each server block.
	map $access_scheme $sts {
		'https' 'max-age={{ $hstsConfig.MaxAge }}{{ if $hstsConfig.IncludeSubDomains }}; includeSubDomains{{ end }}{{ if $hstsConfig.Preload }}; preload{{ end }}';
	}
	{{ end }}

	{{/* Since HSTS headers are not permitted on HTTP requests, 301 redirects to HTTPS resources are also necessary. */}}
	{{/* This means we force HTTPS if HSTS is enabled. */}}
	{{ $enforceHTTPS := or $sslConfig.Enforce $hstsConfig.Enabled }}

	# Default server handles requests for unmapped hostnames, including healthchecks
	server {
		listen 8080 default_server reuseport{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		listen 6443 default_server ssl http2{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		set $app_name "router-default-vhost";
		{{ if $routerConfig.PlatformCertificate }}
		ssl_protocols {{ $sslConfig.Protocols }};
		ssl_certificate /opt/router/ssl/platform.crt;
		ssl_certificate_key /opt/router/ssl/platform.key;
		{{ else }}
		ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
		ssl_certificate /opt/router/ssl/default/default.crt;
		ssl_certificate_key /opt/router/ssl/default/default.key;
		{{ end }}
		server_name _;
		location ~ ^/healthz/?$ {
			access_log off;
			default_type 'text/plain';
			return 200;
		}
		location / {
			return 404;
		}
	}

	# Healthcheck on 9090 -- never uses proxy_protocol
	server {
		listen 9090 default_server;
		server_name _;
		set $app_name "router-healthz";
		location ~ ^/healthz/?$ {
			access_log off;
			default_type 'text/plain';
			return 200;
		}
		location ~ ^/stats/?$ {
			vhost_traffic_status_display;
			vhost_traffic_status_display_format json;
			allow 127.0.0.1;
			deny all;
		}
		location / {
			return 404;
		}
	}

	{{range $appConfig := $routerConfig.AppConfigs}}{{range $domain := $appConfig.Domains}}server {
		listen 8080{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		server_name {{ if contains "." $domain }}{{ $domain }}{{ else if ne $routerConfig.PlatformDomain "" }}{{ $domain }}.{{ $routerConfig.PlatformDomain }}{{ else }}~^{{ $domain }}\.(?<domain>.+)${{ end }};
		server_name_in_redirect off;
		port_in_redirect off;
		set $app_name "{{ $appConfig.Name }}";

		{{ if index $appConfig.Certificates $domain }}
		listen 6443 ssl http2{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		ssl_protocols {{ $sslConfig.Protocols }};
		{{ if ne $sslConfig.Ciphers "" }}ssl_ciphers {{ $sslConfig.Ciphers }};{{ end }}
		ssl_prefer_server_ciphers on;
		ssl_certificate /opt/router/ssl/{{ $domain }}.crt;
		ssl_certificate_key /opt/router/ssl/{{ $domain }}.key;
		{{ if ne $sslConfig.SessionCache "" }}ssl_session_cache {{ $sslConfig.SessionCache }};
		ssl_session_timeout {{ $sslConfig.SessionTimeout }};{{ end }}
		ssl_session_tickets {{ if $sslConfig.UseSessionTickets }}on{{ else }}off{{ end }};
		ssl_buffer_size {{ $sslConfig.BufferSize }};
		{{ if ne $sslConfig.DHParam "" }}ssl_dhparam /opt/router/ssl/dhparam.pem;{{ end }}
		{{ end }}

		{{ if or $routerConfig.EnforceWhitelists (or (ne (len $routerConfig.DefaultWhitelist) 0) (ne (len $appConfig.Whitelist) 0)) }}
		{{ if or (eq (len $appConfig.Whitelist) 0) (eq $routerConfig.WhitelistMode "extend") }}{{ range $whitelistEntry := $routerConfig.DefaultWhitelist }}allow {{ $whitelistEntry }};{{ end }}{{ end }}
		{{ range $whitelistEntry := $appConfig.Whitelist }}allow {{ $whitelistEntry }};{{ end }}
		deny all;
		{{ end }}

		vhost_traffic_status_filter_by_set_key {{ $appConfig.Name }} application::*;

		location / {
			{{ if $appConfig.Available }}proxy_buffering off;
			proxy_set_header Host $host;
			proxy_set_header X-Forwarded-For $remote_addr;
			proxy_set_header X-Forwarded-Proto $access_scheme;
			proxy_set_header X-Forwarded-Port $forwarded_port;
			proxy_redirect off;
			proxy_connect_timeout {{ $appConfig.ConnectTimeout }};
			proxy_send_timeout {{ $appConfig.TCPTimeout }};
			proxy_read_timeout {{ $appConfig.TCPTimeout }};
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection $connection_upgrade;

			{{ if $enforceHTTPS }}if ($access_scheme != "https") {
				return 301 https://$host$request_uri;
			}{{ end }}

			{{ if $hstsConfig.Enabled }}add_header Strict-Transport-Security $sts always;{{ end }}

			proxy_pass http://{{$appConfig.ServiceIP}}:80;{{ else }}return 503;{{ end }}
		}
	}

	{{end}}{{end}}
}

{{ if $routerConfig.BuilderConfig }}{{ $builderConfig := $routerConfig.BuilderConfig }}stream {
	server {
		listen 2222;
		proxy_connect_timeout {{ $builderConfig.ConnectTimeout }};
		proxy_timeout {{ $builderConfig.TCPTimeout }};
		proxy_pass {{$builderConfig.ServiceIP}}:2222;
	}
}{{ end }}
`
)

// WriteCerts writes SSL certs to file from router configuration.
func WriteCerts(routerConfig *model.RouterConfig, sslPath string) error {
	// Start by deleting all certs and their corresponding keys. This will ensure certs we no longer
	// need are deleted. Certs that are still needed will simply be re-written.
	allCertsGlob, err := filepath.Glob(filepath.Join(sslPath, "*.crt"))
	if err != nil {
		return err
	}
	allKeysGlob, err := filepath.Glob(filepath.Join(sslPath, "*.key"))
	if err != nil {
		return err
	}
	for _, cert := range allCertsGlob {
		if err := os.Remove(cert); err != nil {
			return err
		}
	}
	for _, key := range allKeysGlob {
		if err := os.Remove(key); err != nil {
			return err
		}
	}
	if routerConfig.PlatformCertificate != nil {
		err = writeCert("platform", routerConfig.PlatformCertificate, sslPath)
		if err != nil {
			return err
		}
	}
	for _, appConfig := range routerConfig.AppConfigs {
		for domain, certificate := range appConfig.Certificates {
			if certificate != nil {
				err = writeCert(domain, certificate, sslPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeCert(context string, certificate *model.Certificate, sslPath string) error {
	certPath := filepath.Join(sslPath, fmt.Sprintf("%s.crt", context))
	keyPath := filepath.Join(sslPath, fmt.Sprintf("%s.key", context))
	err := ioutil.WriteFile(certPath, []byte(certificate.Cert), 0644)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keyPath, []byte(certificate.Key), 0600)
}

// WriteDHParam writes router DHParam to file from router configuration.
func WriteDHParam(routerConfig *model.RouterConfig, sslPath string) error {
	dhParamPath := filepath.Join(sslPath, "dhparam.pem")
	if routerConfig.SSLConfig.DHParam == "" {
		err := os.RemoveAll(dhParamPath)
		if err != nil {
			return err
		}
	} else {
		err := ioutil.WriteFile(dhParamPath, []byte(routerConfig.SSLConfig.DHParam), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteConfig dynamically produces valid nginx configuration by combining a Router configuration
// object with a data-driven template.
func WriteConfig(routerConfig *model.RouterConfig, filePath string) error {
	tmpl, err := template.New("nginx").Funcs(sprig.TxtFuncMap()).Parse(confTemplate)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, routerConfig)
	if err != nil {
		return err
	}
	return nil
}
