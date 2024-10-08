# nuklaivm

## Disclaimer

**IMPORTANT NOTICE:** This project is currently in the alpha stage of development and is not intended for use in production environments. The software may contain bugs, incomplete features, or other issues that could cause it to malfunction. Use at your own risk.

We welcome contributions and feedback to help improve this project, but please be aware that the codebase is still under active development. It is recommended to thoroughly test any changes or additions before deploying them in a production environment.

Thank you for your understanding and support!

## Overview

`nuklaivm` takes inspiration from [morpheusvm](https://github.com/ava-labs/hypersdk/tree/main/examples/morpheusvm) and
[tokenvm](https://github.com/ava-labs/hypersdk/tree/main/examples/tokenvm) and implements the functionality of both of
these VMs. In addition, `nuklaivm` also adds additional functionality such as staking native token `NAI`, has an
emission balancer that keeps track of total supply of NAI, max supply of NAI, staking rewards per block and the emission
address to direct 50% of all fees to.

## Status

`nuklaivm` is considered **ALPHA** software and is not safe to use in
production. The framework is under active development and may change
significantly over the coming months as its modules are optimized and
audited.

## Features

### Actions

- ☑ Transfer both the native asset `NAI` and any other token created by users
- ☑ Create a token(fungible, non-fungible, dataset)
- ☑ Mint a token(fungible, non-fungible)
- ☑ Burn a token(fungible, non-fungible)
- ☑ Register validator for staking
- ☑ Withdraw validator from staking
- ☑ Delegate NAI to any currently staked validator
- ☑ Undelegate NAI from a staked validator
- ☑ Claim Validator staking rewards
- ☑ Claim User delegation rewards
- ☑ Create dataset
- ☑ Create dataset using an existing token of type dataset
- ☑ Initiate contribution to the dataset
- ☑ Complete contribution to the dataset
- ☑ Publish the dataset to Nuklai marketplace
- ☑ Subscribe to the dataset in the Nuklai marketplace
- ☑ Claim accumulated subscription payment from the Nuklai marketplace

### Emission Balancer

- ☑ Tracks total supply of NAI, max supply of NAI, staking rewards per block and the emission address to direct 50% of all fees to
- ☑ Register validator for staking
- ☑ Unregister validator from staking
- ☑ Delegate `NAI` to a validator
- ☑ Undelegate `NAI` from a validator
- ☑ Claim the staking/delegation rewards
- ☑ Track the staking information for each users and validators
- ☑ Distribute 50% fees to emission balancer address and 50% to all the staked validators per block
- ☑ Distribute NAI as staking rewards to the validators that have a minimum stake of at least 100 NAI per block

### Deep Dive on different features of `nuklaivm`

#### Arbitrary Token Minting

The basis of the `nuklaivm` is the ability to create, mint, and transfer user-generated
tokens with ease. When creating an asset, the owner is given "admin control" of
the asset functions and can later mint more of an asset, update its metadata
(during a reveal for example), or transfer/revoke ownership (if rotating their
key or turning over to their community).

Assets are a native feature of the `nuklaivm` and the storage engine is
optimized specifically to support their efficient usage (each balance entry
requires only 72 bytes of state = `assetID|publicKey=>balance(uint64)`). This
storage format makes it possible to parallelize the execution of any transfers
that don't touch the same accounts. This parallelism will take effect as soon
as it is re-added upstream by the `hypersdk` (no action required in the
`nuklaivm`).

#### Emission Balancer and Staking Mechanism

On `nuklaivm`, the emission balancer handles the staking mechanism whereby it tracks the
total supply of `NAI`, max supply of `NAI`, staking rewards per block and the emission address to direct 50% of all fees to.
Furthermore, it also rewards all the validators that have a minimum stake and all the users who have a minimum delegated stake to a validator of their choice.

Read more about [Emission Balancer](./docs/emission_balancer/README.md).

## Getting Started

### Building NuklaiVM from Source

#### Clone the NuklaiVM Repo

Clone the NuklaiVM repository:

```sh
git clone git@github.com:Nuklai/nuklaivm.git
cd nuklaivm
```

This will clone and checkout the `main` branch.

#### Building NuklaiVM

Build NuklaiVM by running the build script:

```sh
./scripts/build.sh
```

The `nuklaivm` binary is now in the `build/./` directory.

### Run Integration Tests

To run the integration tests for NuklaiVM, run the following command:

```sh
./scripts/tests.integration.sh
```

You should see the following output:

```bash
[DeferCleanup (Suite)]
/Users/user/go/src/github.com/ava-labs/hypersdk/tests/integration/integration.go:156
[DeferCleanup (Suite)] PASSED [0.000 seconds]
------------------------------

Ran 12 of 12 Specs in 1.120 seconds
SUCCESS! -- 12 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS
coverage: [no statements]
composite coverage: [no statements]

Ginkgo ran 1 suite in 6.748482403s
Test Suite Passed
```

### Running a Local Network

```bash
./scripts/run.sh;
```

By default, it will store all the subnet files under `$HOME/.hypersdk`

This also allocates all funds on the network to `created address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9`. The private
key for this address is `323b1d8f4eed5f0da9da93071b034f2dce9d2d22692c172f3cb252a64ddfafd01b057de320297c29ad0c1f589ea216869cf1938d88c9fbd70d6748323dbf2fa7`.
For convenience, this key has is also stored at `demo.pk`.\_

To stop the network, run:

```sh
./scripts/stop.sh
```

The run script uses AvalancheGo's [tmpnet](https://github.com/ava-labs/avalanchego/tree/master/tests/fixture/tmpnet) to launch a 2 node network with one node's server running at the hardcoded URI: `http://127.0.0.1:9650/ext/bc/nuklaivm`.

Each default API comes with its own extension. For example, you can get the `networkID`, `subnetID`, and `chainID` by hitting the `/coreapi` extension:

```bash
curl -X POST --data '{
    "jsonrpc":"2.0",
    "id"     :1,
    "method" :"hypersdk.network",
    "params" : {}
}' -H 'content-type:application/json;' 127.0.0.1:9650/ext/bc/nuklaivm/coreapi
```

This should return the following JSON:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "networkId": 88888,
    "subnetId": "yfnUMucMT51SgDQRHS5dGZrHKSLzhp49ReUfp9abfHBhQ4XV2",
    "chainId": "2P3GpJ3BDN9u8ejJ9tfqvMy5deub1BevNmt2A9qhAcUWskioqW"
  },
  "id": 1
}
```

Note: if you run into any issues starting your network, try running the following commands to troubleshoot and create a GitHub issue to report:

```bash
ps aux | grep avalanchego
```

If this prints out a number of AvalancheGo processes that are still running:

```bash
user    18157  18.0  1.4 413704400 466256   ??  Ss   Tue08AM 293:30.87 /Users/user/.hypersdk/avalanchego-d729e5c7ef9f008c3e89cd7131148ad3acda2e34/avalanchego --config-file /Users/user/.tmpnet/networks/20240903-083857.802339-nuklaivm-e2e-tests/NodeID-u9eMTbMcPWAj3yer1jhdheJTZL3yvC75/flags.json
user    18164  11.9  1.4 413686944 468480   ??  Ss   Tue08AM 262:35.56 /Users/user/.hypersdk/avalanchego-d729e5c7ef9f008c3e89cd7131148ad3acda2e34/avalanchego --config-file /Users/user/.tmpnet/networks/20240903-083857.802339-nuklaivm-e2e-tests/NodeID-KDrJ72L2Uvc2sgxsg22T4CanD2bTdNyqD/flags.json
user    18163   5.8  1.8 414329136 588672   ??  S    Tue08AM  59:58.94 /Users/user/.hypersdk/avalanchego-d729e5c7ef9f008c3e89cd7131148ad3acda2e34/plugins/qCNyZHrs3rZX458wPJXPJJypPf6w423A84jnfbdP2TPEmEE9u
user    18174   1.2  1.8 414329120 590544   ??  S    Tue08AM  59:20.53 /Users/user/.hypersdk/avalanchego-d729e5c7ef9f008c3e89cd7131148ad3acda2e34/plugins/qCNyZHrs3rZX458wPJXPJJypPf6w423A84jnfbdP2TPEmEE9u
user    33458   1.1  0.2 411703840  76336   ??  Ss   Fri05PM  55:42.04 /Users/user/.hypersdk/avalanchego-d729e5c7ef9f008c3e89cd7131148ad3acda2e34/avalanchego --config-file /Users/user/.tmpnet/networks/20240830-171145.374858-nuklaivm-e2e-tests/NodeID-PC24PqjXQFN81PA6a21msrhqtd6Axkvre/flags.json
user    22880   0.0  0.0 410059824     48 s001  R+    2:42PM   0:00.00 grep avalanchego
```

The following command will clean up to ensure that you can start the network:

```bash
killlall avalanchego
```

### Build `nuklai-cli`

To make it easy to interact with the `nuklaivm`, we implemented the `nuklai-cli`.
Next, you'll need to build this tool. You can use the following command:

```bash
./scripts/build.sh
```

_This command will put the compiled CLI in `./build/nuklai-cli`._

### Configure `nuklai-cli`

Next, you'll need to add the chains you created and the default key to the
`nuklai-cli`. You can use the following commands from this location to do so:

```bash
./build/nuklai-cli key import ed25519 demo.pk
```

If the key is added correctly, you'll see the following log:

```bash
database: .nuklai-cli
imported address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
```

Next, you'll need to store the URLs of the nodes running on your Subnet:

```bash
./build/nuklai-cli chain import
```

This will automatically import the uri with the value `http://127.0.0.1:9650/ext/bc/nuklaivm`.

