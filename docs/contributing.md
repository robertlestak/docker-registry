# Registry Manager Contribution

The `registrymanager` enables authenticating, authorizing, and transforming, requests before proxying them to the Docker Registry.

Authentication is enabled through LDAP / AD and local users.

## GoDoc

Until `godoc` can support go modules, run the following to generate the latest documentation:

````
mkdir -p $GOROOT/src/registrymanager
cp -r registrymanager $GOROOT/src/registrymanager
godoc -http=":6060"
````

Then navigate to `http://localhost:6060/pkg/registrymanager/` to view the source documentation.
