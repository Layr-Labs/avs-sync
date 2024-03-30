############################# HELP MESSAGE #############################
# Make sure the help command stays first, so that it's printed by default when `make` is called without arguments
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Currently exporting the .env vars directly in the commands that need them
# see for eg. the run-avs-sync command
# Uncommenting the two lines below will export the .env vars for all commands
# include .env
# export

start-anvil-holesky-fork: ## 
	anvil --fork-url https://ethereum-holesky-rpc.publicnode.com

run-avs-sync: ## 
	@# we export the env vars from .env file and then run the go program
	@set -a; . .env; set +a; go run .

test: ## 
	go test ./...

generate-test-coverage: ## 
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

docker-build: ## builds avs-sync docker image
	docker build -t ghcr.io/layr-labs/avs-sync .

docker-push: ## publishes avs-sync docker image (run docker login ghcr.io -u <username> first)
	docker push ghcr.io/layr-labs/avs-sync
