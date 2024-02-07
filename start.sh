#!/bin/bash
source .env

go run . \
    --ecdsa-private-key $PRIVATE_KEY \
    --eth-http-url $ETH_HTTP_URL \
    --registry-coordinator-addr $REGISTRY_COORDINATOR_ADDR \
    --operator-state-retriever-addr $OPERATOR_STATE_RETRIEVER_ADDR \
    --sync-interval $SYNC_INTERVAL
