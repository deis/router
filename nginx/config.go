package nginx

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/deis/router/model"
)

const (
	confTemplate = `{{ $routerConfig := . }}user nginx;
daemon off;
pid /run/nginx.pid;
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

	set_real_ip_from {{ $routerConfig.ProxyRealIPCIDR }};
	{{ if $routerConfig.UseProxyProtocol }}
	real_ip_header proxy_protocol;{{ else }}real_ip_header X-Forwarded-For;
	{{ end }}

	log_format upstreaminfo '[$time_local] - $remote_addr - $remote_user - $status - "$request" - $bytes_sent - "$http_referer" - "$http_user_agent" - "$server_name" - $upstream_addr - $http_host - $upstream_response_time - $request_time';

	access_log /opt/nginx/logs/access.log upstreaminfo;
	error_log  /opt/nginx/logs/error.log {{ $routerConfig.ErrorLogLevel }};

	map $http_upgrade $connection_upgrade {
		default upgrade;
		'' close;
	}

	# Trust http_x_forwarded_proto headers correctly indicate ssl offloading.
	map $http_x_forwarded_proto $access_scheme {
		default $http_x_forwarded_proto;
		'' $scheme;
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
		listen 80 default_server reuseport{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		{{ if $routerConfig.DefaultCertificate }}
		listen 443 default_server ssl{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		ssl_protocols {{ $sslConfig.Protocols }};
		ssl_certificate /opt/nginx/ssl/default.crt;
		ssl_certificate_key /opt/nginx/ssl/default.key;
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
		location ~ ^/healthz/?$ {
			access_log off;
			default_type 'text/plain';
			return 200;
		}
		location / {
			return 404;
		}
	}

	{{range $appConfig := $routerConfig.AppConfigs}}{{range $domain := $appConfig.Domains}}server {
		listen 80{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		server_name {{ if contains "." $domain }}{{ $domain }}{{ else if ne $routerConfig.DefaultDomain "" }}{{ $domain }}.{{ $routerConfig.DefaultDomain }}{{ else }}~^{{ $domain }}\.(?<domain>.+)${{ end }};
		server_name_in_redirect off;
		port_in_redirect off;

		{{ if index $appConfig.Certificates $domain }}
		listen 443 ssl{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		ssl_protocols {{ $sslConfig.Protocols }};
		{{ if ne $sslConfig.Ciphers "" }}ssl_ciphers {{ $sslConfig.Ciphers }};{{ end }}
		ssl_prefer_server_ciphers on;
		ssl_certificate /opt/nginx/ssl/{{ $domain }}.crt;
		ssl_certificate_key /opt/nginx/ssl/{{ $domain }}.key;
		{{ if ne $sslConfig.SessionCache "" }}ssl_session_cache {{ $sslConfig.SessionCache }};
		ssl_session_timeout {{ $sslConfig.SessionTimeout }};{{ end }}
		ssl_session_tickets {{ if $sslConfig.UseSessionTickets }}on{{ else }}off{{ end }};
		ssl_buffer_size {{ $sslConfig.BufferSize }};
		{{ if ne $sslConfig.DHParam "" }}ssl_dhparam /opt/nginx/ssl/dhparam.pem;{{ end }}
		{{ end }}

		{{ if or $routerConfig.EnforceWhitelists (or (ne (len $routerConfig.DefaultWhitelist) 0) (ne (len $appConfig.Whitelist) 0)) }}
		{{ if or (eq (len $appConfig.Whitelist) 0) (eq $routerConfig.WhitelistMode "extend") }}{{ range $whitelistEntry := $routerConfig.DefaultWhitelist }}allow {{ $whitelistEntry }};{{ end }}{{ end }}
		{{ range $whitelistEntry := $appConfig.Whitelist }}allow {{ $whitelistEntry }};{{ end }}
		deny all;
		{{ end }}

		location / {
			proxy_buffering off;
			proxy_set_header Host $host;
			proxy_set_header X-Forwarded-For $remote_addr;
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

			proxy_pass http://{{$appConfig.ServiceIP}}:80;
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

func WriteCerts(routerConfig *model.RouterConfig, sslPath string) error {
	if routerConfig.DefaultCertificate != nil {
		err := writeCert("default", routerConfig.DefaultCertificate, sslPath)
		if err != nil {
			return err
		}
	}
	for _, appConfig := range routerConfig.AppConfigs {
		for domain, certificate := range appConfig.Certificates {
			if certificate != nil {
				err := writeCert(domain, certificate, sslPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeCert(context string, certificate *model.Certificate, sslPath string) error {
	certPath := path.Join(sslPath, fmt.Sprintf("%s.crt", context))
	keyPath := path.Join(sslPath, fmt.Sprintf("%s.key", context))
	err := ioutil.WriteFile(certPath, []byte(certificate.Cert), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(keyPath, []byte(certificate.Key), 0600)
	if err != nil {
		return err
	}
	return nil
}

func WriteDHParam(routerConfig *model.RouterConfig, sslPath string) error {
	if routerConfig.SSLConfig.DHParam != "" {
		dhParamPath := path.Join(sslPath, "dhparam.pem")
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
