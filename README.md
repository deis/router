# Deis Router v2

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

When generating configuration, the program parses structured data (JSON) found in each service's `routerConfig` annotation.  This data describes all the configuration options that allow the program to dynamically construct Nginx configuration, including virtual hosts for all the domain names associated with each routable application.

Similarly, the router watches its _own_ `routerConfig` annotations to dynamically construct global Nginx configuration.

## License

Copyright 2013, 2014, 2015 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.