package nginx

import (
	"os"
	"text/template"

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
	vhost_traffic_status_zone;

	sendfile on;
	tcp_nopush on;
	tcp_nodelay on;

	# The timeout value must be greater than the front facing load balancers timeout value.
	# Default is the deis recommended timeout value for ELB - 1200 seconds + 100s extra.
	keepalive_timeout {{ $routerConfig.DefaultTimeout }}s;

	types_hash_max_size 2048;
	server_names_hash_max_size {{ $routerConfig.ServerNameHashMaxSize }};
	server_names_hash_bucket_size {{ $routerConfig.ServerNameHashBucketSize }};

	{{ if $routerConfig.GzipConfig }}{{ $gzipConfig := $routerConfig.GzipConfig }}gzip on;
	gzip_comp_level {{ $gzipConfig.CompLevel }};
	gzip_disable {{ $gzipConfig.Disable }};
	gzip_http_version {{ $gzipConfig.HTTPVersion }};
	gzip_min_length {{ $gzipConfig.MinLength }};
	gzip_types {{ $gzipConfig.Types }};
	gzip_proxied {{ $gzipConfig.Proxied }};
	gzip_vary {{ $gzipConfig.Vary }};{{ end }}

	client_max_body_size {{ $routerConfig.BodySize }}m;

	{{ if $routerConfig.UseProxyProtocol }}set_real_ip_from {{ $routerConfig.ProxyRealIPCIDR }};
	real_ip_header proxy_protocol;
	{{ end }}

	log_format upstreaminfo '[$time_local] - {{ if $routerConfig.UseProxyProtocol }}$proxy_protocol_addr{{ else }}$remote_addr{{ end }} - $remote_user - $status - "$request" - $bytes_sent - "$http_referer" - "$http_user_agent" - "$server_name" - $upstream_addr - $http_host - $upstream_response_time - $request_time';

	access_log /opt/nginx/logs/access.log upstreaminfo;
	error_log  /opt/nginx/logs/error.log error;

	map $http_upgrade $connection_upgrade {
		default upgrade;
		'' close;
	}

	# Default server handles requests for unmapped hostnames
	server {
		listen 80{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		server_name _;
		location / {
			return 404;
		}
	}

	{{range $appConfig := $routerConfig.AppConfigs}}{{range $domain := $appConfig.Domains}}server {
		listen 80{{ if $routerConfig.UseProxyProtocol }} proxy_protocol{{ end }};
		server_name {{$domain}};
		server_name_in_redirect off;
		port_in_redirect off;
		{{ if $appConfig.Available }}location / {
			proxy_buffering off;
			proxy_set_header Host $host;
			proxy_set_header X-Forwarded-For {{ if $routerConfig.UseProxyProtocol }}$proxy_protocol_addr{{ else }}$proxy_add_x_forwarded_for{{ end }};
			proxy_redirect off;
			proxy_connect_timeout 30s;
			proxy_send_timeout {{ $routerConfig.DefaultTimeout }}s;
			proxy_read_timeout {{ $routerConfig.DefaultTimeout }}s;
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection $connection_upgrade;
			proxy_pass http://{{$appConfig.ServiceIP}}:80;
		}{{ else }}location / {
			return 503;
		}{{ end }}
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

// WriteConfig dynamically produces valid nginx configuration by combining a Router configuration
// object with a data-driven template.
func WriteConfig(routerConfig *model.RouterConfig, filePath string) error {
	tmpl, err := template.New("nginx").Parse(confTemplate)
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
