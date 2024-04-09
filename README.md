# nuklaivm

<p align="center">
  <!-- <a href="https://goreportcard.com/report/github.com/Nuklai/nuklaivm"><img src="https://goreportcard.com/badge/github.com/Nuklai/nuklaivm" /></a> -->
  <a href="https://github.com/Nuklai/nuklaivm/actions/workflows/unit-tests.yml"><img src="https://github.com/Nuklai/nuklaivm/actions/workflows/unit-tests.yml/badge.svg" /></a>
  <a href="https://github.com/Nuklai/nuklaivm/actions/workflows/static-analysis.yml"><img src="https://github.com/Nuklai/nuklaivm/actions/workflows/static-analysis.yml/badge.svg" /></a>
</p>

---

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

- ☑ Transfer both the native asset `NAI` and any other token created by users within the same subnet
- ☑ Transfer both the native asset `NAI` and any other token created by users to another subnet using Avalanche Warp Messaging(AWM)
- ☑ Create a token
- ☑ Mint a token
- ☑ Burn a token
- ☑ Export both the native asset `NAI` and any other user tokens to another subnet that is also a `nuklaivm`
- ☑ Import both the native asset `NAI` and any other user tokens from another subnet that is also a `nuklaivm`
- ☑ Register validator for staking
- ☑ Unregister validator from staking
- ☑ Delegate NAI to any currently staked validator
- ☑ Undelegate NAI from a staked validator
- ☑ Claim Validator staking rewards
- ☑ Claim User delegation rewards

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

#### Avalanche Warp Support

We take advantage of the Avalanche Warp Messaging (AWM) support provided by the
`hypersdk` to enable any `nuklaivm` to send assets to any other `nuklaivm` without
relying on a trusted relayer or bridge (just the validators of the `nuklaivm`
sending the message).

By default, a `nuklaivm` will accept a message from another `nuklaivm` if 80% of
the stake weight of the source has signed it. Because each imported asset is
given a unique `AssetID` (hash of `sourceChainID + sourceAssetID`), it is not
possible for a malicious/rogue Subnet to corrupt token balances imported from
other Subnets with this default import setting. `nuklaivms` also track the
amount of assets exported to all other `nuklaivms` and ensure that more assets
can't be brought back from a `nuklaivm` than were exported to it (prevents
infinite minting).

To limit "contagion" in the case of a `nuklaivm` failure, we ONLY allow the
export of natively minted assets to another `nuklaivm`. This means you can
transfer an asset between two `nuklaivms` A and B but you can't export from
`nuklaivm` A to `nuklaivm` B to `nuklaivm` C. This ensures that the import policy
for an external `nuklaivm` is always transparent and is never inherited
implicitly by the transfers between other `nuklaivms`. The ability to impose
this restriction (without massively driving up the cost of each transfer) is
possible because AWM does not impose an additional overhead per Subnet
connection (no "per connection" state to maintain). This means it is just as
cheap/scalable to communicate with every other `nuklaivm` as it is to only
communicate with one.

Lastly, the `nuklaivm` allows users to both tip relayers (whoever sends
a transaction that imports their message) and to swap for another asset when
their message is imported (so they can acquire fee-paying tokens right when
they arrive).

You can see how this works by checking out the [E2E test suite](./tests/e2e/e2e_test.go)
that runs through these flows.

#### Emission Balancer and Staking Mechanism

On `nuklaivm`, the emission balancer handles the staking mechanism whereby it tracks the
total supply of `NAI`, max supply of `NAI`, staking rewards per block and the emission address to direct 50% of all fees to.
Furthermore, it also rewards all the validators that have a minimum stake and all the users who have a minimum delegated stake to a validator of their choice.

Read more about [Emission Balancer](./docs/emission_balancer/README.md).

## Demo

### Launch Subnet

The first step to running this demo is to launch your own `nuklaivm` Subnet. You
can do so by running the following command from this location (may take a few
minutes):

```bash
./scripts/run.sh;
```

By default, it will store all the subnet files under `/tmp/nuklaivm`

Or you can also pass in the root directory for where all the subnet files will be stored by doing:

```bash
./scripts/run.sh /tmp/my-dir
```

You can also choose to skip tests by doing

```bash
./scripts/run.sh --ginkgo.skip
```

When the Subnet is running, you'll see the following logs emitted:

