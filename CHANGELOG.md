### v2.0.0-beta1 -> v2.0.0-beta2

#### Features

 - [`0c743f0`](https://github.com/deis/router/commit/0c743f0a08be1e837b1ec3e90b0ffbcd325c2014) certs: Add support for format of ingress TLS cert secrets
 - [`e674c68`](https://github.com/deis/router/commit/e674c68a0262ff094d19487dac0244885f285f7d) README.md: add quay.io container badge
 - [`a9568f5`](https://github.com/deis/router/commit/a9568f549135c78b1e978df8ec2d0843db4c82b1) _scripts: add CHANGELOG.md and generator script
 - [`431f7ac`](https://github.com/deis/router/commit/431f7ac3fbdfc1a17607b359af9212534b494a2d) router: Respond with 503 for upstream svcs w/ no endpoints

#### Maintenance

 - [`5b9df61`](https://github.com/deis/router/commit/5b9df61b4cccc82fbc826ac392198bf69f2c7fc9) Makefile: update go-dev to 0.10.0 and compress router binary

### 2.0.0-alpha -> v2.0.0-beta1

#### Features

 - [`f291e35`](https://github.com/deis/router/commit/f291e35dfc4b992c0f0946c40b2a0ff115815d92) .travis.yml: have this job notify its sister job in Jenkins
 - [`931e9ab`](https://github.com/deis/router/commit/931e9ab63d2b9e0cd63b38d3bb09164525c7a4e6) router: Group stats by app
 - [`80e23b0`](https://github.com/deis/router/commit/80e23b009f9fe266f7ba183bd533d5db61f5a342) travis: add travis webhook -> e2e tests
 - [`76e9ff4`](https://github.com/deis/router/commit/76e9ff4d551122d78eccebd9d8e41a8c61e4f1b7) router: Enable vts module
 - [`769140b`](https://github.com/deis/router/commit/769140b1fcc9deede7be518f2a24633b146246f5) router: Make parsing of boolean config values case-insensitive
 - [`a227568`](https://github.com/deis/router/commit/a227568a0cea932fb971595de6fb9a3442bff4f7) router: Make default vhost ssl protocols configurable
 - [`ad8788c`](https://github.com/deis/router/commit/ad8788c596c3badfc24b6232d0759e32cf148341) router: Implement various ssl-related options
 - [`06de12d`](https://github.com/deis/router/commit/06de12d82232e7f92a6d4e5c2d289fe624545c49) router: Add support for forcing HTTPS and using HSTS
 - [`ef40bc6`](https://github.com/deis/router/commit/ef40bc625f1e5c4021107be4691d801dd51b681d) router: add ssl support for custom domains
 - [`754f8cb`](https://github.com/deis/router/commit/754f8cb38490c2fe4b2b12f55ebe12fc42680ee5) manifests: Add liveness and readiness probes
 - [`d79bf4d`](https://github.com/deis/router/commit/d79bf4dceee854debb0714deb9d088c22160885d) router: add initial ssl support

#### Fixes

 - [`bba285e`](https://github.com/deis/router/commit/bba285eb252f2b3d6a3e56ec6d97e4b5d19eeccf) ssl: Don't ever use the default/platform-wide wildcard cert for "custom" (fully qualified
 - [`f2c0dfb`](https://github.com/deis/router/commit/f2c0dfb8bb81e14a8413701697e832f2f078fbc0) ssl: Add default default vhost with self-signed cert
 - [`a52293f`](https://github.com/deis/router/commit/a52293f363189891ede28c0be69b12e5ff409278) router: Only use unprivileged ports
 - [`90611f3`](https://github.com/deis/router/commit/90611f3f76684b66aeacbb9860208c9039abcc00) router: Fix log permissions for non-root user
 - [`fb1bee0`](https://github.com/deis/router/commit/fb1bee0cdad14f6a300c0e4a44ad8afbd82e12f2) examples: Make examples routable again
 - [`24b6ebf`](https://github.com/deis/router/commit/24b6ebf4b7d36ac189b70fc517750555c3825f0a) router: Correct and improve whitelist behavior
 - [`ee3d768`](https://github.com/deis/router/commit/ee3d76891e17aef543f2d64e164dfac2a354ec23) router: Fix issues establishing real end-client IP
 - [`bb1dfc4`](https://github.com/deis/router/commit/bb1dfc48d5d777479ffc1d0158409897b003f7e7) router: Make gzip disableable again

#### Maintenance

 - [`f8ebbb1`](https://github.com/deis/router/commit/f8ebbb182061d365ef632b9e17ee9b97917e3588) Makefile: Enable immutable (git-based
 - [`49af1ad`](https://github.com/deis/router/commit/49af1ad345c2208a7a989f41c7841d6208319a43) build/test: Upgrade to latest dev environment
 - [`8dbc15f`](https://github.com/deis/router/commit/8dbc15ffcfd8340effb42a4404a3293cf75985c0) ci: Always run docker build during CI
 - [`72667e1`](https://github.com/deis/router/commit/72667e1eeb169ebfcfe3524fbc9073f780b2b017) Makefile: Bump docker-go-dev version
 - [`4d3c319`](https://github.com/deis/router/commit/4d3c319a729d6ecc9d17d68440e8026b4dad4900) router: Make optional gzip config a little cleaner
 - [`8d2c8ec`](https://github.com/deis/router/commit/8d2c8ec5cf9351cd03632b2bd5601ec9bf28a291) image: bump alpine to 3.3
 - [`1e4f998`](https://github.com/deis/router/commit/1e4f998f0bebf58737b9fa01d9b634df4c6f9419) release: bump version to v2-beta

### 2.0.0-alpha

#### Features

 - [`b044436`](https://github.com/deis/router/commit/b04443627770f620b019ce6c1ae7c240d1e54ae8) router: make platform domain optional
 - [`2d18ae0`](https://github.com/deis/router/commit/2d18ae0fe161dc99594de7401f06e078a0cc0789) Makefile/travis: use standard vars
 - [`ef6bd0a`](https://github.com/deis/router/commit/ef6bd0ab0b1c38c85821f06855715ce9d90d37e7) router: honor application ip/cidr whitelists
 - [`1ed1b41`](https://github.com/deis/router/commit/1ed1b41a84cd00eb8c7dc91573dc3adc322cf4ec) router: make log level configurable
 - [`3aa775f`](https://github.com/deis/router/commit/3aa775fc96bcffa89815a85b46d598b141fcc5fe) router: make timeouts configurable at app level
 - [`187adc7`](https://github.com/deis/router/commit/187adc7a2396fbe990450c4579979dd9c62490ba) router: add router healthchecks
 - [`abb4ccd`](https://github.com/deis/router/commit/abb4ccd70dafd0afab52b4fbf0a7c5fd62cbd82f) router: support configurable client_max_body_size
 - [`fa6d289`](https://github.com/deis/router/commit/fa6d2899c87ed2773e076e4fd34c8c6eaf2f32f4) router: trust ip range to set real ip
 - [`a94e38a`](https://github.com/deis/router/commit/a94e38aa031d67db1c9deaa3bf388f3725eeccf6) router: add gzip support (option 2
 - [`2132bce`](https://github.com/deis/router/commit/2132bcefbb625208eff88ec1e15e17c0b09af037) router: make hash sizes configurable
 - [`c66699f`](https://github.com/deis/router/commit/c66699f4c7108fe6c38c37b84aa53bb4f1c5f730) router: support configuration of worker procs and connections
 - [`ce8411f`](https://github.com/deis/router/commit/ce8411f64365e0640e114edcce825e462c37a944) router: configure "basic settings"
 - [`fc1534b`](https://github.com/deis/router/commit/fc1534b87d2578d37d028e57868cea797cbe3cee) router: add support for keepalive_timeout
 - [`dcb4278`](https://github.com/deis/router/commit/dcb42784bf1f5df060114ac5adf4f351b53d4934) router: make platform domain configurable
 - [`51baec3`](https://github.com/deis/router/commit/51baec304ce820c3681a37085c7c5e5d28b28827) router: support stream for builder
 - [`88fc466`](https://github.com/deis/router/commit/88fc4662ae663093994dd7b1928a2a57d903ffae) ci: add build status and report card badges to readme
 - [`22e8677`](https://github.com/deis/router/commit/22e86772baec4f5ec0c40fbba25ef91179879219) ci: integrate with travis
 - [`d8aa506`](https://github.com/deis/router/commit/d8aa506f5d0ac928ce57df173f6fd2e3951e9e3d) Makefile: add make targets for testing
 - [`2392dfd`](https://github.com/deis/router/commit/2392dfda61d062a105599e8ff0054302479997f2) router: send nginx logs to stdout and stderr

#### Fixes

 - [`3a6f182`](https://github.com/deis/router/commit/3a6f1823826deb514f48e0cb6b23cb27fe62ffdd) router: drop error-prone check for services w/o endpoints
 - [`afe9357`](https://github.com/deis/router/commit/afe9357b93cc7375053f4e3c3151a6c78b8c5bf9) router: do not hard code namespace in go code
 - [`758e0c2`](https://github.com/deis/router/commit/758e0c2fd1d544096dcf9628631526a15dad28ac) ci: drive volume mounts of CURDIR instead of PWD
 - [`50d3932`](https://github.com/deis/router/commit/50d39328421574c2677281d63b8913b4a6f6760e) router: pass lint checks by providing missing docstrings
 - [`2242ae5`](https://github.com/deis/router/commit/2242ae57d9f228fcb88160c8c4c687d170660b56) Makefile: check for dev registry before build

#### Documentation

 - [`5668c1d`](https://github.com/deis/router/commit/5668c1d65d26652075eaa16eaab576b9d36124b4) readme: Add note on registry vars on make build

#### Maintenance

 - [`e376f81`](https://github.com/deis/router/commit/e376f815b4ae249bee626d158ce5ed8e8017dfe2) release: set version and lock to deis registry
 - [`0096fc6`](https://github.com/deis/router/commit/0096fc6074915f5862b7ce4e508eddc60cc209e9) glide.lock: add a glide lock file
 - [`8f1464a`](https://github.com/deis/router/commit/8f1464a96825e47f881bc482a99322d5a448b26c) router: set server_name_in_redirect and port_in_redirect off
 - [`c3e8102`](https://github.com/deis/router/commit/c3e8102ae2c61adcfb44e08334e3764410e4d83d) router: inherit more app proxy settings from v1.x
 - [`5bff0dc`](https://github.com/deis/router/commit/5bff0dc0dacb3deda06f3dfffa06d83f99b157db) router: move pid to be consistent with v1.x
 - [`a47bbc1`](https://github.com/deis/router/commit/a47bbc13d7762ae15cb50cc1b6d0821acbc91f06) ci: use generic as travis language
 - [`d8e8803`](https://github.com/deis/router/commit/d8e8803aad4b8e13f4bebda3c86289f98a705725) router: use downward api to set POD_NAMESPACE env var
 - [`4f60d05`](https://github.com/deis/router/commit/4f60d05b2b40ffcf43b7e93e93de4b97eed04e29) router: fix format
 - [`dad4181`](https://github.com/deis/router/commit/dad4181e67563b3d5301f6f9817cfed9371be442) manifests: improve formatting
 - [`9a27e25`](https://github.com/deis/router/commit/9a27e253fc7ac51c0f602096a1ff8a38301facb9) router: prefix annotations with deis.com/
 - [`9791f5a`](https://github.com/deis/router/commit/9791f5ab45623a6dbf8d4fbd6f9152b4913472be) manifest: improve formatting of structured data in routerConfig annotation
 - [`a49393d`](https://github.com/deis/router/commit/a49393da71c8451adabe46e24b1757d3a2b86bc6) manifest: add more port mappings
 - [`2ce48b5`](https://github.com/deis/router/commit/2ce48b53ec9e2bb7084e728fadb15b6755ecde44) Makefile: switch OFF of alpine-based dev environment
 - [`b1e1a89`](https://github.com/deis/router/commit/b1e1a892d5c6be51e8c6b58be58badd0d781b2d7) Makefile: switch to alpine-based dev environment
 - [`2949cad`](https://github.com/deis/router/commit/2949cad4ffcec0092ffa2d296f6e8fea6f0b6333) manifests: remove superfluous app label on rc; add heritage
 - [`9662009`](https://github.com/deis/router/commit/9662009ed13c336a4e21d75b2cfced46bc3efa3f) manifests: replace blank image attribute with comment
