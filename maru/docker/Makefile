docker-start:
	docker compose -f compose.yaml up -d

docker-stop:
	docker compose -f compose.yaml down

docker-clean:
	make docker-stop
	docker compose -f compose.yaml rm -v

patch-genesis:
	cp genesis-besu.json.template genesis-besu.json
	cp genesis-geth.json.template genesis-geth.json
	./patch_genesis.sh genesis-besu.json genesis-geth.json

docker-clean-start:
	make docker-clean
	make patch-genesis
	make docker-start
