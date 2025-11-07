include .env

.PHONY: run-http-account

run-http-account:
	ACCOUNT_DB_STR=${ACCOUNT_DB_STR} BROKER_STR=${BROKER_STR} go run account/cmd/http/main.go

run-http-auth:
	BROKER_STR=${BROKER_STR} AUTH_JWT_SECRET=${AUTH_JWT_SECRET} go run auth/cmd/http/main.go

run-http-feeds:
	BROKER_STR=${BROKER_STR} FEEDS_DB_STR=${FEEDS_DB_STR} go run feeds/cmd/http/main.go

run-worker-account:
	ACCOUNT_DB_STR=${ACCOUNT_DB_STR} BROKER_STR=${BROKER_STR} go run account/cmd/worker/main.go

run-worker-auth:
	BROKER_STR=${BROKER_STR} AUTH_JWT_SECRET=${AUTH_JWT_SECRET} go run auth/cmd/worker/main.go

docs-feeds:
	oapi-codegen -config feeds/cmd/http/api/config.yaml docs/api/v0/feeds.yaml

docs-comments:
	oapi-codegen -config comments/cmd/http/api/config.yaml docs/api/v0/comments.yaml
