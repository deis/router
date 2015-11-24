FROM alpine:3.2

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

# add nginx user
RUN addgroup -S nginx && \
  adduser -S -G nginx -H -h /opt/nginx -s /sbin/nologin -D nginx

COPY . /

# compile nginx from source
RUN build

# Re-copy this file because the previous step will have overwritten it
COPY opt/nginx/conf/nginx.conf /opt/nginx/conf/nginx.conf

CMD ["/bin/router"]
EXPOSE 80 2222 9090

ENV DEIS_RELEASE 2.0.0
