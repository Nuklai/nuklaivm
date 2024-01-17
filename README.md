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

Note the working directory for run.sh is `/data/github/tmp/nuklaivm` so you will need to create it first and give it appropriate permissions.

```bash
sudo mkdir -p /data/github/tmp/nuklaivm;
sudo chown 777 /data/github/tmp/nuklaivm
```

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

You can also check the balance of another address by passing in the address as the argument

```bash
./build/nuklai-cli key balance nuklai1de6kkmrpd9mx6anpde5hg72rv2u288wgh6rsrjmup0uzjamjskakqqjyt9u
```

Should give output

```
database: .nuklai-cli
address: nuklai1de6kkmrpd9mx6anpde5hg72rv2u288wgh6rsrjmup0uzjamjskakqqjyt9u
chainID: 2fc6PDKZ9Dvhcb1vaxbWUoUyvoEp14nbk6NHWhbzfr4HEkTVWi
balance: 0 NAI
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
Address: nuklai1de6kkmrpd9mx6anpde5hg7f6t7dvn7u3dxfgfsz8yh3fggguqqf3wdd7xfy
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

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2fc6PDKZ9Dvhcb1vaxbWUoUyvoEp14nbk6NHWhbzfr4HEkTVWi
emission info:
TotalSupply=853000000000000000 MaxSupply=10000000000000000000 RewardsPerBlock=2000000000
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2rQM1CBDWGtBcdgx8XnAvoDvQkfvSW7Q7XaD58aosAKS4wHMup
validator 0: NodeID=NodeID-6Eey6vF4nhR6w9PVLtjekPtuF2ayWE5zJ NodePublicKey=tBr6uQRcQkBnVaIxqrac22vIJPhlaUQ7teHbF8xSk8KbGZW5s1LYH82Y/L43ejG3 UserStake=map[] StakedAmount=0 StakedReward=0
validator 1: NodeID=NodeID-DgK4r1LPjqSouCzzUixyKB5714jiZRF1k NodePublicKey=iPcS9KgoLzMrXi1emUfCwsEIWbACYZ1Hx3/A/+4SzRrzTSwdbM7KA2x80RqFcV3/ UserStake=map[] StakedAmount=0 StakedReward=0
validator 2: NodeID=NodeID-Fw4yPxf5WCPrVt4QH1pNfHy92uqWF8UC2 NodePublicKey=qiAeDWL37pJfDBuERd7tX52PjCs+kRwMKuULKKafjYaPRw1VKCo5ZoRu7sRVNAsa UserStake=map[] StakedAmount=0 StakedReward=0
validator 3: NodeID=NodeID-6sPngJ34vevxBRd4Nj9jCCGPPWYA5Boni NodePublicKey=tBNgtWRhWGcMXP7WycOT2M7hsqyRa7uXG2FtF8UMgaHPG9UCKVKO+uaftiHXf7ki UserStake=map[] StakedAmount=0 StakedReward=0
validator 4: NodeID=NodeID-9GDrz1zozSus8oSmQfYBzcSkL95tCAQcd NodePublicKey=qZV7xwq5zCG/J8hSf1rK/M2YKvVYaLp/acJGl2D6VxXMm+l+lpGPWgpN8X2yD7Q0 UserStake=map[] StakedAmount=0 StakedReward=0
```

### Stake to a validator

We can stake to a validator of our choice

```bash
./build/nuklai-cli action stake-validator
```