If you want to import another chain with its uri, you can do the following:

```bash
./build/nuklai-cli chain import "http://127.0.0.1:41177/ext/bc/nuklaivm"
```

You can confirm the chain was imported correctly by running:

```bash
./build/morpheus-cli chain info
```

This should output something like the following:

```bash
available chains: 1
0) chainID: 2F1QmuxSSVntNHXnEevYHBZzyhsNGvAE5Y2pqJW2a4iBugTMWd
select chainID: 0 [auto-selected]
networkID: 88888 subnetID: MBor2t7Ahsr8hn1mW6QCNs5mYXxRnYM3KvkTjrg2m5MeHjB8o chainID: 2F1QmuxSSVntNHXnEevYHBZzyhsNGvAE5Y2pqJW2a4iBugTMWd
```

### Check Balance

To confirm you've done everything correctly up to this point, run the
following command to get the current balance of the key you added:

```bash
./build/nuklai-cli key balance
```

If successful, the balance response should look like this:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: PAroBtUb83kcU5m3XiD37D263cB8kaZsiZ1rG7DLKxs8EX7Cq
address: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb balance: 853000000.000000000 NAI
```

You can also check the balance of another address by passing in the address as the argument

```bash
./build/nuklai-cli key balance 01b27c7ce992cdb7ff039294d7901851902394bb85fa4f3dc4cbb960b07284b7f9
```

Should give output

```bash
chainID: PAroBtUb83kcU5m3XiD37D263cB8kaZsiZ1rG7DLKxs8EX7Cq
address: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb balance: 0.000000000 NAI
```

### Generate Another Address

Now that we have a balance to send, we need to generate another address to send to. Because
we use bech32 addresses, we can't just put a random string of characters as the recipient
(won't pass checksum test that protects users from sending to off-by-one addresses).

```bash
./build/nuklai-cli key generate secp256r1
```

Note that we are now generating a key with curve secp256r1 instead of ed25519 like our first account. We can do this because the vm supports three types of keys(ed25519, secp256r1 and bls)

If successful, the `nuklai-cli` will emit the new address:

```bash
created address: 011ddbf62f227dd32deea73b31945d65bb6676cccae6cf0b829dfc21b290387bac
Private Key String(Base64): J5kwBjyvMuvV4PfjTUOO8ZF40Db3KzhFidhH7ER+9Jg=
```

We can also generate a bls key doing

```bash
./build/nuklai-cli key generate bls
```

By default, the `nuklai-cli` sets newly generated addresses to be the default. We run
the following command to set it back to `demo.pk`:

```bash
./build/nuklai-cli key set
```

You should see something like this:

```bash
chainID: PAroBtUb83kcU5m3XiD37D263cB8kaZsiZ1rG7DLKxs8EX7Cq
stored keys: 3
chainID: PAroBtUb83kcU5m3XiD37D263cB8kaZsiZ1rG7DLKxs8EX7Cq
0) address: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb balance: 853000000.000000000 NAI
chainID: PAroBtUb83kcU5m3XiD37D263cB8kaZsiZ1rG7DLKxs8EX7Cq
1) address: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb balance: 0.000000000 NAI
chainID: PAroBtUb83kcU5m3XiD37D263cB8kaZsiZ1rG7DLKxs8EX7Cq
2) address: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb balance: 0.000000000 NAI
set default key: 0
```

### Generate vanity addresses with no private keys

There may be times when you just want to generate random nuklaivm addresses that have no associated private key. In theory, there could be a private key that corresponds to any given address, but the probability of such an occurrence is extremely low, especially when dealing with a sufficiently large random space.

However, to be certain that an address does not correspond to any private key, we can construct it in such a way that it falls outside the normal range of addresses generated from private keys.

The vanity address generation process involves creating random blockchain addresses and checking if they start with a specific prefix (e.g., "nuklai"). We generate cryptographically secure random IDs and construct addresses using these IDs. To speed up the search, the process runs in parallel across multiple CPU cores, with each worker testing a batch of addresses. The process continues until an address matching the desired prefix is found.

When you do

```bash
./build/nuklai-cli key generate-vanity-address nuklai
```

This will put "kiran" as the prefix and use that to generate a new vanity address.

You should see something like:

```bash
Using 24 workers to generate vanity address
Vanity Address: ffeb380a5e8b4ee882081e3f0f9b6378cbdb935245f6bd82df13246dae80e5f056
```

We are creating an address that includes the word "nuklaivmvanity" followed by random 19 bytes of data. This kind of address is highly unlikely to be generated from a private key because it does not follow the typical structure of addresses derived from private keys.

### Send Tokens

Lastly, we trigger the transfer:

```bash
./build/nuklai-cli action transfer
```

The `nuklai-cli` will emit the following logs when the transfer is successful:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: PAroBtUb83kcU5m3XiD37D263cB8kaZsiZ1rG7DLKxs8EX7Cq
assetAddress (use NAI for native token): NAI
address: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb balance: 853000000.000000000 NAI
✔ amount: 1█
continue (y/n): y
✅ txID: wjzqXJeYedVSyBWfapoGiHYC9EQr2HgkA7xrVW4KZQyP1jxxG
txID: wjzqXJeYedVSyBWfapoGiHYC9EQr2HgkA7xrVW4KZQyP1jxxG
fee consumed: 0.000048500 NAI
```

