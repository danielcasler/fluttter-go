# ==============================================================================
# Modules support
SHELL := /bin/bash

server:
	go run cmd/server/main.go

graphql:
	go run cmd/graphql/main.go

.PHONY: server graphql
