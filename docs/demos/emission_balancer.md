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
address: nuklai1qthtcezagawz0musdn9y4l7wx02j3xjm4q423fup9qpvkgquhmmkzmad9nh
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
Loading private key for node1
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
balance: 0.000000000 NAI
please send funds to nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3
exiting...
Balance of validator signer: 0.000000000
 You need a minimum of 100 NAI to register a validator
```

So, all we need to do is send at least 100 NAI to `nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qthtcezagawz0musdn9y4l7wx02j3xjm4q423fup9qpvkgquhmmkzmad9nh
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
Loading private key for node6
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
balance: 999.000000000 NAI
Balance of validator signer: 999.000000000
Loading validator signer key : nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3
address: nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
Validator Signer Address: nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3
Validator NodeID: NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb
balance: 999.000000000 NAI
Staked amount: 100
continue (y/n): y
Register Validator Stake Info - stakeStartTime: 2024-03-07 16:05:31 stakeEndTime: 2024-03-07 16:06:31 delegationFeeRate: 10 rewardAddress: nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3
✅ txID: osg6LKLLv9UaLtwJTe4QXquNBiimKWqzWntGQY3ioZj19PnJm
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
imported address: nuklai1qgrkfqwkedr80ka8lhd8kefjtdvvh0e2rq9el72pye86cy5mnr3w6k6c8hu
```

Let's make sure we have enough balance to send

```bash
./build/nuklai-cli key balance
```

Which outputs:

```
database: .nuklai-cli
address: nuklai1qgrkfqwkedr80ka8lhd8kefjtdvvh0e2rq9el72pye86cy5mnr3w6k6c8hu
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:33759/ext/bc/UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
balance: 1000.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```
database: .nuklai-cli
imported address: nuklai1qgrkfqwkedr80ka8lhd8kefjtdvvh0e2rq9el72pye86cy5mnr3w6k6c8hu
[~/repos/github.com/nuklai/nuklaivm]
ke manual)-(kpachhai)-(1008)-> ./build/nuklai-cli action register-validator-sta
database: .nuklai-cli
address: nuklai1qgrkfqwkedr80ka8lhd8kefjtdvvh0e2rq9el72pye86cy5mnr3w6k6c8hu
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
Validator Signer Address: nuklai1qgrkfqwkedr80ka8lhd8kefjtdvvh0e2rq9el72pye86cy5mnr3w6k6c8hu
Validator NodeID: NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG
balance: 1000.000000000 NAI
Staked amount: 100
✔ Staking Start Time(must be after 2024-03-07 16:05:24) [YYYY-MM-DD HH:MM:SS]: 2024-03-07 16:07:00█
Staking End Time(must be after 2024-03-07 16:07:00) [YYYY-MM-DD HH:MM:SS]: 2024-03-07 16:25:00
Delegation Fee Rate(must be over 2): 90
✔ Reward Address: nuklai1qgrkfqwkedr80ka8lhd8kefjtdvvh0e2rq9el72pye86cy5mnr3w6k6c8hu█
continue (y/n): y
Register Validator Stake Info - stakeStartTime: 2024-03-07 16:07:00 stakeEndTime: 2024-03-07 16:25:00 delegationFeeRate: 90 rewardAddress: nuklai1qgrkfqwkedr80ka8lhd8kefjtdvvh0e2rq9el72pye86cy5mnr3w6k6c8hu
✅ txID: FDAa8RKAV1s1nTfd1LCboGFwmiCLUpnaA7JAuxoNCeDACu7ZN
```

### Get Validator stake info

You may want to check your validator staking info such as stake start time, stake end time, staked amount, delegation fee rate and reward address. To do so, you can do:

