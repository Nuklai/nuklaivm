### Register a validator for staking

Even if there may be validators that are already taking part in the consensus of `nuklaivm` blocks, it doesn't mean they are automatically registered for the
staking mechanism. In order for a validator to register for staking on `nuklaivm`, they need to use the exact same account as they used while setting up the validator for
the Avalanche primary network. Validator Owners will need to stake a minimum of 100 NAI.

When you run the `run.sh` script, it runs the `tests/e2e/e2e_test.go` which in turn copies over the `signer.key` for all the auto-generated validator nodes. The reason we do
this is to make it easier to test the registration of the validator for staking on `nuklaivm` using `nuklai-cli`.

There are two ways of registering your validator for staking.

#### Automatic run

We can let everything be configured automatically which means it'll set the values for staking automatically such as for:

- stakeStartTime: Sets it to 2 minutes from now
- stakeEndTime: Sets it to 3 minutes from now
- delegationFeeRate: Sets it to 10%
- rewardAddress: Sets it to the transaction actor

The only thing we would need to do is send some NAI that will be used for staking.

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

What this does is it imports the `staking.key` file located at `/tmp/nuklaivm/nodes/node1-bls/signer.key` and then tries to use it to register `node1` for staking.

This is because in order for registration of validator staking to work, you will need to use the same account you used while registering the validator to the Avalanche primary network to prevent unauthorized users from registering someone else's validator node.

If you don't have enough NAI in your account, you will see something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
Loading private key for node1
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
balance: 0.000000000 NAI
please send funds to nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v
exiting...
Balance of validator signer: 0.000000000
 You need a minimum of 100 NAI to register a validator
```

So, all we need to do is send at least 100 NAI to `nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
Loading private key for node1
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
balance: 1000.000000000 NAI
Balance of validator signer: 1000.000000000
Loading validator signer key : nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v
address: nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
Validator Signer Address: nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v
Validator NodeID: NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
balance: 1000.000000000 NAI
✔ Staked amount: 100█
continue (y/n): y
Register Validator Stake Info - stakeStartTime: 2024-03-07 21:52:46 stakeEndTime: 2024-03-07 21:53:46 delegationFeeRate: 10 rewardAddress: nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v
✅ txID: 2qhwgQswvMhD4kPYZ1NRJ2KZSXsWRdyDGWgxHfy2kpc1n92hHY
```

#### Manual run

Here, we can be granular and set our own values for stakeStartTime, stakeEndTime, delegationFeeRate and rewardAddress.

First, let's import the key manually:

```bash
./build/nuklai-cli key import bls /tmp/nuklaivm/nodes/node2-bls/signer.key
```

Which should output:

```
database: .nuklai-cli
imported address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
```

Let's make sure we have enough balance to send

```bash
./build/nuklai-cli key balance
```

Which outputs:

```
database: .nuklai-cli
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:34091/ext/bc/b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
balance: 1000.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
Validator Signer Address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
Validator NodeID: NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
balance: 1000.000000000 NAI
✔ Staked amount: 100█
Staking Start Time(must be after 2024-03-07 21:52:48) [YYYY-MM-DD HH:MM:SS]: 2024-03-07 21:54:00
Staking End Time(must be after 2024-03-07 21:54:00) [YYYY-MM-DD HH:MM:SS]: 2024-03-07 21:55:00
✔ Delegation Fee Rate(must be over 2): 90█
✔ Reward Address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032█
✔ continue (y/n): y█
Register Validator Stake Info - stakeStartTime: 2024-03-07 21:54:00 stakeEndTime: 2024-03-07 21:55:00 delegationFeeRate: 90 rewardAddress: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
✅ txID: 2tHMLx5wz3YvqP6acMv4RS7szrze7jbZE4BwLtDfHcpkCyXhPT
```

### Get Validator stake info

You may want to check your validator staking info such as stake start time, stake end time, staked amount, delegation fee rate and reward address. To do so, you can do:

Let's check the validator staking info for node1 or `NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB` which we staked above.

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
✔ validator to get staking info for: 0█
validator stake:
StakeStartTime=2024-03-07 21:52:46 +0000 UTC StakeEndTime=2024-03-07 21:53:46 +0000 UTC StakedAmount=100000000000 DelegationFeeRate=10 RewardAddress=nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v OwnerAddress=nuklai1q200jft8exqtek35frn3ztnmnaehldcj9q7am5xx7fzd3wg60gfw28zx34v
```

