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
	&& rm -rf /var/cache/apk/*

# add router user
RUN addgroup -S router && \
	adduser -S -G router -H -h /opt/router -D router

COPY . /

# compile nginx from source
RUN build

# Re-copy these files because the previous step will have overwritten them
COPY opt/router /opt/router

# Fix some permission since we'll be running as a non-root user.
RUN chown -R router:router /opt/router

USER router

CMD ["/opt/router/sbin/boot"]
EXPOSE 2222 8080 6443 9090

ENV WORKFLOW_RELEASE 2.0.0
