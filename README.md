# docker-registry

A basic self-hosted Docker Registry server with NGINX reverse proxy and authentication.

## Configuration

````
cp .env-sample .env
````

To utilize SSL, set `CERT_PATH` and `CERT_KEY` to the path to your SSL cert and keys which are accessible from within the container - therefore it is recommended to place these somewhere in `proxy_data`.

### Authentication

`make init` will create `access/htpasswds` and `access/services` directories.

In `access/htpasswds/users`, enter the `htpasswd` user:pass for each restricted-access user.

`make services` will create a file in `access/services` for each user listed in `access/htpasswds/users`.

In `access/services/[user]`, list all of the repositories to which you would like to enable the user to access.

Any user listed in `access/htpasswds/admin` will be granted full access to the registry API.

Once you have set up the desired user access levels, run `make config`.

You can modify the users / services at any time - run `make reload` for your changes to take effect if the proxy is already running.

## Deployment

````
make deploy
````
