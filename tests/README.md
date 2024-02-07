# Tests
the files in this directory are copied from eigencert:
- https://github.com/Layr-Labs/eigencert/blob/master/contracts/script/output/31337/eigencert_deployment_output.json
- https://github.com/Layr-Labs/eigencert/blob/master/tests/anvil/eigenlayer-eigencert-eigenda-strategies-deployed-operators-registered-anvil-state.json

We use them in the integration test since they already have an avs (eigencert) deployed with operators registered into eigenlayer.

TODO: we should have a proper integration test setup in eigenlayer-middleware directly with db files set there for a dummy avs instead of copying from eigencert. Also this way we will have to keep updating the db files here as eigencert changes.