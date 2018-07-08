FROM nginx

COPY build.sh /etc/nginx/
COPY nginx.base.conf /etc/nginx/nginx.base.conf
COPY nginx.ssl.base.conf /etc/nginx/nginx.ssl.base.conf

ARG PROXY_DOMAIN=localhost
ENV PROXY_DOMAIN=$PROXY_DOMAIN

ARG PROXY_SSL=false
ENV PROXY_SSL=$PROXY_SSL

RUN cd /etc/nginx/ && \
    chmod +x build.sh && \
    sh ./build.sh

VOLUME ["/data"]

EXPOSE 80
EXPOSE 443