You can also retrieve other useful info from Emission Balancer by doing

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validator 0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB PublicKey=isWM94nBcji75tg5HclJqtDXgsVnK4eCFh9VoLMBcRO36i1s1s8G8uqfMp4EQmd6 StakedAmount=100000000000 UnclaimedStakedReward=245381 DelegationFeeRate=0.100000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g PublicKey=rTABL5MYL/Y92PovojmBluEDpd1i/ippIB//f34avlgpC6GzD24u5XX+uYFIjoal StakedAmount=100000000000 UnclaimedStakedReward=60989 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

### Delegate stake to a validator

On `nuklaivm`, in addition to validators registering their nodes for staking, users can also delegate NAI for staking which means they get
to share the rewards sent to the validator. The delegation fee rate is set by the validator node when they register for staking so it is up to
the users to choose which validator to choose for staking purpose.

If a user chooses to delegate, they need to stake at least 25 NAI. A user can delegate to multiple validators at once.

To do so, you do:

```bash
./build/nuklai-cli action delegate-user-stake [auto | manual]
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
validator to delegate to: 0
balance: 852997999.999940634 NAI
✔ Staked amount: 1000000█
continue (y/n): y
Delegate User Stake Info - stakeStartTime: 2024-03-07 21:56:37 stakeEndTime: 2024-03-07 21:57:37 rewardAddress: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✅ txID: 2m3GusdVuKwYQDXFVxLQ59T86zpJALDpTnmsuUxWcCTg1WogQu
```

Let's delegate to another validator as well:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
1: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
validator to delegate to: 0
balance: 851997999.999908686 NAI
Staked amount: 1000000
continue (y/n): y
Delegate User Stake Info - stakeStartTime: 2024-03-07 21:57:06 stakeEndTime: 2024-03-07 21:58:06 rewardAddress: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✅ txID: KLMc4F6FzQRmRpmfvJEgDX5kGKtdwYfFxn6WUBfWBGPruAPJK
```

### Get Delegated User stake info

You may want to check your delegated staking info such as stake start time, stake end time, staked amount, and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
✔ validator to get staking info for: 0█
validator stake:
StakeStartTime=2024-03-07 21:56:37 +0000 UTC StakedAmount=1000000000000000 RewardAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx OwnerAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
```

Let's check our stake info for another validator as well:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
validator to get staking info for: 1
validator stake:
StakeStartTime=2024-03-07 21:57:06 +0000 UTC StakedAmount=1000000000000000 RewardAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx OwnerAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
```

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
emission info:
TotalSupply=853000002379090550 MaxSupply=10000000000000000000 TotalStaked=2000200000000000 RewardsPerEpoch=475694444 NumBlocksInEpoch=10 EmissionAddress=nuklai1qqmzlnnredketlj3cu20v56nt5ken6thchra7nylwcrmz77td654w2jmpt9 EmissionUnclaimedBalance=115300
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validator 0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB PublicKey=isWM94nBcji75tg5HclJqtDXgsVnK4eCFh9VoLMBcRO36i1s1s8G8uqfMp4EQmd6 StakedAmount=100000000000 UnclaimedStakedReward=1712790694 DelegationFeeRate=0.100000 DelegatedAmount=1000000000000000 UnclaimedDelegatedReward=190280167
validator 1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g PublicKey=rTABL5MYL/Y92PovojmBluEDpd1i/ippIB//f34avlgpC6GzD24u5XX+uYFIjoal StakedAmount=100000000000 UnclaimedStakedReward=166602413 DelegationFeeRate=0.900000 DelegatedAmount=1000000000000000 UnclaimedDelegatedReward=1498444670
```

We can also retrieve all the validators(both staked and unstaked):

