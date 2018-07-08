# nginx-proxy

Simple NGINX reverse proxy server.

## Configuration

The reverse proxy domain is defined on container build.

If building as a standalone service, `docker build . -t proxyserver --build-arg PROXY_DOMAIN=example.com --build-arg PROXY_SSL=true`.

To utilize Let's Encrypt SSL, set the `PROXY_SSL` build arg to `true`, otherwise set it to false (or leave empty, the default is `false`).

If deploying with `docker-compose`, set the build arg in the `docker-compose.yml` file.