Let's check the validator staking info for node1 or `NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb` which we staked above.

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validators: 2
0: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG
1: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb
validator to get staking info for: 1
validator stake:
StakeStartTime=2024-03-07 16:05:31 +0000 UTC StakeEndTime=2024-03-07 16:06:31 +0000 UTC StakedAmount=100000000000 DelegationFeeRate=10 RewardAddress=nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3 OwnerAddress=nuklai1qtfc4q9vz30v8nvgllf4decucpdux007495sh5kuc3hx53h264q5uvwp5m3
```

You can also retrieve other useful info from Emission Balancer by doing

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validator 0: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb PublicKey=pcBOMIh9EaslHXShtEEezRz1db6B6WGHMVymUkr2d2nkuBr/nqXS0ZUwMVCpNDfd StakedAmount=100000000000 UnclaimedStakedReward=245205  DelegationFeeRate=0.100000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG PublicKey=t0Jq/YIqsoXx75qpQ6BOhWx2rzhsB93MQjKLlQGDQGe/J+tHrh4WdU2xvahf1RRZ StakedAmount=100000000000 UnclaimedStakedReward=53817  DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
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
address: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validators: 2
0: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb
1: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG
validator to delegate to: 1
balance: 100000.000000000 NAI
Staked amount: 40000
continue (y/n): y
Delegate User Stake Info - stakeStartTime: 2024-03-07 16:10:44 stakeEndTime: 2024-03-07 16:11:44 rewardAddress: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
✅ txID: 51puQx5uSHocJjwFXvLtnX5cDY2feSddF9KWsdHxPbCmgMRSX
```

Let's delegate to another validator as well:

```
database: .nuklai-cli
address: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validators: 2
0: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb
1: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG
validator to delegate to: 0
balance: 59999.999968100 NAI
✔ Staked amount
: 40000█
continue (y/n): y
Delegate User Stake Info - stakeStartTime: 2024-03-07 16:10:57 stakeEndTime: 2024-03-07 16:11:57 rewardAddress: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
✅ txID: 2jYu8UB5Bwxgza1zdcecVTcpqJfNPFM3xax8myYhSnh1EmtNoU
```

### Get Delegated User stake info

You may want to check your delegated staking info such as stake start time, stake end time, staked amount, and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validators: 2
0: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb
1: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG
validator to get staking info for: 0
validator stake:
StakeStartTime=2024-03-07 16:10:57 +0000 UTC StakedAmount=40000000000000 RewardAddress=nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d OwnerAddress=nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
```

Let's check our stake info for another validator as well:

```
database: .nuklai-cli
address: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validators: 2
0: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb
1: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG
validator to get staking info for: 1
validator stake:
StakeStartTime=2024-03-07 16:10:44 +0000 UTC StakedAmount=40000000000000 RewardAddress=nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d OwnerAddress=nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
```

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
emission info:
TotalSupply=853000000059555314 MaxSupply=10000000000000000000 RewardsPerBlock=1906038 EmissionAddress=nuklai1qqmzlnnredketlj3cu20v56nt5ken6thchra7nylwcrmz77td654w2jmpt9 EmissionUnclaimedBalance=147650
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validator 0: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb PublicKey=pcBOMIh9EaslHXShtEEezRz1db6B6WGHMVymUkr2d2nkuBr/nqXS0ZUwMVCpNDfd StakedAmount=100000000000 UnclaimedStakedReward=34694495  DelegationFeeRate=0.100000 DelegatedAmount=40000000000000 UnclaimedDelegatedReward=3812837
validator 1: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG PublicKey=t0Jq/YIqsoXx75qpQ6BOhWx2rzhsB93MQjKLlQGDQGe/J+tHrh4WdU2xvahf1RRZ StakedAmount=100000000000 UnclaimedStakedReward=4559195  DelegationFeeRate=0.900000 DelegatedAmount=40000000000000 UnclaimedDelegatedReward=39476472
```

We can also retrieve all the validators(both staked and unstaked):