If successful, the output should be:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
validators: 5
0: NodeID=NodeID-wGVvYo7jBvtmnfUTaay2vcL8j8GJyokb NodePublicKey=rnp1CCGFvbni4bjFRFCJo7b3SxdnBIJ8qzOPPbB6HPR8VK8hvHaO37lZGNLJs30S
1: NodeID=NodeID-NuCwBadeuYbzntFJgBAS8Ut9pwo52XrT6 NodePublicKey=syBPqN0eU9nCBJDyFOVynneq/nia8lM0apG/DpboYtc7CJdm0hXlKGNZF5fwyjWp
2: NodeID=NodeID-3yxghtfwRdYcG69FjoxZrwjUkSXJAGhY9 NodePublicKey=kvFhrcEVW5Ooann3NaqqE2nANL/XS86AnCUFgrdyBQa2z+xAlCFcwuPHPDnvHyZp
3: NodeID=NodeID-6dvn9WTA4i7qG2pT3GKUXP46xa2SVY7Po NodePublicKey=oXMYzibvB7gHaGVAKVEB5z0+IFEPcWb0TxjrIz26p3eVjmaHkmKK41S64HDg8paD
4: NodeID=NodeID-423bGHFH5exxQfuNiRFUqxDquWD9svj6E NodePublicKey=rC9RaeHAUs4mSMw4YoAKBQaWecfrkLHEgKSq/JnfU2EnTXHjYuZu94aDQSTh1M7b
validator to stake to: 3
balance: 852999899.999972820 NAI
✔ Staked amount: 100█
End LockUp Height: 70
✔ continue (y/n): y█
✅ txID: EYnnHR9jtJvfAAE9UE5tkV9BDvajBhZ25YqJUD82LrRsJrZTo
```

### Get user staking info

We can retrieve our staking info by passing in which validator we have staked to and the address to look up staking for using the new RPC API we defined as part of this exercise.

```bash
./build/nuklai-cli emission user-stake-info
```

If successful, the output should be:

```
database: .nuklai-cli
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
validators: 5
0: NodeID=NodeID-NuCwBadeuYbzntFJgBAS8Ut9pwo52XrT6 NodePublicKey=syBPqN0eU9nCBJDyFOVynneq/nia8lM0apG/DpboYtc7CJdm0hXlKGNZF5fwyjWp
1: NodeID=NodeID-3yxghtfwRdYcG69FjoxZrwjUkSXJAGhY9 NodePublicKey=kvFhrcEVW5Ooann3NaqqE2nANL/XS86AnCUFgrdyBQa2z+xAlCFcwuPHPDnvHyZp
2: NodeID=NodeID-6dvn9WTA4i7qG2pT3GKUXP46xa2SVY7Po NodePublicKey=oXMYzibvB7gHaGVAKVEB5z0+IFEPcWb0TxjrIz26p3eVjmaHkmKK41S64HDg8paD
3: NodeID=NodeID-423bGHFH5exxQfuNiRFUqxDquWD9svj6E NodePublicKey=rC9RaeHAUs4mSMw4YoAKBQaWecfrkLHEgKSq/JnfU2EnTXHjYuZu94aDQSTh1M7b
4: NodeID=NodeID-wGVvYo7jBvtmnfUTaay2vcL8j8GJyokb NodePublicKey=rnp1CCGFvbni4bjFRFCJo7b3SxdnBIJ8qzOPPbB6HPR8VK8hvHaO37lZGNLJs30S
✔ choose validator whom you have staked to: 2█
address to get staking info for: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
user stake:  Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx StakedAmount=100000000000
stake #1: TxID=EYnnHR9jtJvfAAE9UE5tkV9BDvajBhZ25YqJUD82LrRsJrZTo Amount=100000000000 StartLockUp=53
```

### Unstake from a validator

Let's first check the balance on our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
uri: http://127.0.0.1:41743/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
balance: 852999799.999945203 NAI
```

We can unstake specific stake from a chosen validator.

```bash
./build/nuklai-cli action unstake-validator
```

Which produces result:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
validators: 5
0: NodeID=NodeID-wGVvYo7jBvtmnfUTaay2vcL8j8GJyokb NodePublicKey=rnp1CCGFvbni4bjFRFCJo7b3SxdnBIJ8qzOPPbB6HPR8VK8hvHaO37lZGNLJs30S
1: NodeID=NodeID-NuCwBadeuYbzntFJgBAS8Ut9pwo52XrT6 NodePublicKey=syBPqN0eU9nCBJDyFOVynneq/nia8lM0apG/DpboYtc7CJdm0hXlKGNZF5fwyjWp
2: NodeID=NodeID-3yxghtfwRdYcG69FjoxZrwjUkSXJAGhY9 NodePublicKey=kvFhrcEVW5Ooann3NaqqE2nANL/XS86AnCUFgrdyBQa2z+xAlCFcwuPHPDnvHyZp
3: NodeID=NodeID-6dvn9WTA4i7qG2pT3GKUXP46xa2SVY7Po NodePublicKey=oXMYzibvB7gHaGVAKVEB5z0+IFEPcWb0TxjrIz26p3eVjmaHkmKK41S64HDg8paD
4: NodeID=NodeID-423bGHFH5exxQfuNiRFUqxDquWD9svj6E NodePublicKey=rC9RaeHAUs4mSMw4YoAKBQaWecfrkLHEgKSq/JnfU2EnTXHjYuZu94aDQSTh1M7b
validator to unstake from: 3
stake info:
0: TxID=EYnnHR9jtJvfAAE9UE5tkV9BDvajBhZ25YqJUD82LrRsJrZTo StakedAmount=100000000000 StartLockUpHeight=53 CurrentHeight=200
stake ID to unstake: 0 [auto-selected]
continue (y/n): y
✅ txID: 2eSkTRQa4KqHDXidoeoQ8XSsjeSbga3x5B52hetwhGw68enbHt
```

Now, if we check the balance again, we should have our 100 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
uri: http://127.0.0.1:41743/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
balance: 852999899.999896407 NAI
```
