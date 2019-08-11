secrets.json:
	./scripts/secrets load

vaultsecrets:
	./scripts/secrets create

config.yml: secrets.json
	./scripts/storage
	rm -f secrets.json

database:
	./database/scripts/init

registrymanager:
	cd registrymanager && docker build . -t registrymanager:latest

deploy: registrymanager config.yml
	docker-compose up -d
	#docker stack deploy -c docker-compose.yml registry
	sleep 10
	$(MAKE) database

.PHONY: deploy database config.yml secrets.json registrymanager
