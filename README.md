# Deis Router v2

[![Build Status](https://travis-ci.org/deis/router.svg?branch=master)](https://travis-ci.org/deis/router) [![Go Report Card](http://goreportcard.com/badge/deis/router)](http://goreportcard.com/report/deis/router)

Deis (pronounced DAY-iss) is an open source PaaS that makes it easy to deploy and manage applications on your own servers. Deis builds on [Kubernetes](http://kubernetes.io/) to provide a lightweight, [Heroku-inspired](http://heroku.com) workflow.

The router component, specifically, handles ingress and routing of HTTP/S traffic bound for the Deis API and for your own applications.  This component is 100% Kubernetes native and is useful even without the rest of Deis!

## Work in Progress

![Deis Graphic](https://s3-us-west-2.amazonaws.com/get-deis/deis-graphic-small.png)

Deis Router v2 is changing quickly. Your feedback and participation are more than welcome, but be aware that this project is considered a work in progress.

## Hacking Router

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
    export DEIS_REGISTRY=192.168.99.102:5000
```

Following those instructions will make your local registry usable by the various `make` targets mentioned in following sections.

If you do not want to run your own local registry or if the Kubernetes cluster you will be deploying to is remote, then you can easily make use of a public registry such as [hub.docker.com](http://hub.docker.com), provided you have an account.  To do so:

```
$ export DEIS_REGISTRY=registry.hub.docker.com
$ export IMAGE_PREFIX=your-username/
```

__Do not miss the trailing slash in the `IMAGE_PREFIX`!__


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

## How it Works

The router is implemented as a simple Go program that manages Nginx and Nginx configuration.  It regularly queries the Kubernetes API for services labeled with `routable=true`.  Such services are compared to known services resident in memory.  If there are differences, new Nginx configuration is generated and Nginx is reloaded.

When generating configuration, the program parses structured data (JSON) found in each service's `deis.io/routerConfig` annotation.  This data describes all the configuration options that allow the program to dynamically construct Nginx configuration, including virtual hosts for all the domain names associated with each routable application.

Similarly, the router watches its _own_ `deis.io/routerConfig` annotations to dynamically construct global Nginx configuration.

## Configuration Guide

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

All remaining configuration options are configured through annotations.  Any of the following three Kubernetes resources can be configured through a JSON object provided as a value of the `deis.io/routerConfig` annotation:

* deis-router replication controller
* deis-builder service (if in use)
* routable applications (labeled with `routable=true`)

Note that although the annotation containing router configuration for each of the above is consistently named `deis.io/routerConfig`, the structure of the JSON object used for each of those differs by use case.  The table below summarizes the configuration options that are currently available for each:

| Component            | Name             | Type       | Default Value | Description |
|----------------------|------------------|------------|---------------|-------------|
| deis-router          | workerProcesses  | `string`   | `auto` (number of CPU cores) | Number of worker processes to start. |
| deis-router          | workerConnections | `integer` | `768`         | Maximum number of simultaneous connections that can be opened by a worker process. |
| deis-router          | defaultTimeout   | `integer`  | 1300          | Default timeout value in seconds.  Should be greater than the front-facing load balancer's timeout value. |
| deis-router          | serverNameHashMaxSize | `integer` | `512`     | nginx `server_names_hash_max_size` setting. |
| deis-router          | serverNameHashBucketSize | `integer` | 64     | nginx `server_names_hash_bucket_size` setting. |
| deis-router          | gzipConfig             | `GzipConfig`  | Described by following lines.        | Set to `null` to disable gzip entirely. |
| deis-router          | gzipConfig.compLevel   | `integer` | `5`        | nginx `gzip_comp_level` setting. |
| deis-router          | gzipConfig.disable     | `string`  | `msie6`    | nginx `gzip_disable` setting. |
| deis-router          | gzipConfig.httpVersion | `string`  | `1.1`      | nginx `gzip_http_version` setting. |
| deis-router          | gzipConfig.minLength   | `integer` | `256`      | nginx `gzip_min_length` setting. |
| deis-router          | gzipConfig.proxied     | `string`  | `any`      | nginx `gzip_proxied` setting. |
| deis-router          | gzipConfig.types       | `string`  | `application/atom+xml application/javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component` | nginx `gzip_types` setting. |
| deis-router          | gzipConfig.vary        | `string`  | `on`       | nginx `gzip_vary` setting. |
| deis-router          | bodySize         | `integer` | `1`            | nginx `client_max_body_size` setting (in megabytes). |
| deis-router          | proxyRealIpCidr  | `string`   | `10.0.0.0/8`  | nginx `set_real_ip_from` setting.  Defines trusted addresses that are known to send correct replacement addresses. |
| deis-router          | domain           | `string`   | N/A           | This defines the router's default domain.  Any domains added to a routable application _not_ containing the `.` character will be assumed to be subdomains of this default domain.  Thus, for example, a default domain of `example.com` coupled with a routable app counting `foo` among its domains will result in router configuration that routes traffic for `foo.example.com` to that application. |
| deis-router          | useProxyProtocol | `boolean`  | `false`       | PROXY is a simple protocol supported by nginx, HAProxy, Amazon ELB, and others.  It provides a method to obtain information about a request's originating IP address from an external (to Kubernetes) load balancer in front of the router.  Enabling this option allows the router to select the originating IP from the HTTP `X-Forwarded-For` header. |
| deis-builder         | connectTimeout   | `integer`  | `10`       | nginx `proxy_connect_timeout` setting (in seconds). |
| deis-builder         | tcpTimeout       | `integer`  | `1200`     | nginx `proxy_timeout` setting (in seconds). |
| routable application | domains          | `[]string` | N/A           | List of domains for which traffic should be routed to the application.  These may be fully qualified (e.g. `foo.example.com`) or, if not containing any `.` character, may be relative to the router's default domain. |
| routable application | connectTimeout   | `integer`  | `30`       | nginx `proxy_connect_timeout` setting (in seconds). |
| routable application | tcpTimeout       | `integer`  | router's `defaultTimeout` | nginx `proxy_send_timeout` and `proxy_read_timeout` settings (in seconds). |

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
    deis.io/routerConfig: |
      {
        "domain": "example.com",
        "useProxyProtocol": true
      }
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
    deis.io/routerConfig: |
      {
        "connectTimeout": 20000,
        "tcpTimeout": 2400000
      }
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
    deis.io/routerConfig: |
      {
        "domains": ["foo", "bar", "www.foobar.com"]
      }
# ...
```

## License

Copyright 2013, 2014, 2015 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.