FROM quay.io/deis/base:0.2.0

RUN adduser --system \
	--shell /bin/bash \
	--disabled-password \
	--home /opt/router \
	--group \
	router

COPY . /

RUN apt-get update \
	&& apt-get install -y \
		gcc \
		libgeoip1 \
		libgeoip-dev \
		libssl-dev \
		libpcre3-dev \
		make \
	&& export NGINX_VERSION=1.10.1 VTS_VERSION=0.1.8 BUILD_PATH=/tmp/build PREFIX=/opt/router \
	&& rm -rf "$PREFIX" \
	&& mkdir "$PREFIX" \
	&& mkdir "$BUILD_PATH" \
	&& cd "$BUILD_PATH" \
	&& get_src 1fd35846566485e03c0e318989561c135c598323ff349c503a6c14826487a801 \
		"http://nginx.org/download/nginx-$NGINX_VERSION.tar.gz" \
	&& get_src 6bb9a36d8d70302d691c49557313fb7262cafd942a961d11a2730d9a5d9f70e0 \
		"https://github.com/vozlt/nginx-module-vts/archive/v$VTS_VERSION.tar.gz" \
	&& cd "$BUILD_PATH/nginx-$NGINX_VERSION" \
	&& ./configure \
	  --prefix="$PREFIX" \
	  --pid-path=/tmp/nginx.pid \
	  --with-debug \
	  --with-pcre-jit \
	  --with-ipv6 \
	  --with-http_ssl_module \
	  --with-http_stub_status_module \
	  --with-http_realip_module \
	  --with-http_auth_request_module \
	  --with-http_addition_module \
	  --with-http_dav_module \
	  --with-http_geoip_module \
	  --with-http_gzip_static_module \
	  --with-http_sub_module \
	  --with-http_v2_module \
	  --with-mail \
	  --with-mail_ssl_module \
	  --with-stream \
	  --add-module="$BUILD_PATH/nginx-module-vts-$VTS_VERSION" \
	&& make \
	&& make install \
	&& apt-get remove -y --auto-remove --purge \
		gcc \
		libgeoip-dev \
		libssl-dev \
		libpcre3-dev \
		make \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /usr/share/man /usr/share/doc \
	&& rm -rf "$BUILD_PATH"

# Re-copy these files because the previous step will have overwritten them
COPY opt/router /opt/router

# Fix some permissions since we'll be running as a non-root user
RUN chown -R router:router /opt/router

USER router

CMD ["/opt/router/sbin/boot"]
EXPOSE 2222 8080 6443 9090
