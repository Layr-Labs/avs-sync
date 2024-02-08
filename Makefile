############################# HELP MESSAGE #############################
# Make sure the help command stays first, so that it's printed by default when `make` is called without arguments
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

start-anvil-goerli-fork: ## 
	anvil --fork-url https://goerli.gateway.tenderly.co

run-avs-sync: ## 
	./run.sh

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
