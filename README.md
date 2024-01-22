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
stored chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob uri: http://127.0.0.1:32913/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
stored chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob uri: http://127.0.0.1:43531/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
stored chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob uri: http://127.0.0.1:45469/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
stored chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob uri: http://127.0.0.1:38145/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
stored chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob uri: http://127.0.0.1:45849/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
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
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
uri: http://127.0.0.1:45849/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
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
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
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

Note that we are now generating a key with curve secp256r1 instead of ed25519 like our first account. We can do this because the vm supports three types of keys(ed25519, secp256r1 and bls)

If successful, the `nuklai-cli` will emit the new address:

```
database: .nuklai-cli
created address: nuklai1q9zvhxdftggsdjmpskrg75dd9xyvmxpwfgy0desjuk5dlz78a0r7ueax5f5
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
1) address (secp256r1): nuklai1q9zvhxdftggsdjmpskrg75dd9xyvmxpwfgy0desjuk5dlz78a0r7ueax5f5 balance: 0.000000000 NAI
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
Address: nuklai1de6kkmrpd9mx6anpde5hg7faxphvw7uz79f93uftnnq7wyeudknpjav4env
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
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
balance: 853000000.000000000 NAI
✔ recipient: nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2█
amount: 100
continue (y/n): y
✅ txID: 2F5m7Bmy5P4CN6DZBsK7pfDVwGLKCe1cqc2MPGsGnY9Mu9d3qH
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
available chains: 1 excluded: []
0) chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
select chainID: 0 [auto-selected]
uri: http://127.0.0.1:45849/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
watching for new blocks on 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob 👀
height:114 txs:0 root:21Z8mBjjifsx1kDqcDugQm8fxRczvDQmdYG75aezwbnoiya8eG size:0.09KB units consumed: [bandwidth=0 compute=0 storage(read)=0 storage(allocate)=0 storage(write)=0] unit prices: [bandwidth=100 compute=100 storage(read)=100 storage(allocate)=100 storage(write)=100]
✅ 2F5m7Bmy5P4CN6DZBsK7pfDVwGLKCe1cqc2MPGsGnY9Mu9d3qH actor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx summary (*actions.Transfer): [100.000000000 NAI -> nuklai1qyf889stx7rjrgh8tsa4acv4we94kf4w652gwq462tm4vau9ee20gq6k5l2] fee (max 72.34%): 0.000023800 NAI consumed: [bandwidth=191 compute=7 storage(read)=14 storage(allocate)=0 storage(write)=26]
```

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
emission info:
TotalSupply=853000282000000000 MaxSupply=10000000000000000000 RewardsPerBlock=2000000000
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
validator 0: NodeID=NodeID-GpaAwHcn4vjqjN2kWDSWpv6KZ3K7XkfpB NodePublicKey=sG4M9T1zxgXc2UvlZPjCGdijCMcf+TaKKebQZLe2ZUNTQvy6ZG4AadZKtacipoOP UserStake=map[] StakedAmount=0 StakedReward=0
validator 1: NodeID=NodeID-NC1hkGBb1iFx7cu5vmkw7cMUYBgo6R4YY NodePublicKey=gvUIBc3RrlKyia9SD8AP/tLEHKhoK29upLyCzzNUQtCpARrPrCFTD/hCrFgL2H5K UserStake=map[] StakedAmount=0 StakedReward=0
validator 2: NodeID=NodeID-Ph81A5DgtyC7Q1fGhxVW3b94zXDcJJngC NodePublicKey=gXKx9IpvEzyDDjHSwEh259XrmL2Vv8JA4nupfegu3fG7F3YDAeGh+sAAsoNGTabA UserStake=map[] StakedAmount=0 StakedReward=0
validator 3: NodeID=NodeID-Ji9mis6M4D3kNsEv6iGa2VqkMTbruMxH6 NodePublicKey=mO/+NoK/32r4SV4erRl5sw2VeEj+PDfL6ONeDC12haCroLPf6RdhVMNvvS2BImTl UserStake=map[] StakedAmount=0 StakedReward=0
validator 4: NodeID=NodeID-KSTcHsTmUvfSNKdeDBhaC3KyHLXVygNPK NodePublicKey=pkEMnmJZD6a1nKNsod3LCiRU3TBFk+9A3P3QIU/dlFpUba5pORy6tGdmlZaVAUEG UserStake=map[] StakedAmount=0 StakedReward=0
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
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
validators: 5
0: NodeID=NodeID-KSTcHsTmUvfSNKdeDBhaC3KyHLXVygNPK NodePublicKey=pkEMnmJZD6a1nKNsod3LCiRU3TBFk+9A3P3QIU/dlFpUba5pORy6tGdmlZaVAUEG
1: NodeID=NodeID-GpaAwHcn4vjqjN2kWDSWpv6KZ3K7XkfpB NodePublicKey=sG4M9T1zxgXc2UvlZPjCGdijCMcf+TaKKebQZLe2ZUNTQvy6ZG4AadZKtacipoOP
2: NodeID=NodeID-NC1hkGBb1iFx7cu5vmkw7cMUYBgo6R4YY NodePublicKey=gvUIBc3RrlKyia9SD8AP/tLEHKhoK29upLyCzzNUQtCpARrPrCFTD/hCrFgL2H5K
3: NodeID=NodeID-Ph81A5DgtyC7Q1fGhxVW3b94zXDcJJngC NodePublicKey=gXKx9IpvEzyDDjHSwEh259XrmL2Vv8JA4nupfegu3fG7F3YDAeGh+sAAsoNGTabA
4: NodeID=NodeID-Ji9mis6M4D3kNsEv6iGa2VqkMTbruMxH6 NodePublicKey=mO/+NoK/32r4SV4erRl5sw2VeEj+PDfL6ONeDC12haCroLPf6RdhVMNvvS2BImTl
✔ validator to stake to: 0█
balance: 852999899.999973893 NAI
✔ End LockUp Height: 200█
✔ continue (y/n): y█
✅ txID: FTyDSapHcfd58osSk89nLfhMbYQmTGpLLS2bn6n4hzEsaCEFD
```

### Get user staking info

We can retrieve our staking info by passing in which validator we have staked to and the address to look up staking for

```bash
./build/nuklai-cli emission user-stake-info
```

If successful, the output should be:

```
database: .nuklai-cli
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
validators: 5
0: NodeID=NodeID-Ph81A5DgtyC7Q1fGhxVW3b94zXDcJJngC NodePublicKey=gXKx9IpvEzyDDjHSwEh259XrmL2Vv8JA4nupfegu3fG7F3YDAeGh+sAAsoNGTabA
1: NodeID=NodeID-Ji9mis6M4D3kNsEv6iGa2VqkMTbruMxH6 NodePublicKey=mO/+NoK/32r4SV4erRl5sw2VeEj+PDfL6ONeDC12haCroLPf6RdhVMNvvS2BImTl
2: NodeID=NodeID-KSTcHsTmUvfSNKdeDBhaC3KyHLXVygNPK NodePublicKey=pkEMnmJZD6a1nKNsod3LCiRU3TBFk+9A3P3QIU/dlFpUba5pORy6tGdmlZaVAUEG
3: NodeID=NodeID-GpaAwHcn4vjqjN2kWDSWpv6KZ3K7XkfpB NodePublicKey=sG4M9T1zxgXc2UvlZPjCGdijCMcf+TaKKebQZLe2ZUNTQvy6ZG4AadZKtacipoOP
4: NodeID=NodeID-NC1hkGBb1iFx7cu5vmkw7cMUYBgo6R4YY NodePublicKey=gvUIBc3RrlKyia9SD8AP/tLEHKhoK29upLyCzzNUQtCpARrPrCFTD/hCrFgL2H5K
✔ choose validator whom you have staked to: 2█
✔ address to get staking info for: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx█
user stake:  Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx StakedAmount=100000000000
stake #1: TxID=FTyDSapHcfd58osSk89nLfhMbYQmTGpLLS2bn6n4hzEsaCEFD Amount=100000000000 StartLockUp=176
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
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
uri: http://127.0.0.1:45849/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
balance: 852999799.999946713 NAI
```

We can unstake specific stake from a chosen validator.

```bash
./build/nuklai-cli action unstake-validator
```

Which produces result:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
validators: 5
0: NodeID=NodeID-Ph81A5DgtyC7Q1fGhxVW3b94zXDcJJngC NodePublicKey=gXKx9IpvEzyDDjHSwEh259XrmL2Vv8JA4nupfegu3fG7F3YDAeGh+sAAsoNGTabA
1: NodeID=NodeID-Ji9mis6M4D3kNsEv6iGa2VqkMTbruMxH6 NodePublicKey=mO/+NoK/32r4SV4erRl5sw2VeEj+PDfL6ONeDC12haCroLPf6RdhVMNvvS2BImTl
2: NodeID=NodeID-KSTcHsTmUvfSNKdeDBhaC3KyHLXVygNPK NodePublicKey=pkEMnmJZD6a1nKNsod3LCiRU3TBFk+9A3P3QIU/dlFpUba5pORy6tGdmlZaVAUEG
3: NodeID=NodeID-GpaAwHcn4vjqjN2kWDSWpv6KZ3K7XkfpB NodePublicKey=sG4M9T1zxgXc2UvlZPjCGdijCMcf+TaKKebQZLe2ZUNTQvy6ZG4AadZKtacipoOP
4: NodeID=NodeID-NC1hkGBb1iFx7cu5vmkw7cMUYBgo6R4YY NodePublicKey=gvUIBc3RrlKyia9SD8AP/tLEHKhoK29upLyCzzNUQtCpARrPrCFTD/hCrFgL2H5K
validator to unstake from: 2
stake info:
0: TxID=FTyDSapHcfd58osSk89nLfhMbYQmTGpLLS2bn6n4hzEsaCEFD StakedAmount=100000000000 StartLockUpHeight=176 CurrentHeight=208
stake ID to unstake: 0 [auto-selected]
✔ continue (y/n): y█
✅ txID: YcSbyqGCaGgPshuMDvMrzWQPrxv6xEB6dTi6MDPjx7QyrsdYa
```

Now, if we check the balance again, we should have our 100 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
uri: http://127.0.0.1:45849/ext/bc/2CvxouoGcBva3xxisHDmdQggad8bChYbMvu45oAXV7nXGK1Yob
balance: 852999899.999922037 NAI
```
