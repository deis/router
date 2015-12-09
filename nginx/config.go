package nginx

import (
	"os"
	"text/template"

	"github.com/deis/router/model"
)

const (
	confTemplate = `{{ $routerConfig := . }}user nginx;
daemon off;

events {
	worker_connections 1024;
}

http {
	types_hash_max_size 2048;
	server_names_hash_max_size 512;
	server_names_hash_bucket_size 64;

	log_format upstreaminfo '[$time_local] - {{ if $routerConfig.UseProxyProtocol }}$proxy_protocol_addr{{ else }}$remote_addr{{ end }} - $remote_user - $status - "$request" - $bytes_sent - "$http_referer" - "$http_user_agent" - "$server_name" - $upstream_addr - $http_host - $upstream_response_time - $request_time';

	access_log /opt/nginx/logs/access.log upstreaminfo;
	error_log  /opt/nginx/logs/error.log error;

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
		{{ if $appConfig.Available }}location / {
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
