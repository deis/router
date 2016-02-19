FROM alpine:3.3

# install common packages
RUN apk add --update-cache \
	bash \
	curl \
	geoip \
	libssl1.0 \
	openssl \
	pcre \
	sudo \
	libcap \
	&& rm -rf /var/cache/apk/*

# add router user
RUN addgroup -S router && \
	adduser -S -G router -H -h /opt/router -D router

COPY . /

# compile nginx from source
RUN build

# Re-copy these files because the previous step will have overwritten them
COPY opt/router /opt/router

RUN mkdir /opt/router/ssl

# Fix some permission since we'll be running as a non-root user.  This includes enabling
# Nginx to always bind to low/privileged ports, even when running as a non-root user.
RUN chown -R router:router /opt/router && \
	setcap 'cap_net_bind_service=+ep' /opt/router/sbin/nginx

USER router

CMD ["/opt/router/sbin/boot"]
EXPOSE 80 2222 9090

ENV DEIS_RELEASE 2.0.0