### Bonus: Watch Activity in Real-Time

To provide a better sense of what is actually happening on-chain, the
`nuklai-cli` comes bundled with a simple explorer that logs all blocks/txs that
occur on-chain. You can run this utility by running the following command from
this location:

```bash
./build/nuklai-cli chain watch
```

If you run it correctly, you'll see the following input (will run until the
network shuts down or you exit):

```bash
available chains: 1
0) chainID: 2F1QmuxSSVntNHXnEevYHBZzyhsNGvAE5Y2pqJW2a4iBugTMWd
select chainID: 0 [auto-selected]
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
watching for new blocks on 2F1QmuxSSVntNHXnEevYHBZzyhsNGvAE5Y2pqJW2a4iBugTMWd 👀
height:3003 txs:1 root:uNXBoJRGNo8JCJ8XDEioqVnJjwrVSSKBMgbaTd9AWFiUke2vE size:0.30KB units consumed: [bandwidth=224 compute=7 storage(read)=14 storage(allocate)=50 storage(write)=26] unit prices: [bandwidth=100 compute=100 storage(read)=100 storage(allocate)=100 storage(write)=100] [TPS:0.10 latency:66ms gap:142ms]
✅ 2deLZJJpfXm1Wrad1f8uZL1aZnrCkEtBZ6aFXNcR7stFYN8Rm8 actor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 summary (*actions.Transfer): [assetID: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb amount: 1000000000 -> 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb
]
```

## Demos

### Emission Balancer Demo

Refer to [Emission Balancer Demo](./docs/demos/emission_balancer.md) to learn how to retrieve info such as totalsupply, rewardsperblock, staking info, etc from Emission Balancer.

### Assets Demo

Refer to [Assets Demo](./docs/demos/tokens.md) to learn how to mint an asset, transfer it within the same subnet or to another subnet with AWM, etc.

### Datasets Demo

Refer to [Datasets Demo](./docs/demos/datasets.md) to learn how to create a dataset, update it, add data to dataset, etc.

### Nuklai Marketplace Demo

Refer to [Marketplace Demo](./docs/demos/marketplace.md) to learn how to create a publish your dataset up for sale on the Nuklai Marketplace and how to subscribe to a dataset.

## Faucet

You can run the faucet by doing:

```bash
FAUCET_PRIVATE_KEY_HEX="323b1d8f4eed5f0da9da93071b034f2dce9d2d22692c172f3cb252a64ddfafd01b057de320297c29ad0c1f589ea216869cf1938d88c9fbd70d6748323dbf2fa7" RPC_ENDPOINT="http://127.0.0.1:9650" go run ./cmd/faucet
```
