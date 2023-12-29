# nuklaivm

<p align="center">
  <!-- <a href="https://goreportcard.com/report/github.com/Nuklai/nuklaivm"><img src="https://goreportcard.com/badge/github.com/Nuklai/nuklaivm" /></a> -->
  <a href="https://github.com/Nuklai/nuklaivm/actions/workflows/unit-tests.yml"><img src="https://github.com/Nuklai/nuklaivm/actions/workflows/unit-tests.yml/badge.svg" /></a>
  <a href="https://github.com/Nuklai/nuklaivm/actions/workflows/static-analysis.yml"><img src="https://github.com/Nuklai/nuklaivm/actions/workflows/static-analysis.yml/badge.svg" /></a>
</p>

## Status

`nuklaivm` is considered **ALPHA** software and is not safe to use in
production. The framework is under active development and may change
significantly over the coming months as its modules are optimized and
audited.

## Demo

### Launch Subnet

The first step to running this demo is to launch your own `nuklaivm` Subnet. You
can do so by running the following command from this location (may take a few
minutes):

```bash
./scripts/run.sh;
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
stored chainID: 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh uri: http://127.0.0.1:38583/ext/bc/29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
stored chainID: 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh uri: http://127.0.0.1:45605/ext/bc/29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
stored chainID: 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh uri: http://127.0.0.1:38453/ext/bc/29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
stored chainID: 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh uri: http://127.0.0.1:46509/ext/bc/29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
stored chainID: 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh uri: http://127.0.0.1:35183/ext/bc/29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
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
chainID: 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
uri: http://127.0.0.1:45605/ext/bc/29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
balance: 853000000.000000000 NAI
```

### Generate Another Address

Now that we have a balance to send, we need to generate another address to send to. Because
we use bech32 addresses, we can't just put a random string of characters as the recipient
(won't pass checksum test that protects users from sending to off-by-one addresses).

```bash
./build/nuklai-cli key generate secp256r1
```

Note that we are now generating a key with curve secp256r1 instead of ed25519 like our first account. We can do this because the vm supports both types of keys

If successful, the `nuklai-cli` will emit the new address:

```
database: .nuklai-cli
created address: nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2
```

By default, the `nuklai-cli` sets newly generated addresses to be the default. We run
the following command to set it back to `demo.pk`:

```bash
./build/nuklai-cli key set
```

You should see something like this:

```
database: .nuklai-cli
chainID: Abcauv672Yv5xju7B4iuTsg9yMGVsTqxQYBYemDv8JLeHMWrb
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 853000000.000000000 NAI
1) address (secp256r1): nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2 balance: 0.000000000 NAI
✔ set default key: 0█
```

### Send Tokens

Lastly, we trigger the transfer:

```bash
./build/nuklai-cli action transfer
```

The `nuklai-cli` will emit the following logs when the transfer is successful:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
balance: 853000000.000000000 NAI
recipient: nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2
amount: 100
continue (y/n): y
✅ txID: sqLP8ZJk1BFLXAexkAx4fAmEbg4gvxJYseT99UW1fHsFvm4QL
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
select chainID: 0 [auto-selected]
uri: http://127.0.0.1:45605/ext/bc/29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh
watching for new blocks on 29LTS1j68jsRDRkDazY4bRw9MQTdpmjb3LVG6Lx7voUqDPQPgh 👀
height:53 txs:0 root:2TRxW2hWUP2DTKgGTPkJLTzzzkCxDESqjriHhdYU7PqTPWEaR5 size:0.09KB units consumed: [bandwidth=0 compute=0 storage(read)=0 storage(allocate)=0 storage(write)=0] unit prices: [bandwidth=100 compute=100 storage(read)=100 storage(allocate)=100 storage(write)=100]
✅ sqLP8ZJk1BFLXAexkAx4fAmEbg4gvxJYseT99UW1fHsFvm4QL actor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx summary (*actions.Transfer): [100.000000000 NAI -> nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2] fee (max 72.34%): 0.000023800 NAI consumed: [bandwidth=191 compute=7 storage(read)=14 storage(allocate)=0 storage(write)=26]
```
