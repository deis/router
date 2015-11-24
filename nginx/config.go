package nginx

import (
	"os"
	"text/template"

	"github.com/deis/router/model"
)

const (
	confTemplate = `events {
	worker_connections 1024;
}

http {
	types_hash_max_size 2048;
	server_names_hash_max_size 512;
	server_names_hash_bucket_size 64;

	# Default server handles requests for unmapped hostnames
	{{ $routerConfig := . }}server {
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
`
)

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