```
cluster is ready!
avalanche-network-runner is running in the background...

use the following command to terminate:

./scripts/stop.sh;
```

_By default, this allocates all funds on the network to `created address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx`. The private
key for this address is `0x323b1d8f4eed5f0da9da93071b034f2dce9d2d22692c172f3cb252a64ddfafd01b057de320297c29ad0c1f589ea216869cf1938d88c9fbd70d6748323dbf2fa7`.
For convenience, this key has is also stored at `demo.pk`._

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

```
database: .nuklai-cli
imported address: created address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
```

Next, you'll need to store the URLs of the nodes running on your Subnet:

```bash
./build/nuklai-cli chain import-anr
```

If `nuklai-cli` is able to connect to ANR, it will emit the following logs:

```
database: .nuklai-cli
stored chainID: QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD uri: http://127.0.0.1:36133/ext/bc/QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD
stored chainID: QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD uri: http://127.0.0.1:46813/ext/bc/QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD
stored chainID: QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD uri: http://127.0.0.1:33441/ext/bc/QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD
stored chainID: QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD uri: http://127.0.0.1:33917/ext/bc/QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD
stored chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX uri: http://127.0.0.1:39387/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
stored chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX uri: http://127.0.0.1:42111/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
stored chainID: QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD uri: http://127.0.0.1:37135/ext/bc/QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD
stored chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX uri: http://127.0.0.1:44089/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
stored chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX uri: http://127.0.0.1:36167/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
stored chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX uri: http://127.0.0.1:38297/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
```

_`./build/nuklai-cli chain import-anr` connects to the Avalanche Network Runner server running in
the background and pulls the URIs of all nodes tracking each chain you
created._

### Check Balance

To confirm you've done everything correctly up to this point, run the
following command to get the current balance of the key you added:

```bash
./build/nuklai-cli key balance
```

If successful, the balance response should look like this:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
assetID (use NAI for native token): NAI
✔ assetID (use NAI for native token): NAI█
balance: 853000000.000000000 NAI
```

You can also check the balance of another address by passing in the address as the argument

```bash
./build/nuklai-cli key balance nuklai1de6kkmrpd9mx6anpde5hg72rv2u288wgh6rsrjmup0uzjamjskakqqjyt9u
```

Should give output

```
database: .nuklai-cli
address: nuklai1de6kkmrpd9mx6anpde5hg72rv2u288wgh6rsrjmup0uzjamjskakqqjyt9u
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
assetID (use NAI for native token): NAI
balance: 0.000000000 NAI
please send funds to nuklai1de6kkmrpd9mx6anpde5hg72rv2u288wgh6rsrjmup0uzjamjskakqqjyt9u
exiting...
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

