# Plasma x Celestia:

## Overview:

This repo implements a celestia `da-server` for plasma mode using generic
commitments.

The `da-server` connects to a celestia-node running as a sidecar process.

celestia da-server accepts the following flags for celestia storage using
[celestia-openrpc](https://github.com/celestiaorg/celestia-openrpc)

````
    --celestia.auth-token value                                            ($OP_PLASMA_DA_SERVER_CELESTIA_AUTH_TOKEN)
          celestia auth token

    --celestia.namespace value                                             ($OP_PLASMA_DA_SERVER_CELESTIA_NAMESPACE)
          celestia namespace

    --celestia.server value             (default: "http://localhost:26658") ($OP_PLASMA_DA_SERVER_CELESTIA_SERVER)
          celestia server endpoint

````

The celestia server endpoint should be set to the celestia-node rpc server,
usually `http://127.0.0.1:26658`.

A random valid namespace can be generated by running the following command:

```sh
export NAMESPACE=00000000000000000000000000000000000000$(openssl rand -hex 10)
```

The auth token is auto generated and is required to submit blob transactions to
the celestia-node.

## Testing:

Follow the usual instructions for setting up devnet.

There are two ways to test the celestia `da-server` in devnet and testnet
modes.

### Using rollop:

The first way uses a fork of [rollop](https://github.com/0xFableOrg/roll-op) to
automate deployment. This method is recommended for beginners.

Clone the [rollop](https://github.com/celestiaorg/roll-op) repository, checkout
the `plasma` branch and follow the instructions in the README to setup the
rollop stack.

```sh
git clone https://github.com/celestiaorg/roll-op.git
cd roll-op
git checkout plasma
./rollop setup --yes
````

rollop is configured to use `mocha-4` celestia network by default in both
devnet and testnet modes.

#### Devnet

Once the rollop stack is setup, you can start the devnet with the following
command:

```sh
./rollop --clean devnet
```

This will start the devnet in plasma mode with celestia-node running in testnet
mode.

You can verify that the devnet is running by checking da-server logs:

```sh
tail -f deployments/rollup/logs/da_server.log
```

Upon successful operation, the logs should contain the following message:

```sh
t=2024-05-30T19:08:30+0000 lvl=info msg="Using celestia storage" url=http://da:26658
t=2024-05-30T19:08:32+0000 lvl=info msg="celestia: blob successfully submitted" id=0900000000000000b25a32154ab00902cfc0269b3239b612ebfe7f7263545119ee7251cc72728142
t=2024-05-30T19:08:34+0000 lvl=info msg="celestia: blob successfully submitted" id=0a00000000000000cb559bc3c6a01b0819460ce86c13165fdc58ac9c81c1e52404f8c4b36097db87
t=2024-05-30T19:08:34+0000 lvl=info msg="celestia: blob request" id=010c0900000000000000b25a32154ab00902cfc0269b3239b612ebfe7f7263545119ee7251cc72728142
```

#### Testnet

Copy example config `config.toml.example` to `config.toml` and update any
fields as required.

Next start the testnet with the following command:


```sh
./rollop --name=celopstia-11177111 --preset=prod --config=config.toml l2
```

You can verify that the testnet is running by checking da-server logs:

```sh
tail -f deployments/celopstia-11177111/logs/da_server.log
```

Upon successful operation, the logs should contain the following message:

```sh
t=2024-06-03T21:36:32+0000 lvl=info msg="Using celestia storage" url=http://127.0.0.1:26658
t=2024-06-03T21:36:32+0000 lvl=info msg="Started DA Server"
t=2024-06-03T21:37:16+0000 lvl=debug msg=PUT url=/put/
t=2024-06-03T21:37:19+0000 lvl=info msg="celestia: blob successfully submitted" id=430f1e0000000000f34842f2c61d3b1fc1fad749ed442320fdce7023669f3116ea8df024ce792b7f
t=2024-06-03T21:38:23+0000 lvl=debug msg=PUT url=/put/
t=2024-06-03T21:38:28+0000 lvl=debug msg=GET url=/get/0x010c430f1e0000000000f34842f2c61d3b1fc1fad749ed442320fdce7023669f3116ea8df024ce792b7f
t=2024-06-03T21:38:28+0000 lvl=info msg="celestia: blob request" id=010c430f1e0000000000f34842f2c61d3b1fc1fad749ed442320fdce7023669f3116ea8df024ce792b7f
```

To configure the da-server to use a different namespace, update the
`da_namespace` field in `config.toml`.

To configure the da-server to use a different celestia network, you'll need to
update the `--p2p.network` flag in `l2_batcher.py`, `l2_node.py` and
`celestia_light_node.py` for now. This will be configurable in the future.

##### Troubleshooting:

Check the logs for the `da-server` to see if it is running successfully:

```sh
tail -f deployments/rollup/logs/da_server.log
```

Additionally verify that the batcher is able to submit transactions to the celestia-node:

```sh
tail -f deployments/rollup/logs/l2_batcher.log
```

### Using docker-compose:

The second way to test the celestia `da-server` uses docker-compose to deploy
the rollup stack. This method reuses the bedrock devnet docker-compose setup
and is recommended for developers.

#### Devnet:

For testing devnet, we'll use the local-celestia-devnet docker image which
simulates the whole celestia network including a local celestia-node and does
not require a peer-to-peer connection.

To run a local devnet with local-celestia-devnet:

```sh
DEVNET_PLASMA="true" GENERIC_PLASMA="true"  make devnet-up
```

This will start the devnet in plasma mode with local-celestia-devnet.

You can verify that the devnet is running by checking da-server logs:

```sh
docker logs -f ops-da-server-1 | grep celestia
```

Upon successful operation, the logs should contain the following message:

```sh
t=2024-05-30T19:08:30+0000 lvl=info msg="Using celestia storage" url=http://da:26658
t=2024-05-30T19:08:32+0000 lvl=info msg="celestia: blob successfully submitted" id=0900000000000000b25a32154ab00902cfc0269b3239b612ebfe7f7263545119ee7251cc72728142
t=2024-05-30T19:08:34+0000 lvl=info msg="celestia: blob successfully submitted" id=0a00000000000000cb559bc3c6a01b0819460ce86c13165fdc58ac9c81c1e52404f8c4b36097db87
t=2024-05-30T19:08:34+0000 lvl=info msg="celestia: blob request" id=010c0900000000000000b25a32154ab00902cfc0269b3239b612ebfe7f7263545119ee7251cc72728142
```

#### Testnet:

Usually, it's better to deploy both OP stack and celestia-node in the same
network, but for ease of testing, we will modify the devnet docker-compose
file to submit to celestia testnet instead.

For running plasma in celestia testnet mode, we'll use the celestia-node docker
image which requires a [celestia testnet network like mocha or
arabica](https://docs.celestia.org/nodes/participate).

To run a celestia-node in testnet mode,
follow the instructions for [setting up a light node](https://docs.celestia.org/developers/optimism#setting-up-your-light-node)
and changes required to the [docker-compose file](https://docs.celestia.org/developers/optimism#docker-changes)

Note that in this case, the auth token needs to be configured using the
CELESTIA_NODE_AUTH_TOKEN environment variable.

To obtain the auth token for celestia-node,
check [the documentation for the node type you are using](https://docs.celestia.org/developers/node-tutorial#auth-token)

Once the docker compose file is modified, you can run the devnet with the following command:

```sh
DEVNET_PLASMA="true" GENERIC_PLASMA="true"  make devnet-up
```

This will start the devnet in plasma mode with celestia-node running in testnet
mode.

You can verify that the devnet is running by checking da-server logs:

```sh
docker logs -f ops-bedrock-da-server-1 | grep celestia
```

Once again, upon successful operation, the logs should contain the following
message:

```sh
t=2024-05-30T19:08:30+0000 lvl=info msg="Using celestia storage" url=http://da:26658
t=2024-05-30T19:08:32+0000 lvl=info msg="celestia: blob successfully submitted" id=0900000000000000b25a32154ab00902cfc0269b3239b612ebfe7f7263545119ee7251cc72728142
t=2024-05-30T19:08:34+0000 lvl=info msg="celestia: blob successfully submitted" id=0a00000000000000cb559bc3c6a01b0819460ce86c13165fdc58ac9c81c1e52404f8c4b36097db87
t=2024-05-30T19:08:34+0000 lvl=info msg="celestia: blob request" id=010c0900000000000000b25a32154ab00902cfc0269b3239b612ebfe7f7263545119ee7251cc72728142
```

As an alternative, you may wish to deploy OP stack and celestia-node both to
testnets e.g. sepolia and mocha.

In this case, run a [celestia light node](https://docs.celestia.org/nodes/light-node#setting-up-your-light-node)
and da-server, then follow the instructions for
[creating your own L2 rollup testnet](https://docs.optimism.io/builders/chain-operators/tutorials/create-l2-rollup)
including the following flags:

```sh
op-node
      --plasma.enabled=true
      --plasma.da-service=true
```

```sh
op-batcher
      --plasma.enabled=true
      --plasma.da-service=true
```

```sh
da-server
      --generic-commitment=true
      --celestia.server=http://localhost:26658
      --celestia.auth-token=$CELESTIA_NODE_AUTH_TOKEN
      --celestia.namespace=$NAMESPACE
```

where `$CELESTIA_NODE_AUTH_TOKEN` is the auth token for the celestia-node and `$NAMESPACE` is a random valid namespace.

##### Troubleshooting:

Check the logs for the `da-server` to see if it is running successfully:

```sh
docker logs -f ops-bedrock-da-server-1
```

Additionally verify that the batcher is able to submit transactions to the celestia-node:

```sh
docker logs -f ops-bedrock-op-batcher-1
```
