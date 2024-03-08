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
balance: 0.000000000 NAI
please send funds to nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt
exiting...
Balance of validator signer: 0.000000000
 You need a minimum of 100 NAI to register a validator
```

So, all we need to do is send at least 100 NAI to `nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```
Loading private key for node1
chainID: sFv8uJFL9XnLBDbn3xJ2pusQHxC3yfNJ9iPV7xN4WeZGQ9HTa
balance: 1000.000000000 NAI
Balance of validator signer: 1000.000000000
Loading validator signer key : nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt
address: nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt
chainID: sFv8uJFL9XnLBDbn3xJ2pusQHxC3yfNJ9iPV7xN4WeZGQ9HTa
Validator Signer Address: nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt
Validator NodeID: NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
balance: 1000.000000000 NAI
✔ Staked amount: 999█
✔ continue (y/n): y█
Register Validator Stake Info - stakeStartTime: 2024-03-08 17:07:05 stakeEndTime: 2024-03-08 17:20:05 delegationFeeRate: 50 rewardAddress: nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt
✅ txID: 21vrzuZpg4UE2x3LpKkvFtcgyj215rc311mkTQgWT7hYEdRaF9
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
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:43321/ext/bc/sFv8uJFL9XnLBDbn3xJ2pusQHxC3yfNJ9iPV7xN4WeZGQ9HTa
balance: 1000.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```
Validator Signer Address: nuklai1q2lwstvq7dwx7jeae4klwzp2psfpq7gmfvp2c4nr0evaj8gjgu3zx56dud3
Validator NodeID: NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
balance: 1000.000000000 NAI
✔ Staked amount: 999█
Staking Start Time(must be after 2024-03-08 17:06:27) [YYYY-MM-DD HH:MM:SS]: 2024-03-08 17:08:00
Staking End Time(must be after 2024-03-08 17:08:00) [YYYY-MM-DD HH:MM:SS]: 2024-03-08 17:25:00
Delegation Fee Rate(must be over 2): 90
Reward Address: nuklai1q2lwstvq7dwx7jeae4klwzp2psfpq7gmfvp2c4nr0evaj8gjgu3zx56dud3
continue (y/n): y
Register Validator Stake Info - stakeStartTime: 2024-03-08 17:08:00 stakeEndTime: 2024-03-08 17:25:00 delegationFeeRate: 90 rewardAddress:✔ continue (y/n): y█
✅ txID: DQ9ApYWKy6Cev9F7dVBYhdJGodCCezrnA9wxYYyseTM3HecUM
```

### Get Validator stake info

You may want to check your validator staking info such as stake start time, stake end time, staked amount, delegation fee rate and reward address. To do so, you can do:

Let's check the validator staking info for node1 or `NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9` which we staked above.

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```
validators: 2
0: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
1: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
✔ validator to get staking info for: 1█
validator stake:
StakeStartTime=2024-03-08 17:07:05 +0000 UTC StakeEndTime=2024-03-08 17:20:05 +0000 UTC StakedAmount=999000000000 DelegationFeeRate=50 RewardAddress=nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt OwnerAddress=nuklai1qgmr8jxysf6c47rc7v86cz5h7y4zj5twmpwkl0js30dpnq5vt0n3jqz2mkt
```

You can also retrieve other useful info from Emission Balancer by doing

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
validator 0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9 PublicKey=h3QqCh/LwqpFbuRRtsB5/kQ2dT+EmZC+q7DdaYY/rRNoHpdSBbOQR/+Tl85lyPaW Active=true StakedAmount=999000000000 UnclaimedStakedReward=0 DelegationFeeRate=0.500000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs PublicKey=mDSuRu8Z7sYQDASICxvvKoAz5NcnOq4z+dYRzUMld5JqLoL9R0OKrXDOIdDDmxiU Active=false StakedAmount=999000000000 UnclaimedStakedReward=0 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
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
validators: 2
0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
validator to delegate to: 0
balance: 1000000.000000000 NAI
✔ continue (y/n): y█
✅ txID: 2RosZKQY3K74Q9i1yJfm6DdvMPMsSbZCLFke9v2oFQEEBFAMT5
```

### Get Delegated User stake info

You may want to check your delegated staking info such as stake start time, stake end time, staked amount, and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```
validators: 2
0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
validator to get staking info for: 0
validator stake:
StakeStartTime=2024-03-08 17:11:11 +0000 UTC StakedAmount=100000000000000 RewardAddress=nuklai1qgtvmjhh5xkjh5tf993s05fptc2l0mzn6j8yw72pmrqpa947xsp5scqsrma OwnerAddress=nuklai1qgtvmjhh5xkjh5tf993s05fptc2l0mzn6j8yw72pmrqpa947xsp5scqsrma
```

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
emission info:
TotalSupply=853000000051841218 MaxSupply=10000000000000000000 TotalStaked=313986000000000 RewardsPerEpoch=74673230 NumBlocksInEpoch=10 EmissionAddress=nuklai1qqmzlnnredketlj3cu20v56nt5ken6thchra7nylwcrmz77td654w2jmpt9 EmissionUnclaimedBalance=116850
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
validator 0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9 PublicKey=h3QqCh/LwqpFbuRRtsB5/kQ2dT+EmZC+q7DdaYY/rRNoHpdSBbOQR/+Tl85lyPaW Active=true StakedAmount=999000000000 UnclaimedStakedReward=24986235 DelegationFeeRate=0.500000 DelegatedAmount=100000000000000 UnclaimedDelegatedReward=24028467
validator 1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs PublicKey=mDSuRu8Z7sYQDASICxvvKoAz5NcnOq4z+dYRzUMld5JqLoL9R0OKrXDOIdDDmxiU Active=true StakedAmount=999000000000 UnclaimedStakedReward=1195518 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

