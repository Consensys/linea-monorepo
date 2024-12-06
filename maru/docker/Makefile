docker-start:
	docker compose -f compose.yaml up -d

docker-stop:
	docker compose -f compose.yaml down

docker-clean:
	make docker-stop
	docker compose -f compose.yaml rm -v

docker-clean-start:
	make docker-clean
	make docker-start
