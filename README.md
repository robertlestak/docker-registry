# docker-registry

A basic self-hosted Docker Registry server with NGINX reverse proxy and authentication.

## Configuration

Use `certbot` to install a Let's Encrypt SSL cert on the machine if you are using SSL.

`mkdir proxy_data`, create a `.htpasswd` file for your authorized users and place it in this directory.

````
cp .env-sample .env
# Configure accordingly

# Deploy with proxy
docker-compose -f docker-compose.yml -f docker-compose-proxy.yml up -d
````

## Potential Improvements

- I have been meaning to look into the possibility of using a Traefik container for the reverse proxy as it is much more flexible than and automated the current NGINX configuration.

- Currently any user authed with the `.htpasswd` file has access to all the repositories in the registry. A potential future modification would be a system to limit user access to specific repositories.
