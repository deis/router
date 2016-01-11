# Deis Router v2

[![Build Status](https://travis-ci.org/deis/router.svg?branch=master)](https://travis-ci.org/deis/router) [![Go Report Card](http://goreportcard.com/badge/deis/router)](http://goreportcard.com/report/deis/router)

Deis (pronounced DAY-iss) is an open source PaaS that makes it easy to deploy and manage applications on your own servers. Deis builds on [Kubernetes](http://kubernetes.io/) to provide a lightweight, [Heroku-inspired](http://heroku.com) workflow.

The router component, specifically, handles ingress and routing of HTTP/S traffic bound for the Deis API and for your own applications.  This component is 100% Kubernetes native and is useful even without the rest of Deis!

## Work in Progress

![Deis Graphic](https://s3-us-west-2.amazonaws.com/get-deis/deis-graphic-small.png)

Deis Router v2 is changing quickly. Your feedback and participation are more than welcome, but be aware that this project is considered a work in progress.

## Installation

This section documents simple procedures for installing the Deis Router for evaluation or use.  Those wishing to contribute to Deis Router development might consider the more developer-oriented instructions in the [Hacking Router](#hacking) section.

Deis Router can be installed with or without the rest of the Deis platform.  In either case, begin with a healthy Kubernetes cluster.  Kubernetes getting started documentation is available [here](http://kubernetes.io/gettingstarted/).

Next, install the [helm](http://helm.sh) package manager, then use the commands below to initialize that tool and load the [deis/charts](https://github.com/deis/charts) repository.

```
$ helm update
$ helm repo add deis https://github.com/deis/charts
```

To install the router:

```
$ helm install deis/<chart>
```
Where `<chart>` is selected from the options below:

| Chart | Description |
|-------|-------------|
| deis | Install the latest router release along with the rest of the latest Deis platform release. |
| deis-dev | Install the edge router (from master) with the rest of the edge Deis platform. |
| router | Install the latest router release with its minimal set of dependencies. |
| router-dev | Install the edge router (from master) with its minimal set of dependencies. |


For next steps, skip ahead to the [How it Works](#how-it-works) and [Configuration Guide](#configuration) sections.

## <a name="hacking"></a>Hacking Router

The only dependencies for hacking on / contributing to this component are:

* `git`
* `make`
* `docker`
* `kubectl`, properly configured to manipulate a healthy Kubernetes cluster that you presumably use for development
* Your favorite text editor

Although the router is written in Go, you do _not_ need Go or any other development tools installed.  Any parts of the developer workflow requiring tools not listed above are delegated to a containerized Go development environment.

### Registry

If your Kubernetes cluster is running locally on one or more virtual machines, it's advisable to also run your own local Docker registry.  This provides a place where router images built from source-- possibly containing your own experimental hacks-- can be pushed relatively quickly and can be accessed readily by your Kubernetes cluster.

Fortunately, this is very easy to set up as long as you have Docker already functioning properly:

```
$ make dev-registry
```

This will produce output containing further instructions such as:

```
59ba57a3628fe04016634760e039a3202036d5db984f6de96ea8876a7ba8a945

To use a local registry for Deis development:
    export DEIS_REGISTRY=192.168.99.102:5000/
```

Following those instructions will make your local registry usable by the various `make` targets mentioned in following sections.

If you do not want to run your own local registry or if the Kubernetes cluster you will be deploying to is remote, then you can easily make use of a public registry such as [hub.docker.com](http://hub.docker.com), provided you have an account.  To do so:

```
$ export DEIS_REGISTRY=registry.hub.docker.com/
$ export IMAGE_PREFIX=your-username
```

### If I can `make` it there, I'll `make` it anywhere...

The entire developer workflow for anyone hacking on the router is implemented as a set of `make` targets.  They are simple and easy to use, and collectively provide a workflow that should feel familiar to anyone who has hacked on Deis v1.x in the past.

#### Setup:

To "bootstrap" the development environment:

```
$ make bootstrap
```

In router's case, this step carries out some extensive dependency management using glide within the containerized development environment.  Because the router leverages the Kubernetes API, which in turn has in excess of one hundred dependencies, this step can take quite some time.  __Be patient, and allow up to 20 minutes.  You generally only ever have to do this once.__


#### To build:

```
$ make build
```

Make sure to have defined the variable `DEIS_REGISTRY` previous to this step, as your image tags will be prefixed according to this.

Built images will be tagged with the sha of the latest git commit.  __This means that for a new image to have its own unique tag, experimental changes should be committed _before_ building.  Do this in a branch.  Commits can be squashed later when you are done hacking.__

#### To deploy:

```
$ make deploy
```

The deploy target will implicitly build first, then push the built image (which has its own unique tags) to your development registry (i.e. that specified by `DEIS_REGISTRY`).  A Kubernetes manifest is prepared, referencing the uniquely tagged image, and that manifest is submitted to your Kubernetes cluster.  If a router component is already running in your Kubernetes cluster, it will be deleted and replaced with your build.

To see that the router is running, you can look for its pod(s):

```
$ kubectl get pods --namespace=deis
```

## Trying it Out

To deploy some sample routable applications:

```
$ make examples
```

This will deploy Nginx and Apache to your Kubernetes cluster as if they were user applications.

To test, first modify your `/etc/hosts` such that the following four hostnames are resolvable to the IP of the Kubernetes node that is hosting the router:

* nginx.example.com
* apache.example.com
* httpd.example.com
* unknown.example.com

By requesting the following three URLs from your browser, you should find that one is routed to a pod running Nginx, while the other two are routed to a pod running Apache:

* http://nginx.example.com
* http://apache.example.com
* http://httpd.example.com

Requesting http://unknown.example.com should result in a 404 from the router since no route exists for that domain name.

## <a name="how-it-works"></a>How it Works

The router is implemented as a simple Go program that manages Nginx and Nginx configuration.  It regularly queries the Kubernetes API for services labeled with `routable=true`.  Such services are compared to known services resident in memory.  If there are differences, new Nginx configuration is generated and Nginx is reloaded.

When generating configuration, the program reads all annotations of each service prefixed with `router.deis.io`.  These annotations describe all the configuration options that allow the program to dynamically construct Nginx configuration, including virtual hosts for all the domain names associated with each routable application.

Similarly, the router watches the annotations on its _own_ replication controller to dynamically construct global Nginx configuration.

## <a name="configuration"></a>Configuration Guide

### Environment variables

Router configuration is driven almost entirely by annotations on the router's replication controller and the services of all routable applications-- those labeled with `routable=true`.

One exception to this, however, is that in order for the router to discover its own annotations, the router must be configured via environment variable with some awareness of its own namespace.  (It cannot query the API for information about itself without knowing this.)

The `POD_NAMESPACE` environment variable is required by the router and it should be configured to match the Kubernetes namespace that the router is deployed into.  If no value is provided, the router will assume a value of `default`.

For example, consider the following Kubernetes manifest.  Given a manifest containing the following metadata:

```
apiVersion: v1
kind: ReplicationController
metadata:
  name: deis-router
  namespace: deis
# ...
```

The corresponding template must inject a `POD_NAMESPACE=deis` environment variable into router containers.  The most elegant way to achieve this is by means of the Kubernetes "downward API," as in this snippet from the same manifest:

```
# ...
spec:
  # ...
  template:
    # ...
    spec:
      containers:
      - name: deis-router
        # ...
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
# ...
```

Altering the value of the `POD_NAMESPACE` environment variable requires the router to be restarted for changes to take effect.

### Annotations

All remaining configuration options are configured through annotations.  Any of the following three Kubernetes resources can be configured:

* deis-router replication controller
* deis-builder service (if in use)
* routable applications (labeled with `routable=true`)

The table below summarizes the configuration options that are currently available for each.

_Note that Kubernetes annotation maps are all of Go type `map[string]string`.  As such, all configuration values must also be strings.  To avoid Kubernetes attempting to populate the `map[string]string` with non-string values, all numeric and boolean configuration values should be enclosed in double quotes to help avoid confusion._


| Component | Annotation | Default Value | Description |
|-----------|------------|---------------|-------------|
| deis-router | router.deis.io/workerProcesses | `"auto"` (number of CPU cores) | Number of worker processes to start. |
| deis-router | router.deis.io/workerConnections| `"768"` | Maximum number of simultaneous connections that can be opened by a worker process. |
| deis-router | router.deis.io/defaultTimeout | `"1300"` | Default timeout value in seconds.  Should be greater than the front-facing load balancer's timeout value. |
| deis-router | router.deis.io/serverNameHashMaxSize | `"512"` | nginx `server_names_hash_max_size` setting. |
| deis-router | router.deis.io/serverNameHashBucketSize | `"64"` | nginx `server_names_hash_bucket_size` setting. |
| deis-router | router.deis.io/gzipConfig.compLevel | `"5"` | nginx `gzip_comp_level` setting. |
| deis-router | router.deis.io/gzipConfig.disable | `"msie6"` | nginx `gzip_disable` setting. |
| deis-router | router.deis.io/gzipConfig.httpVersion | `"1.1"` | nginx `gzip_http_version` setting. |
| deis-router | router.deis.io/gzipConfig.minLength | `"256"` | nginx `gzip_min_length` setting. |
| deis-router | router.deis.io/gzipConfig.proxied | `"any"` | nginx `gzip_proxied` setting. |
| deis-router | router.deis.io/gzipConfig.types | `"application/atom+xml application/javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"` | nginx `gzip_types` setting. |
| deis-router | router.deis.io/gzipConfig.vary | `"on"` | nginx `gzip_vary` setting. |
| deis-router | router.deis.io/bodySize | `"1"`| nginx `client_max_body_size` setting (in megabytes). |
| deis-router | router.deis.io/proxyRealIpCidr | `"10.0.0.0/8"` | nginx `set_real_ip_from` setting.  Defines trusted addresses that are known to send correct replacement addresses. |
| deis-router | router.deis.io/errorLogLevel | `"error"` | Log level used in the nginx `error_log` setting (valid values are: `debug`, `info`, `notice`, `warn`, `error`, `crit`, `alert`, and `emerg`). |
| deis-router | router.deis.io/domain | N/A | This defines the router's default domain.  Any domains added to a routable application _not_ containing the `.` character will be assumed to be subdomains of this default domain.  Thus, for example, a default domain of `example.com` coupled with a routable app counting `foo` among its domains will result in router configuration that routes traffic for `foo.example.com` to that application. |
| deis-router | router.deis.io/useProxyProtocol | `"false"` | PROXY is a simple protocol supported by nginx, HAProxy, Amazon ELB, and others.  It provides a method to obtain information about a request's originating IP address from an external (to Kubernetes) load balancer in front of the router.  Enabling this option allows the router to select the originating IP from the HTTP `X-Forwarded-For` header. |
| deis-router | router.deis.io/enforceWhitelists | `"false"` | Whether to honor application-level IP / CIDR whitelists. |
| deis-builder | router.deis.io/connectTimeout | `"10"` | nginx `proxy_connect_timeout` setting (in seconds). |
| deis-builder | router.deis.io/tcpTimeout | `"1200"` | nginx `proxy_timeout` setting (in seconds). |
| routable application | router.deis.io/domains | N/A | Comma-delimited list of domains for which traffic should be routed to the application.  These may be fully qualified (e.g. `foo.example.com`) or, if not containing any `.` character, will be considered subdomains of the router's domain, if that is defined. |
| routable application | router.deis.io/whitelist | N/A | Comma-delimited list of addresses permitted to access the application (using IP or CIDR notation).  Requests from all other addresses are denied. |
| routable application | router.deis.io/connectTimeout | `"30"` | nginx `proxy_connect_timeout` setting (in seconds). |
| routable application | router.deis.io/tcpTimeout | router's `defaultTimeout` | nginx `proxy_send_timeout` and `proxy_read_timeout` settings (in seconds). |

#### Annotations by example

##### router replication controller:

```
apiVersion: v1
kind: ReplicationController
metadata:
  name: deis-router
  namespace: deis
  # ...
  annotations:
    router.deis.io/domain: example.com
    router.deis.io/useProxyProtocol: "true"
# ...
```

##### builder service:

```
apiVersion: v1
kind: Service
metadata:
  name: deis-builder
  namespace: deis
  # ...
  annotations:
    router.deis.io/connectTimeout: "20000"
    router.deis.io/tcpTimeout: "2400000"
# ...
```

##### routable service:

```
apiVersion: v1
kind: Service
metadata:
  name: foo
  labels:
  	routable: "true"
  namespace: examples
  # ...
  annotations:
    router.deis.io/domains: foo,bar,www.foobar.com
# ...
```

### SSL

Router currently has limited support for HTTPS with the ability to perform SSL termination using a wildcard certificate supplied as a Kubernetes secret.  Just as router utilizes the Kubernetes API to discover routable services, router is also able to discover the secret bearing its own certificate.  This allows the router to dynamically refresh and reload configuration whenever such a certificate is added, updated, or removed.  There is never a need to explicitly restart the router.

A wildcard certificate supplied in the manner described above will be used to provide a secure virtual host (in addition to the insecure virtual host) for every "domain" of a routable service _not_ containing the `.` character.

For instance, if a routable service exists having a "domain" `frozen-wookie` and the platform domain is `example.com`, a supplied wildcard certificate for `*.example.com` will be used to secure a `frozen-wookie.example.com` virtual host.  Similarly, if no platform domain is defined, the supplied wildcard certificate will be used to secure a virtual host matching the expression `~^frozen-wookie\.(?<domain>.+)$`.  (The latter is almost certainly guaranteed to result in certificate warnings in an end user's browser, so it is advisable to always define the platform domain.)

If the same routable service also had a domain `www.frozen-wookie.com`, the `*.example.com` wildcard certificate plays no role in securing a `www.frozen-wookie.com` virtual host.

#### SSL example

Here is an example of a Kubernetes secret bearing a wildcard certificate for use by the router.  The following criteria must be met:

* Namespace must be the same namespace as the router
* Name must be `deis-router-cert`
* Certificate must be supplied as the value of the key `cert`
* Certificate private key must be supplied as the value of the key `key`
* Both the certificate and private key must be base64 encoded

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: deis-router-cert
  namespace: deis
  labels:
    heritage: deis
type: Opaque
data:
  cert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURsakNDQW40Q0NRREQrYlVXY2dzT05UQU5CZ2txaGtpRzl3MEJBUVVGQURDQmpERUxNQWtHQTFVRUJoTUMKVlZNeEN6QUpCZ05WQkFnVEFrTlVNUk13RVFZRFZRUUhFd3BOYVdSa2JHVjBiM2R1TVEwd0N3WURWUVFLRXdSRQpaV2x6TVJRd0VnWURWUVFMRXd0RmJtZHBibVZsY21sdVp6RVlNQllHQTFVRUF4UVBLaTV0YVdOeWJ5MXJkV0psCkxtbHZNUnd3R2dZSktvWklodmNOQVFrQkZnMXJaVzUwUUdSbGFYTXVZMjl0TUI0WERURTJNREV3TlRFNE1ESXcKT0ZvWERURTNNREV3TkRFNE1ESXdPRm93Z1l3eEN6QUpCZ05WQkFZVEFsVlRNUXN3Q1FZRFZRUUlFd0pEVkRFVApNQkVHQTFVRUJ4TUtUV2xrWkd4bGRHOTNiakVOTUFzR0ExVUVDaE1FUkdWcGN6RVVNQklHQTFVRUN4TUxSVzVuCmFXNWxaWEpwYm1jeEdEQVdCZ05WQkFNVUR5b3ViV2xqY204dGEzVmlaUzVwYnpFY01Cb0dDU3FHU0liM0RRRUoKQVJZTmEyVnVkRUJrWldsekxtTnZiVENDQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQgpBT2RKWTFVeC9KRzRuWlJMT3prOHZXSUhpNEN6aURWUWV1cjNJdnFBV0RjRTRIRmRKSkZnNXZxTDN6eHFwTTArCkZQRmFHU2VTbmYvK00xdXN6YTNjRmUrblBTNXdQM1JJSTlSaEpEb21wWlc4Y3VFSDl0b3lLSVM0OEZwS1VmNGEKTWY4RzFQZmV3RXhjOCtJYVZGTlJzb3RVRmxmRXc3MkJwQmF0WGllL2M2Q294QVRMQ1BuSksvemNqMFhzOVJqRgoxeHA5aFNXWDJoZEROVGxEdmt3OEJkVHhVZXlSNWR5bUVmVWRxUlc5YVdBNDZRN2NxdC9xZU9la29hS3lsUVF4ClBQWml3TXZqUjFzcWpyc1RYZ0dMM1ZuU1ZzZC84MWtEdUNtckUvMjRFY0ZjSVhRdC9DTHhnN0tCbUVNSU5wOTkKQWhwWWxzTm0vWlJzUGJJbVNGaWdZQTBDQXdFQUFUQU5CZ2txaGtpRzl3MEJBUVVGQUFPQ0FRRUFYUHgvOGsxZgpMUWVCeFN5SE9QdUJtN3lOSFhIdnBKelQwNFhlQUlMd3YwZ3MreUc2RGUrWFREemkwa2xHbkdsajJIK0xlN3NvClYyTkdUZ3Z1MVZKbEgrelBjUkhYNjE2cHdNK1diNnlKZzUxMmE5L2FYZUM2MUZQcjRuemRRNFJicDduUldEdU8KM1ZqS0ppUStTUnJTS2o2cytyVWRRZjR6SjAzTk05ajYwV1hqVm53RWtlM1BqL1ZKdHNoUWlsMEVwUjZsZVdJUwpQckpKV3lLZmIvbTZYR3AxTTVWc05vSkpaNHNuakkwSlNvcTlNUU5SSEVTaWUrc2pBZU4zR3NFcGlidWhVWE1FCm44M3ZUbkppTW8ycVlLWkJEa252Nkw2aGJTcUJEcG1oS25sNWlLWWV6RDk2d1ZJRTR1QlhGcWdhS3NGZ0tmNlMKOHJwVXg2dmNHOVpraUE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
  key: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBNTBsalZUSDhrYmlkbEVzN09UeTlZZ2VMZ0xPSU5WQjY2dmNpK29CWU53VGdjVjBrCmtXRG0rb3ZmUEdxa3pUNFU4Vm9aSjVLZC8vNHpXNnpOcmR3Vjc2YzlMbkEvZEVnajFHRWtPaWFsbGJ4eTRRZjIKMmpJb2hMandXa3BSL2hveC93YlU5OTdBVEZ6ejRocFVVMUd5aTFRV1Y4VER2WUdrRnExZUo3OXpvS2pFQk1zSQorY2tyL055UFJlejFHTVhYR24yRkpaZmFGME0xT1VPK1REd0YxUEZSN0pIbDNLWVI5UjJwRmIxcFlEanBEdHlxCjMrcDQ1NlNob3JLVkJERTg5bUxBeStOSFd5cU91eE5lQVl2ZFdkSld4My96V1FPNEthc1QvYmdSd1Z3aGRDMzgKSXZHRHNvR1lRd2cybjMwQ0dsaVd3MmI5bEd3OXNpWklXS0JnRFFJREFRQUJBb0lCQVFDN0hPZzcySURZODE2bwp4bVJYWGdobzhtbXQ2WkRjY1F2QVk4ZnNtVURFNGpFaUJaUzcrSVZsVlJ0NUtWM0NuZU1ianlTaW16OHdCMXROClkzUFNvQ2N2QjBHNy9CdVoyM01CTXNXNE5lUG9TOWswck5qTWxaK3FJR1J5NTRXRjZYTmhPd200cm9Fc1JsY1kKS096bnEwU2k2NWhYMFZ2NlNKT1R0UlpqV2NRVnlHa24vdU5TZWxOWUFCMWdpeGVVWU05QkNkTTVVcFhLbTVudgpqTStUUGhoL3YzbVdZR1MwRWxkeEsxcU43TW56Y21MQTN2ZVVUUzVEQUlENlh2ZEtudEpxNGg3cFhOMHgwVkd0Ck1ETmNLNWxuMHViTm54WkRQQmNNZDIvQU5RZTFkU3N0K09SZjc3OHZZd3RkalUzMDlOU2ZXbzdoYllHaGhMMG8KaGpSWVIySUJBb0dCQVAwSnl6OUR5WUlMU29XZ0F5WnBycUpLZHpQM3gxWnZIYW9pcUVQTTRBOExNUzFZTW5HNgpZQ1NPTU9rNWxtNWxnanFBeFR1V2dic0hEYUdmenBnWFFnanBZbVAxOGFic2JPem0rSy90UjhOVjBHZzZVdEFrCkIxbG44elpFOHdoTHphZ21IM3JYZTk2WFUxZE9hbFhaZkpuSU13RkYvS2ZPY00wdjhHNVFHQmZ0QW9HQkFPbisKYXQxMlJvNEdRZjc3NEJLaG02UDQxQzJXbnJXSnFjeWZmT2pWeDJmMHd3T2ZNcnBrM3Q5L3E5R2tPVUJrREREMgprYS9SQzBDTkpHN0NtS2hPT2JXc3JydzZPRy9Uc3cwQmxBLy9KR0dkQzhUZTJBWU5QeVVrVnhEaWhZamdvRXdPCnZ5eGVsRXkyYmsyZlRrOFV2NXByVEFISDFBck5ubTZtQUZSekRDU2hBb0dBWmc3S05FVWlCRnFMZnlQUHdYR0UKcHg2RHY2a2pkRUF5NWNrK3RzSHVhUUxwbmRGaDN5NzkwaUNycDVXN2orRlBTa3Rmd2EzdHUwS1lBK1ZSelZRcwpKUzRFcWk2cmZka0VBZlVTdnRwa1JHdmJOLzYySWc4MkM2QjZueHZtTGlWczJISXY1anI5VWdVVTd6VElmTUhRCm1MSC9xSGJTOWRlMDlYMXRhd0RSQnUwQ2dZQUNJcFlyQ3ljOXFJVDFwS3E0d016QXhFWjdYQUo2dVpUWkR5TisKcEdTN2FCcFhjQjV3WVBjL3B0U1VoeUVXbTFNWkowYlJDZUZHWjROTGwyNzVLaTJRb05DL1d2Q2s1Sk9mck1JRApHMXJqb1NFNFhLZ3Y5Zkw3amNkVm1odUowakxrUHNZenkzRWkrZzhybWlhR3hqT3hYd1ZFaHVSQ2JQbmpISlNvCk1KeU53UUtCZ1FDWUY4VDQxVWw1bW96b1l5MitUQlYrMnRtb255RTNZWndxVHFRRkI5NkliM0owRWM5a2VwRWMKaGt4RDZXcFJ0TENjdnFDTEc0L3JTS2ZSWDJGY2VGMWJ4aUEwMi9wNmxUVk5iM3hnK0VNZE1hblNXQnFsT05UTwpuWjdrRDZLQ2d0TVdWelBDZnNtd2MxZHFaTlczRGxxT21PS0pDbk5xeTU1L25PUjd5WUtFd0E9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
```

## License

Copyright 2013, 2014, 2015 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.