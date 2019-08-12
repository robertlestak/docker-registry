# Docker Registry API

All `/v2/` requests are proxied to the Docker Registry through the `registrymanager` application.

Registry documentation is available at: https://docs.docker.com/registry/spec/api/

## Management API

Authentication, Authorization, and user management is done through the `registrymanager` application.

Properly authenticated and authorized access will be allowed to communicate with the backing registry.

## Endpoints

The following administrative endpoints are available through the API. All endpoints require Basic Authentication. Unless otherwise noted, the user must be authorized as an administrator.

````
GET /user
    Retrieve access data for the current user. If authenticated as admin, `username` parameter can be used to look up other users.
    Params: username (string, optional)
POST /user
    Create a new user. If `ad` is provided, password will be ignored.
    Params: username (string), ad (bool, opt), password (string, opt), admin (bool, opt), namespaces (csv, opt)
PATCH /user
    Update an existing user's data. Will overwrite any existing data with the provided data. If `ad` is provided, password will be ignored.
    Params: username (string), ad (bool, opt), password (string, opt), admin (bool, opt), namespaces (csv, opt)
DELETE /user
    Delete a user from the system and disable access. Will not affect any images pushed by user.
    Params: username (string)
GET /users
    Retrieve list of all users in system.
    Params: offset (int, default 0), limit (int, default / max 50)
POST /user/password
    Update the password for the current user. Can only be executed if the user is not an AD user. If the executing user is an admin and provides the "username" param, the password will be changed for that user.
    Params: password (string), username (string, opt)
POST /user/namespaces
    Update the namespaces for a user.
    Params: username (string), namespaces (csv)
GET /v2/_catalog
    Retrieve the available catalog listing for the current user.

ALL /
    All remaining requests are proxied to the Docker Registry. Authentication and authorization are enabled in the proxy to ensure users are only able to request resources to which they have been granted access.
````
