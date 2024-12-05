docker-start:
	docker compose -f docker/compose.yaml up -d

docker-stop:
	docker compose -f docker/compose.yaml down

docker-clean:
	make docker-stop
	docker compose -f docker/compose.yaml rm -v

patch-genesis:
	cp docker/genesis-besu.json.template docker/genesis-besu.json
	cp docker/genesis-geth.json.template docker/genesis-geth.json
	./docker/patch_genesis.sh docker/genesis-besu.json docker/genesis-geth.json

docker-clean-start:
	make docker-clean
	make patch-genesis
	make docker-start