We can also retrieve all the validators(both staked and unstaked):

```bash
./build/nuklai-cli emission all-validators
```

If successful, the output should be something like:

```
validator 0: NodeID=NodeID-3fHReqX1SQ6CnysHFqPLhGxVu1JkmkzPd PublicKey=mP6hTnPgQ0UfipB4zgrn/Jc4s5J4WDkUT/r9GqV8kSsJcF2lX000ML7kbGHHZeRc StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-QGQzSLyTH811fJeGfoik8gv1pVTSKSe7F PublicKey=oqrBGXJuf/QwaMZoXEX9weibtEwnB8gmTNN9wyAVJEwDkbCZbEGMJi/vCU/3jxyA StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 2: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs PublicKey=mDSuRu8Z7sYQDASICxvvKoAz5NcnOq4z+dYRzUMld5JqLoL9R0OKrXDOIdDDmxiU StakedAmount=999000000000 UnclaimedStakedReward=1195518 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 3: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9 PublicKey=h3QqCh/LwqpFbuRRtsB5/kQ2dT+EmZC+q7DdaYY/rRNoHpdSBbOQR/+Tl85lyPaW StakedAmount=999000000000 UnclaimedStakedReward=24986235 DelegationFeeRate=0.500000 DelegatedAmount=100000000000000 UnclaimedDelegatedReward=24028467
validator 4: NodeID=NodeID-8YfjNYJQe3ZAnW6LBvWXLZTn7qZYG7LeV PublicKey=rEk7wqfpVK8l/7UvT5pGUO/gWDrhUXSdjk/gymIwLsOD5IwjLtxqfblOPkgYqwHq StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
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
address: nuklai1qgtvmjhh5xkjh5tf993s05fptc2l0mzn6j8yw72pmrqpa947xsp5scqsrma
chainID: sFv8uJFL9XnLBDbn3xJ2pusQHxC3yfNJ9iPV7xN4WeZGQ9HTa
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:43321/ext/bc/sFv8uJFL9XnLBDbn3xJ2pusQHxC3yfNJ9iPV7xN4WeZGQ9HTa
balance: 899999.999962800 NAI
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
validators: 2
0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
✔ validator to claim staking rewards from: 0█
continue (y/n): y
✅ txID: kw1goRDWnQEpw9Lqs6wmY1pUZKybumVnNFTMUN1vktt9fFgdn
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
assetID (use NAI for native token): NAI
✔ assetID (use NAI for native token): NAI█
balance: 900000.120003040 NAI
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
validators: 2
0: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
1: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
✔ validator to unstake from: 1█
continue (y/n): y
✅ txID: bTFRsFwyMJT4sESishFHeAihL9SnBFvirrSExp95eJTFVLyLz
```

Now, if we check the balance again, we should have our 1000000 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
assetID (use NAI for native token): NAI
✔ assetID (use NAI for native token): NAI█
balance: 1000000.143995668 NAI
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
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:43321/ext/bc/sFv8uJFL9XnLBDbn3xJ2pusQHxC3yfNJ9iPV7xN4WeZGQ9HTa
balance: 0.999946300 NAI
```

Next, let's check out how many NAI validator rewards have been accumulated with our validator:

```bash
./build/nuklai-cli emission staked-validators
```

Which gives an output:

```
validator 0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9 PublicKey=h3QqCh/LwqpFbuRRtsB5/kQ2dT+EmZC+q7DdaYY/rRNoHpdSBbOQR/+Tl85lyPaW Active=true StakedAmount=999000000000 UnclaimedStakedReward=157810345 DelegationFeeRate=0.500000 DelegatedAmount=0 UnclaimedDelegatedReward=12020234
validator 1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs PublicKey=mDSuRu8Z7sYQDASICxvvKoAz5NcnOq4z+dYRzUMld5JqLoL9R0OKrXDOIdDDmxiU Active=true StakedAmount=999000000000 UnclaimedStakedReward=4521749 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-validator-stake-reward
```

Which should produce a result like:

```
validators: 2
0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
validator to claim staking rewards for: 0
continue (y/n): y
✅ txID: 28AcHZ5RUTDZxdmtYDRB1Lw3GeyugUKcWhMrsF8NWstqae2kin
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
assetID (use NAI for native token): NAI
✔ assetID (use NAI for native token): NAI█
balance: 1.169750479 NAI
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
validators: 2
0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
✔ validator to withdraw from staking: 0█
continue (y/n): y
✅ txID: W3MKCnLxJeEf27xTQdSMwaZLujaMvtqxgE7BZJsA8rhkroBDP

```

Now, if we check the balance again, we should have our 1000000 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:43321/ext/bc/sFv8uJFL9XnLBDbn3xJ2pusQHxC3yfNJ9iPV7xN4WeZGQ9HTa
balance: 1000.172072743 NAI
```

We got back our original staked amount and the validator staking rewards.
