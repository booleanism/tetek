include .env

$(eval export $(shell sed -ne 's/ *#.*$$//; /./ s/=.*$$// p' .env))

VERSION:=v0.0.1

.PHONY: migrate docs

setup: .deps
	go install -tags 'pgx5' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/booleanism/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1-p5
	go mod tidy

.deps:
	docker build -t deps -f deps.dockerfile .
	docker compose up -d
	@touch .deps

clean-deps:
	docker compose down
	@rm .deps

build-account: .deps
	docker build -t account:${VERSION} -f account/service.dockerfile .

build-account-worker: .deps
	docker build -t account-worker:${VERSION} -f account/worker.dockerfile .

build-auth: .deps
	docker build -t auth:${VERSION} -f auth/service.dockerfile .

build-auth-worker: .deps
	docker build -t auth-worker:${VERSION} -f auth/worker.dockerfile .

build-feeds: .deps
	docker build -t feeds:${VERSION} -f feeds/service.dockerfile .

build-feeds-worker: .deps
	docker build -t feeds-worker:${VERSION} -f feeds/worker.dockerfile .

run-account: build-account
	docker run --env-file .env --network=tetek6_accountnet --network=tetek6_brokernet --name account-service -it --rm -p 8082:8082 account:${VERSION} 

run-auth: build-auth
	docker run --env-file .env --network=tetek6_brokernet --name auth-service -it --rm -p 8081:8081 auth:${VERSION} 

run-feeds: build-feeds
	docker run --env-file .env --network=tetek6_feedsnet --network=tetek6_brokernet --name feeds-service -it --rm -p 8083:8083 feeds:${VERSION} 

run-account-worker: build-account-worker
	docker run --env-file .env --network=tetek6_accountnet --network=tetek6_brokernet --name account-worker -it --rm -p 8082:8082 account-worker:${VERSION} 

run-auth-worker: build-auth-worker
	docker run --env-file .env --network=tetek6_brokernet --name auth-worker -it --rm -p 8081:8081 auth-worker:${VERSION} 

run-feeds-worker: build-feeds-worker
	docker run --env-file .env --network=tetek6_feedsnet --network=tetek6_brokernet --name feeds-worker -it --rm -p 8083:8083 feeds-worker:${VERSION} 

gen: docs-feeds docs-comments

gen-feeds:
	oapi-codegen -config feeds/cmd/http/api/config.yaml docs/api/v0/feeds.yaml

gen-comments:
	oapi-codegen -config comments/cmd/http/api/config.yaml docs/api/v0/comments.yaml

migrate: migrate-account migrate-feeds migrate-comments

migrate-account:
	cat .env | grep 'ACCOUNT_DB_STR' | replace 'postgres' 'pgx5' | xargs -I{} -r migrate -database '{}' -path account/schemas/ up

migrate-feeds:
	cat .env | grep 'FEEDS_DB_STR' | replace 'postgres' 'pgx5' | xargs -I{} -r migrate -database '{}' -path feeds/schemas/ up

migrate-comments:
	cat .env | grep 'COMMS_DB_STR' | replace 'postgres' 'pgx5' | xargs -I{} -r migrate -database '{}' -path comments/schemas/ up
