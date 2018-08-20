init: proxy-dirs

config: clean proxy-dirs
	cd nginx-proxy && ./config.sh

clean:
	rm -rf nginx-proxy/proxy_data/repository-access \
		nginx-proxy/proxy_data/htpasswds nginx-proxy/proxy_data/conf \
		nginx-proxy/proxy_data/services


deploy: proxy-dirs proxy
	docker-compose up -d

services: clean proxy-dirs
	cd nginx-proxy && ./config.sh touch-services

proxy-dirs:
	mkdir -p nginx-proxy/proxy_data nginx-proxy/access \
		nginx-proxy/access/htpasswds nginx-proxy/access/services \
		nginx-proxy/proxy_data/repository-access nginx-proxy/proxy_data/htpasswds \
		nginx-proxy/proxy_data/conf
	touch nginx-proxy/access/htpasswds/admin
	touch nginx-proxy/access/htpasswds/users

reload: config
	docker restart registry_proxy

.PHONY: proxy proxy-dirs reload clean init deploy