```bash
./build/nuklai-cli emission all-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: UxCAR64aRBKENSJU5z2qXgkPXkz7fM1dVB3djL7wvmTCb8dU1
validator 0: NodeID=NodeID-NdhCHAP4QcirWB8R2mhCeT9LbTK9tXUK8 PublicKey=k7uzsetalee3UtKpka7t0iD3EccEjmkLgcVwY7SY0O90Qyv1qnbS8Pjswsq27aaJ StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-k8S4xtBu7MkfR45ojmvceCvWeENBmtBK PublicKey=hRo1jpSkQTV0dWAoZPIFZk+j0wXqHgOKUSAoWA7tMSlFU4m1T/dklWKGMZZFluMt StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 2: NodeID=NodeID-LSvpQLK7t8XzMhcNr3TK687cKquG1AtUG PublicKey=t0Jq/YIqsoXx75qpQ6BOhWx2rzhsB93MQjKLlQGDQGe/J+tHrh4WdU2xvahf1RRZ StakedAmount=100000000000 UnclaimedStakedReward=5321611 DelegationFeeRate=0.900000 DelegatedAmount=40000000000000 UnclaimedDelegatedReward=46338208
validator 3: NodeID=NodeID-67n4HUguYgerqGoVJcggpevxodYgccpAb PublicKey=pcBOMIh9EaslHXShtEEezRz1db6B6WGHMVymUkr2d2nkuBr/nqXS0ZUwMVCpNDfd StakedAmount=100000000000 UnclaimedStakedReward=41556239 DelegationFeeRate=0.100000 DelegatedAmount=40000000000000 UnclaimedDelegatedReward=4575245
validator 4: NodeID=NodeID-ACR6jK4YA2xvrLebeYFK9uDbNTYHi25jv PublicKey=iDv+XIqgMxMZBZMVTGQTXG5EBSBeISOahGzM1It4QmSFJrWlMKV5HSjCGdI1s9PD StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
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
address: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
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
address: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
chainID: 2DXc2cEtPDGTy7JqaWLNV6Lmg8tD75KUnSSLaGyidEb2Gvi5ii
validators: 1
0: NodeID=NodeID-3GJnWPtDnvR4g7PbsJFqURNcpZj5UZVgb
validator to claim staking rewards from: 0 [auto-selected]
continue (y/n): y
✅ txID: 2Wm8vHBz2n2DiV2LM7j337PCizgWMjoSvJVAQPZm5GwoEHFLbq
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qr0x7dw7vryvnt43vf7astvetv77y79lczh5zlv9wsv8q9wtqalqxw66h8d
chainID: 2DXc2cEtPDGTy7JqaWLNV6Lmg8tD75KUnSSLaGyidEb2Gvi5ii
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:39231/ext/bc/2DXc2cEtPDGTy7JqaWLNV6Lmg8tD75KUnSSLaGyidEb2Gvi5ii
balance: 1.006778113 NAI
```

NOTE: The longer you stake, the more rewards you will get for delegating. Also, note that as soon as you claim your reward, this timestamp is recorded and your staking duration will be reset. This incentivizes users from not claiming their rewards for as long as they want to and they get rewarded for it.

### Undelegate user stake

Once your delegated stake period has ended, perhaps you may want to undelegate your stake(i.e. you want to get back your NAI you have staked to a validator). When you undelegate,
you will also receive the accumulated rewards automatically so you do not need to call the action to claim your rewards separately.

Let's first check the balance on our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2wBturxbiFoz1VdHx4WH17LFEgeScH2EVYUTwsJkqt4cVoH79M
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:39289/ext/bc/2wBturxbiFoz1VdHx4WH17LFEgeScH2EVYUTwsJkqt4cVoH79M
balance: 852998899.999938488 NAI
```

Now, let's undelegate our stake from the validator we stkaed to from before:

```bash
./build/nuklai-cli action undelegate-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2wBturxbiFoz1VdHx4WH17LFEgeScH2EVYUTwsJkqt4cVoH79M
validators: 1
0: NodeID=NodeID-D2nVXseiQajtsKq8zdqX8vjCGV7Uz7HWZ
validator to unstake from: 0 [auto-selected]
continue (y/n): y
✅ txID: 2AEiT9L1h8o7yXoEq3kt6p4ezSVZ1SMvaeAYWVxqC2DyQDRPUb
```

Now, if we check the balance again, we should have our 100 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:44089/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
balance: 852999899.999917388 NAI
```
