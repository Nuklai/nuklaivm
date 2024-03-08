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
- stakeEndTime: Sets it to 5 minutes from now
- delegationFeeRate: Sets it to 50%
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
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
Loading private key for node1
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 0.000000000 NAI
please send funds to nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
exiting...
Balance of validator signer: 0.000000000
 You need a minimum of 100 NAI to register a validator
```

So, all we need to do is send at least 100 NAI to `nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
Loading private key for node1
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 1000.000000000 NAI
Balance of validator signer: 1000.000000000
Loading validator signer key : nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
address: nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
Validator Signer Address: nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
Validator NodeID: NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
balance: 1000.000000000 NAI
Staked amount: 999
[0m: 999█
continue (y/n): y
Register Validator Stake Info - stakeStartTime: 2024-03-08 15:51:12 stakeEndTime: 2024-03-08 15:59:12 delegationFeeRate: 50 rewardAddress:✔ continue (y/n): y█
✅ txID: br78t3vxESjYqMG9H3a1u3uWasCxZXdXVeFCrEmrFEGQFb6vP
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
imported address: nuklai1qtgvw2u8w9xcjvt85kmzex9pc7rzkcnknu4y6tay9ap3nhk53yk7kmz4xfv
```

Let's make sure we have enough balance to send

```bash
./build/nuklai-cli key balance
```

Which outputs:

```
database: .nuklai-cli
address: nuklai1qtgvw2u8w9xcjvt85kmzex9pc7rzkcnknu4y6tay9ap3nhk53yk7kmz4xfv
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:39255/ext/bc/jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 1000.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qtgvw2u8w9xcjvt85kmzex9pc7rzkcnknu4y6tay9ap3nhk53yk7kmz4xfv
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
Validator Signer Address: nuklai1qtgvw2u8w9xcjvt85kmzex9pc7rzkcnknu4y6tay9ap3nhk53yk7kmz4xfv
Validator NodeID: NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
balance: 1000.000000000 NAI
✔ Staked amount: 500█
✔ Staking Start Time(must be after 2024-03-08 15:50:18) [YYYY-MM-DD HH:MM:SS]: 2024-03-08 15:52:00█
Staking End Time(must be after 2024-03-08 15:52:00) [YYYY-MM-DD HH:MM:SS]: 2024-03-08 15:59:00
Delegation Fee Rate(must be over 2): 90
✔ Reward Address: nuklai1qtgvw2u8w9xcjvt85kmzex9pc7rzkcnknu4y6tay9ap3nhk53yk7kmz4xfv█
continue (y/n): y
Register Validator Stake Info - stakeStartTime: 2024-03-08 15:52:00 stakeEndTime: 2024-03-08 15:59:00 delegationFeeRate: 90 rewardAddress:
✅ txID: di8NckQfxAFzKm8mFTr2KPAb9qvGXYmfZbEVUWBdYCNVZxb3w
```

### Get Validator stake info

You may want to check your validator staking info such as stake start time, stake end time, staked amount, delegation fee rate and reward address. To do so, you can do:

Let's check the validator staking info for node1 or `NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np` which we staked above.

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validators: 2
0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
✔ validator to get staking info for: 0█
validator stake:
StakeStartTime=2024-03-08 15:51:12 +0000 UTC StakeEndTime=2024-03-08 15:59:12 +0000 UTC StakedAmount=999000000000 DelegationFeeRate=50 RewardAddress=nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh OwnerAddress=nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
```

You can also retrieve other useful info from Emission Balancer by doing

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validator 0: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn PublicKey=jV+P0nJG+dznWBtmZPlfneFRLQiYohAWKOc0xqskG1S1/7qThJiP0ZjNBqEUPEQI Active=true StakedAmount=500000000000 UnclaimedStakedReward=127866 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np PublicKey=hKaFqBQlpiUpSLwILT6RvKrEQo8YOtzIoQH2vnoOtGSBLWvCgJ+0gIpgBn9tH0XZ Active=true StakedAmount=999000000000 UnclaimedStakedReward=1247519 DelegationFeeRate=0.500000 DelegatedAmount=0 UnclaimedDelegatedReward=0
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
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validators: 2
0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
validator to delegate to: 0
balance: 852997999.999940634 NAI
Staked amount: 10000000
continue (y/n): y
Delegate User Stake Info - stakeStartTime: 2024-03-08 15:53:49 stakeEndTime: 2024-03-08 15:54:49 rewardAddress: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✅ txID: 12JcUsz7YH1DU5pKdoTkGtr6LfpFoHsqvjZXVDme4EWwL6K8n
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
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validators: 2
0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
validator to get staking info for: 0
validator stake:
StakeStartTime=2024-03-08 15:53:49 +0000 UTC StakedAmount=10000000000000000 RewardAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx OwnerAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
```

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
emission info:
TotalSupply=853000000004039656 MaxSupply=10000000000000000000 TotalStaked=10001499000000000 RewardsPerEpoch=2378590896 NumBlocksInEpoch=10 EmissionAddress=nuklai1qqmzlnnredketlj3cu20v56nt5ken6thchra7nylwcrmz77td654w2jmpt9 EmissionUnclaimedBalance=99350
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validator 0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np PublicKey=hKaFqBQlpiUpSLwILT6RvKrEQo8YOtzIoQH2vnoOtGSBLWvCgJ+0gIpgBn9tH0XZ Active=true StakedAmount=999000000000 UnclaimedStakedReward=1190966656 DelegationFeeRate=0.500000 DelegatedAmount=10000000000000000 UnclaimedDelegatedReward=1189243966
validator 1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn PublicKey=jV+P0nJG+dznWBtmZPlfneFRLQiYohAWKOc0xqskG1S1/7qThJiP0ZjNBqEUPEQI Active=true StakedAmount=500000000000 UnclaimedStakedReward=484599 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

We can also retrieve all the validators(both staked and unstaked):

```bash
./build/nuklai-cli emission all-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validator 0: NodeID=NodeID-v4W63aE5ssBj45RhLFudYepZ9GRifzG7 PublicKey=jJSRjy5dQMY+HVqMy7kDrAY3rGHf/PaDLX9us7oa1TDbRWIlLtUQ9KR7zQHjnBwW StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-35zPcQ7w4Y5apA53L87cryj9YtsE4xKJC PublicKey=l7ks5UHfXbUaHZjzWQyJx+w79oBdf6Fk7GX2qRJwlOs2zbUy2Ox8+WLs3DXpe8W8 StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 2: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn PublicKey=jV+P0nJG+dznWBtmZPlfneFRLQiYohAWKOc0xqskG1S1/7qThJiP0ZjNBqEUPEQI StakedAmount=500000000000 UnclaimedStakedReward=484599 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 3: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np PublicKey=hKaFqBQlpiUpSLwILT6RvKrEQo8YOtzIoQH2vnoOtGSBLWvCgJ+0gIpgBn9tH0XZ StakedAmount=999000000000 UnclaimedStakedReward=1190966656 DelegationFeeRate=0.500000 DelegatedAmount=10000000000000000 UnclaimedDelegatedReward=1189243966
validator 4: NodeID=NodeID-KiBRRHU14tiJSFk11KK4NHCFQncEiqypN PublicKey=ubE71p3trw8cxvW/mmxY86QnaRPwizpgg+KaklLv8XDTR4O/0wdszjILKvgeiUbu StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

### Claim delegator staking reward

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
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:39255/ext/bc/jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 842997999.999908686 NAI
```

Next, let's check out how many NAI delegator rewards have been accumulated with our validator we staked to:

```bash
./build/nuklai-cli emission staked-validators
```

Which gives an output:

```
database: .nuklai-cli
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validator 0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np PublicKey=hKaFqBQlpiUpSLwILT6RvKrEQo8YOtzIoQH2vnoOtGSBLWvCgJ+0gIpgBn9tH0XZ Active=true StakedAmount=999000000000 UnclaimedStakedReward=9515618600 DelegationFeeRate=0.500000 DelegatedAmount=10000000000000000 UnclaimedDelegatedReward=9513895910
validator 1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn PublicKey=jV+P0nJG+dznWBtmZPlfneFRLQiYohAWKOc0xqskG1S1/7qThJiP0ZjNBqEUPEQI Active=true StakedAmount=500000000000 UnclaimedStakedReward=1316976 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-user-stake-reward
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validators: 2
0: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
1: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
✔ validator to claim staking rewards from: 1█
continue (y/n): y
✅ txID: 2HojgJGuRyMzx7WN6KbRtvTMa4Xqp5zoopni7UvebLonQpi2FV
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:39255/ext/bc/jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 842998008.324536204 NAI
```

NOTE: The longer you stake, the more rewards you will get for delegating. Also, note that as soon as you claim your reward, this timestamp is recorded and your staking duration will be reset. This incentivizes users from not claiming their rewards for as long as they want to and they get rewarded for it.

### Undelegate user stake

Note that once your delegate your stake to a validator, it will never expire however, you may stop earning rewards once the validator stake period itself has ended. When you undelegate,
you will also receive the accumulated rewards automatically so you do not need to call the action to claim your rewards separately.

Let's undelegate our stake from the validator we staked to from before:

```bash
./build/nuklai-cli action undelegate-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validators: 2
0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
validator to unstake from: 0
continue (y/n): y
✅ txID: 2BVCh6bgAkdsjaWkZ46scUGWaBPPKMrMkk5D5Ke4J7BzTKXVCw
```

Now, if we check the balance again, we should have our 1000000 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:39255/ext/bc/jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 852998013.081458211 NAI
```

Note that the delegator staking rewards were automatically claimed along with the original staked amount.

### Claim validator staking reward

On NuklaiVM, you are able to claim your validator staking rewards at any point in time without withdrawing your stake.

First, we need to make sure that we're on our account we used for registrating our validator stake.

Let's first check the balance on our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:39255/ext/bc/jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 0.999946300 NAI
```

Next, let's check out how many NAI validator rewards have been accumulated with our validator:

```bash
./build/nuklai-cli emission staked-validators
```

Which gives an output:

```
database: .nuklai-cli
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validator 0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np PublicKey=hKaFqBQlpiUpSLwILT6RvKrEQo8YOtzIoQH2vnoOtGSBLWvCgJ+0gIpgBn9tH0XZ Active=true StakedAmount=999000000000 UnclaimedStakedReward=14273526338 DelegationFeeRate=0.500000 DelegatedAmount=0 UnclaimedDelegatedReward=1189250065
validator 1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn PublicKey=jV+P0nJG+dznWBtmZPlfneFRLQiYohAWKOc0xqskG1S1/7qThJiP0ZjNBqEUPEQI Active=true StakedAmount=500000000000 UnclaimedStakedReward=2271933 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-validator-stake-reward
```

Which should produce a result like:

```
reward:28)-(kpachhai)-(1230)-> ./build/nuklai-cli action claim-validator-stake-
database: .nuklai-cli
address: nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validators: 2
0: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
1: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
✔ validator to claim staking rewards for: 0█
✔ continue (y/n): y█
✅ txID: 2GDvKBwgdK1iD1WPVpDGRatrvZEXfNpW4CPU1b6QfeEyAvRbxQ
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:39255/ext/bc/jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 16.462696303 NAI
```

### Withdraw validator stake

Once your validator stake period has ended, you have two choices:

1. Withdraw your validator stake
2. Restake your validator node
   When you withdraw your validator stake, you will also receive the accumulated rewards automatically so you do not need to call the action to claim your rewards separately.

Let's withdraw our validator stake:

```bash
./build/nuklai-cli action withdraw-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
validators: 2
0: NodeID=NodeID-BEM9UVTj9JRgbX2s6Q74bsMBa675XEEDn
1: NodeID=NodeID-5Pki6UTB67KkqjtGPotj8Y1M5BHauw9np
validator to withdraw from staking: 1
continue (y/n): y
✅ txID: MJzUmJ8VJTdaV7foxVUzp8wzG63nV1XbLjizBn5ShRxKoJzB3

```

Now, if we check the balance again, we should have our 1000000 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1q2kh22fujdanqtevswk77fs79x6qugceppyj05tkhn8vfqy7hd6gvn02gfh
chainID: jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:39255/ext/bc/jf3YiBdAEsezfkMqcq7n3WhBECYyZdox5tjg83EgZUaKK6XQF
balance: 1015.462889085 NAI
```

We got back our original staked amount and the validator staking rewards.
