# AvsSync

### Motivation

Currently, when an AVS is being serviced by some operators, the operators for it are delegated some stake.

The architecture going over how the on chain contracts are set up for an AVS is outlined [here in the AVS middleware contracts repo](https://github.com/Layr-Labs/eigenlayer-middleware#registries). An important one to note for AvsSync is the StakeRegistry "which keeps track of the stakes of different operators for different quorums at different times". AvsSync calls the RegistryCoordinator entrypoint, but this call is proxied to the StakeRegistry contract.

That is, every time an operator registers to serve an AVS or deregisters to no longer serve the AVS, the stake they’ve been delegated to enters or exits (respectively) the relevant quorums in the AVS.

AvsSync is a cron job executable that AVS teams and/or operators can run so that AVS operators keep up-to-date with the latest stake state for the quorums of the AVS they’re in.

### Problem

Upon registration/deregistration and the updating of some quorums with new stake amounts, the current implementation of the StakeRegistry contract only updates the stake amount for the operator performing the action against the AVS. The rest of the operators in the AVS who are in the updated quorums do not have the latest view of the state of the quorum, meaning the current model is not push based.

### Solution

AvsSync implements this update in a pull model by having AVS teams run a cron job as part of their software deployment stack that periodically calls `UpdateOperatorsForQuorum` on the `RegistryCoordinator` (which proxies the call to the `StakeRegistry`). AvsSync periodically (once per day by default) queries the current set of registered operators for each quorum and updates their stake amounts (by calling the DelegationManager contract).

### Configuration

AvsSync is configured via flags passed as arguments, or via environment variables for the respective flags. The list of flags is listed in [flags.go](./flags.go)

### Dependencies

AvsSync makes use of [`eigensdk-go`](https://github.com/Layr-Labs/eigensdk-go), and requires an ethereum node running at `--eth-http-url` to be able to make calls to the chain.

### Running AvsSync

Just run
```
docker run github.com/layr-labs/avs-sync --flags
```
Here's an example of flags for running against an anvil image (fake addresses, this won't work):
```
docker run github.com/layr-labs/avs-sync \
   --ecdsa-private-key ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
   --eth-http-url http://localhost:8545 \
   --registry-coordinator-addr 0x5FbDB2315678afecb367f032d93F642f64180aa3 \
   --operator-state-retriever-addr 0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512 \
   --sync-interval 24h
```

### Running AvsSync Locally (from source)

AvsSync can be run directly (passing the necessary flags) by:
```
go run . --flags
```
or also by building and installing it so that it can be run from any directory:
```
go install .
avs-sync --flags
```

### Building AvsSync Docker Image (for eigenlayer devs)

You can build and push the image to our repository (will require ghcr login) by running:
```
make docker-build
## need to login if not
docker login ghcr.io -u USER_NAME
make docker-push
```

### Testing

#### Against saved anvil db state

The test is run using an eigencert deployment saved [anvil db state file](./tests/eigenlayer-eigencert-eigenda-strategies-deployed-operators-registered-with-eigenlayer-anvil-state.json). It also requires the [ContractsRegistry bindings](./bindings/ContractsRegistry/binding.go), which we copied here from eigencert. 

The test can be run via:
```
make test
```

#### Against a holesky fork

The most recent eigenDA holesky deployment is accessible [here](https://github.com/Layr-Labs/eigenda/blob/master/contracts/script/deploy/holesky/output/holesky_testnet_deployment_data.json). We can test avssync against this deployment by running a holesky fork and running the tests against it.

First create a .env file by copying the example .env.example file, and adjust the variables as needed. You will need to enter a private key that has holesky eth, and you will most likely also need to set `AVS_SYNC_READER_TIMEOUT_DURATION` to at least 1m, since retrieving the operator state can take a while the first time anvil is querying the holesky fork and filling its local cache.

Then run
```
make start-anvil-holesky-fork
```
and in a separate terminal
```
make run-avs-sync
```