```bash
./build/nuklai-cli emission all-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validator 0: NodeID=NodeID-seqhJjbzsCjK5m6N3hzYb8huyabCcE5e PublicKey=pwDrcTXbeVLeXBXXDnagKW4VyOx9iedGkxMa+1e+NaxKp3XIjTT9HWjFLL/fXHE5 StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g PublicKey=rTABL5MYL/Y92PovojmBluEDpd1i/ippIB//f34avlgpC6GzD24u5XX+uYFIjoal StakedAmount=100000000000 UnclaimedStakedReward=166602413 DelegationFeeRate=0.900000 DelegatedAmount=1000000000000000 UnclaimedDelegatedReward=1498444670
validator 2: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB PublicKey=isWM94nBcji75tg5HclJqtDXgsVnK4eCFh9VoLMBcRO36i1s1s8G8uqfMp4EQmd6 StakedAmount=100000000000 UnclaimedStakedReward=1712790694 DelegationFeeRate=0.100000 DelegatedAmount=1000000000000000 UnclaimedDelegatedReward=190280167
validator 3: NodeID=NodeID-jrgby12TYG4yKSd3v4ZfK4xDSk22zb9g PublicKey=rR2gtKX/tOdZO7oKWyK7PmFSMYfOn0FCCbP3oY0IUr7peL0n+OeioulonscnrA/O StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 4: NodeID=NodeID-H25NPpFPihNV5Lu8niWQdHbSjEoSeN665 PublicKey=jJ4A/mri60XRa3n6cOVH6o/MOLKMlPRaCU1DjORzphn5DSxtlFrS9d7VaP5SMz2a StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0UnclaimedDelegatedReward=0
```

### Claim staking reward

On NuklaiVM, you are able to claim your staking rewards at any point in time without undelegating the stake from a validator. Also, as a validator, you're able to also claim your validator staking rewards at any point in time without unstaking your validator node.

First, we need to make sure that we're on our account we used for delegating our stake.

Let's first check the balance on our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:34091/ext/bc/b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
balance: 850997999.999876857 NAIaddress: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:33759/ext/bc/UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
balance: 19999.999936200 NAI
```

Next, let's check out how many NAI delegator rewards have been accumulated with our validator we staked to:

```bash
./build/nuklai-cli emission staked-validators
```

Which gives an output:

```
database: .nuklai-cli
chainID: 2DXc2cEtPDGTy7JqaWLNV6Lmg8tD75KUnSSLaGyidEb2Gvi5ii
validator 0: NodeID=NodeID-3GJnWPtDnvR4g7PbsJFqURNcpZj5UZVgb PublicKey=tt7BhocrKiDUYtI9NJNCwNxw9zSM9xjECMzqtnDEatj8OQ03/DdrlQKRtcuAmi+C StakedAmount=100000000000 UnclaimedStakedReward=103569348  DelegationFeeRate=0.100000 DelegatedAmount=100000000000 UnclaimedDelegatedReward=11467376
```

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-user-stake-reward
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
validator to claim staking rewards from: 0
continue (y/n): y
✅ txID: 2F774LBvGq6YLu4DkK2BBQn46KmizwkZZ2U1MWRQXUdFB3gZPJ
```

Let's claim our accumulated rewards from another validator:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
validator to claim staking rewards from: 1
continue (y/n): y
✅ txID: 2vQMvbpa3sr3qYNva4NpQ6ARvkTRLNJssbvbHskckhHZiFiQS3
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:34091/ext/bc/b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
balance: 850998002.378300190 NAI
```

NOTE: The longer you stake, the more rewards you will get for delegating. Also, note that as soon as you claim your reward, this timestamp is recorded and your staking duration will be reset. This incentivizes users from not claiming their rewards for as long as they want to and they get rewarded for it.

### Undelegate user stake

Once your delegated stake period has ended, perhaps you may want to undelegate your stake(i.e. you want to get back your NAI you have staked to a validator). When you undelegate,
you will also receive the accumulated rewards automatically so you do not need to call the action to claim your rewards separately.

Let's undelegate our stake from the validator we staked to from before:

```bash
./build/nuklai-cli action undelegate-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
validators: 2
0: NodeID=NodeID-LTuNPrd43NoABqqZhM5ATaQsDNsuRRfiB
1: NodeID=NodeID-5XW8fe5MXkL1VD7Jj9WwPU6y2Lv7kt96g
✔ continue (y/n): y█
✅ txID: 2vPJ2wg6ocXA2mj5rBWBrFUxrV5JT5NTXfeFTmZkRR3YXJRAVc

```

Now, if we check the balance again, we should have our 1000000 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
assetID (use NAI for native token): NAI
[0m: NAI█
uri: http://127.0.0.1:34091/ext/bc/b1KuTqM6boQv4AQMujMfoE3YtknnTcCVwo4Mncsg2X8vYMM73
balance: 851998002.497201800 NAI
```