```
database: .nuklai-cli
created address: nuklai1qyfl0vx359gtn7vhj52xl9gp3pdphttqgtat9d8dad55c3rd4cnzgl3ma69
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

```
database: .nuklai-cli
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 853000000.000000000 NAI
1) address (secp256r1): nuklai1qyfl0vx359gtn7vhj52xl9gp3pdphttqgtat9d8dad55c3rd4cnzgl3ma69 balance: 0.000000000 NAI
2) address (bls): nuklai1qgcty39f7z5q86yy7x8w0cnyuht34ut38wp2n5qp0zmyhglr5fjm5hqkejn balance: 0.000000000 NAI
✔ set default key: 0█
```

### Generate vanity addresses with no private keys

There may be times when you just want to generate random nuklaivm addresses that have no associated private key. In theory, there could be a private key that corresponds to any given address, but the probability of such an occurrence is extremely low, especially when dealing with a sufficiently large random space.
However, to be certain that an address does not correspond to any private key, we can construct it in such a way that it falls outside the normal range of addresses generated from private keys. One common approach is to use a clearly invalid or special pattern that cannot be derived from a private key under the normal address generation rules of our blockchain.

When you do

```bash
./build/nuklai-cli key generate-vanity-address
```

You should see something like:

```
database: .nuklai-cli
Address: nuklai1de6kkmrpd9mx6anpde5hg7gxgsjtdrhj6fe8yctv4npqchrey7sfspeyejq
```

We are creating an address that includes the word "nuklaivmvanity" followed by random 19 bytes of data. This kind of address is highly unlikely to be generated from a private key because it does not follow the typical structure of addresses derived from private keys.

### Send Tokens

Lastly, we trigger the transfer:

```bash
./build/nuklai-cli action transfer
```

The `nuklai-cli` will emit the following logs when the transfer is successful:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
assetID (use NAI for native token): NAI
balance: 853000000.000000000 NAI
✔ recipient: nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2
amount: 100
continue (y/n): y
✅ txID: pPmBtqtjpu4eTmLeBZtWLgSMvUf8Y85cZrdvsVn3vUtv24dzY
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

```
database: .nuklai-cli
available chains: 2 excluded: []
0) chainID: QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD
1) chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
select chainID: 1
uri: http://127.0.0.1:44089/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
watching for new blocks on 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX 👀
height:292 txs:1 root:RXYUPPSGL2zJiiR6r3ozakfyhEkPbiXkCHnm26H5UvMT23Sy5 size:0.31KB units consumed: [bandwidth=227 compute=7 storage(read)=12 storage(allocate)=25 storage(write)=26] unit prices: [bandwidth=100 compute=100 storage(read)=100 storage(allocate)=100 storage(write)=100] [TPS:4.30 latency:29273ms gap:0ms]
✅ pPmBtqtjpu4eTmLeBZtWLgSMvUf8Y85cZrdvsVn3vUtv24dzY actor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx summary (*actions.Transfer): [100.000000000 NAI -> nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2] fee (max 81.37%): 0.000029700 NAI consumed: [bandwidth=227 compute=7 storage(read)=12 storage(allocate)=25 storage(write)=26]
```

### Emission Balancer Demo

Refer to [Emission Balancer Demo](./docs/demos/emission_balancer.md) to learn how to retrieve info such as totalsupply, rewardsperblock, staking info, etc from Emission Balancer.

### Tokens Demo

Refer to [Tokens Demo](./docs/demos/tokens.md) to learn how to mint an asset, transfer it within the same subnet or to another subnet with AWM, etc.

## Nuklai Wallet

We have a native wallet for Nuklai network. This wallet integrates the following:

- Local Wallet
- Blockchain Explorer
- Faucet that you can use to request some test NAI to your local wallet
- Mint page where you can create and mint tokens to any address
- Feed page which is like a mini blog that lets you post a message and the url to the blockchain within a transaction memo

To run the wallet, just do:

```bash
cd cmd/nuklai-wallet;
./scripts/dev.sh
```

Note that when you run `dev.sh`, it does the following:

- Start Nuklai Feed from cmd/nuklai-feed
- Start Nuklai Faucet from cmd/nuklai-faucet
- Run Nuklai Wallet app

## Tests

### Running an Integration Test

You can run the integration tests by doing

```bash
./scripts/tests.integration.sh
```

### Running a Load Test

_Before running this demo, make sure to stop the network you started using
`killall avalanche-network-runner`._

The `nuklaivm` load test will provision 5 `nuklaivms` and process 500k transfers
on each between 10k different accounts.

```bash
./scripts/tests.load.sh
```

_This test SOLELY tests the speed of the `nuklaivm`. It does not include any
network delay or consensus overhead. It just tests the underlying performance
of the `hypersdk` and the storage engine used (in this case MerkleDB on top of
Pebble)._

### Measuring Disk Speed

This test is extremely sensitive to disk performance. When reporting any TPS
results, please include the output of:

```bash
./scripts/tests.disk.sh
```

_Run this test RARELY. It writes/reads many GBs from your disk and can fry an
SSD if you run it too often. We run this in CI to standardize the result of all
load tests._

## Zipkin Tracing

To trace the performance of `nuklaivm` during load testing, we use `OpenTelemetry + Zipkin`.

To get started, startup the `Zipkin` backend and `ElasticSearch` database (inside `hypersdk/trace`):

```bash
docker-compose -f trace/zipkin.yml up
```

Once `Zipkin` is running, you can visit it at `http://localhost:9411`.

Next, startup the load tester (it will automatically send traces to `Zipkin`):

```bash
TRACE=true ./scripts/tests.load.sh
```

When you are done, you can tear everything down by running the following
command:

```bash
docker-compose -f trace/zipkin.yml down
```

## Creating a Devnet

Refer to [Creating Devnet Demo](./docs/demos/devnet.md) to learn how to create a private devnet (running on a
custom Primary Network with traffic scoped to the deployer IP) across any number of regions and nodes
in ~30 minutes with a single script.
