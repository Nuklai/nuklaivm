# Emission Balancer

## Register a validator for staking

Even if there may be validators that are already taking part in the consensus of `nuklaivm` blocks, it doesn't mean they are automatically registered for the
staking mechanism. In order for a validator to register for staking on `nuklaivm`, they need to use the exact same account as they used while setting up the validator for
the Avalanche primary network. Validator Owners will need to stake a minimum of 100 NAI.

When you run the `run.sh` script, it runs the `tests/e2e/e2e_test.go` which in turn copies over the `signer.key` for all the auto-generated validator nodes. The reason we do
this is to make it easier to test the registration of the validator for staking on `nuklaivm` using `nuklai-cli`.

There are two ways of registering your validator for staking.

### Automatic run

We can let everything be configured automatically which means it'll set the values for staking automatically such as for:

- stakeStartBlock: Sets it to 40 blocks(2 minutes from now)
- stakeEndBlock: Sets it to 300 blocks(15 minutes from now)
- delegationFeeRate: Sets it to 50%
- rewardAddress: Sets it to the transaction actor

The only thing we would need to do is send some NAI that will be used for staking.

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

What this does is it imports the `staking.key` file located at `/tmp/nuklaivm/nodes/node1-bls/signer.key` and then tries to use it to register `node1` for staking.

This is because in order for registration of validator staking to work, you will need to use the same account you used while registering the validator to the Avalanche primary network to prevent unauthorized users from registering someone else's validator node.

If you don't have enough NAI in your account, you will see something like:

```bash
assetID: 11111111111111111111111111111111LpoYY
name: nuklaivm
symbol: NAI
balance: 0
please send funds to 0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373
exiting...
Balance of validator signer: 0.000000000
```

So, all we need to do is send at least 100 NAI to `0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```bash
Loading private key for node1
Validator Signer Address: 0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373
chainID: 2vhqHF49srJytVGZm8uKbLZshNqoqWxBBrPYpkqoSPTZbjtEEi
balance: 300.000000000 NAI
Balance of validator signer: 300.000000000
Loading validator signer key : 0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373
address: 0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373
chainID: 2vhqHF49srJytVGZm8uKbLZshNqoqWxBBrPYpkqoSPTZbjtEEi
Validator Signer Address: 0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373
Validator NodeID: NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv
balance: 300.000000000 NAI
✔ Staked amount: 100█
continue (y/n): y
Register Validator Stake Info - stakeStartBlock: 66 stakeEndBlock: 216 delegationFeeRate: 50 rewardAddress: 0213bab583ef405d7c12e3d0b0e079✔ continue (y/n): y█
✅ txID: xThHqzNUGiqnW6yokEDiWSm7pw1tzFHpHdVgnuFn4eSozEuQM
fee consumed: 0.000058800 NAI
output:  &{NodeID:NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv StakeStartBlock:66 StakeEndBlock:216 StakedAmount:100000000000 DelegationFeeRate:50 RewardAddress:0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373}
```

### Manual run

Here, we can be granular and set our own values for stakeStartBlock, stakeEndBlock, delegationFeeRate and rewardAddress.

First, let's import the key manually:

```bash
./build/nuklai-cli key import bls /tmp/nuklaivm/nodes/node2/signer.key
```

Which should output:

```bash
imported address: 02a16c8234d8879fc20efc2376af4f522df5f5d264c374b66a8c14c2b75c38d79d
```

Let's make sure we have enough balance to send

```bash
./build/nuklai-cli key balance
```

Which outputs:

```bash
address: 02a16c8234d8879fc20efc2376af4f522df5f5d264c374b66a8c14c2b75c38d79d balance: 200.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```bash
address: 02a16c8234d8879fc20efc2376af4f522df5f5d264c374b66a8c14c2b75c38d79d
chainID: 2pTmekN9DVmVzzzALW4oEwQiLxA4oyu7oE8oCfyEKXpRbK4Ezn
Validator Signer Address: 02a16c8234d8879fc20efc2376af4f522df5f5d264c374b66a8c14c2b75c38d79d
Validator NodeID: NodeID-AG21LM5ZuKNLpqgHr3b6VjDuZetZLCeXv
balance: 200.000000000 NAI
Staked amount: 100
Staking Start Block(must be after 263): 300
Staking End Block(must be after 300): 100000
Delegation Fee Rate(must be over 2): 90
Reward Address: 02a16c8234d8879fc20efc2376af4f522df5f5d264c374b66a8c14c2b75c38d79d
continue (y/n): y
Register Validator Stake Info - stakeStartBlock: 300 stakeEndBlock: 100000 delegationFeeRate: 90 rewardAddress: 02a16c8234d8879fc20efc2376✔ continue (y/n): y█
✅ txID: ZW5uuPsm6bg1DcqV6pccMdCT9k8LZXh1sZt7cvxtnVJ6RNmtw
fee consumed: 0.000058800 NAI
output:  &{NodeID:NodeID-AG21LM5ZuKNLpqgHr3b6VjDuZetZLCeXv StakeStartBlock:300 StakeEndBlock:100000 StakedAmount:100000000000 DelegationFeeRate:90 RewardAddress:02a16c8234d8879fc20efc2376af4f522df5f5d264c374b66a8c14c2b75c38d79d}
```

## Get Validator stake info

You may want to check your validator staking info such as stake start block height, stake end block height, staked amount, delegation fee rate and reward address. To do so, you can do:

Let's check the validator staking info for node1 or node2 we staked above.

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```bash
validators: 1
0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv
validator to get staking info for: 0 [auto-selected]
validator stake:
StakeStartBlock=66 StakeEndBlock=216 StakedAmount=100000000000 DelegationFeeRate=50 RewardAddress=0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373 OwnerAddress=0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373
```

You can also retrieve other useful info from Emission Balancer by doing

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```bash
chainID: 2vhqHF49srJytVGZm8uKbLZshNqoqWxBBrPYpkqoSPTZbjtEEi
validator 0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv PublicKey=sVzz3/tJL/x04uwwTrlB39LN1iuV44e45Fi55Jwk8LN6FX6WWQ/zEhAbSuzD9uHC Active=true StakedAmount=100000000000 AccumulatedStakedReward=268934 DelegationFeeRate=0.500000 DelegatedAmount=50000000000 AccumulatedDelegatedReward=178360
```

## Delegate stake to a validator

On `nuklaivm`, in addition to validators registering their nodes for staking, users can also delegate NAI for staking which means they get
to share the rewards sent to the validator. The delegation fee rate is set by the validator node when they register for staking so it is up to
the users to choose which validator to choose for staking purpose.

If a user chooses to delegate, they need to stake at least 25 NAI. A user can delegate to multiple validators at once.

To do so, you do:

```bash
./build/nuklai-cli action delegate-user-stake [auto | manual]
```

If successful, the output should be something like:

```bash
validators: 1
0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv
validator to delegate to: 0 [auto-selected]
balance: 199.999941200 NAI
✔ Staked amount: 50█
continue (y/n): y
✅ txID: 8PST8UKxbKWUo8RPtGxswnomGkNyqNaYaZ1RPCrAVQWRC9PVH
fee consumed: 0.000043000 NAI
output:  &{StakedAmount:50000000000 BalanceBeforeStake:199999898200 BalanceAfterStake:149999898200}
```

## Get Delegated User stake info

You may want to check your delegated staking info such as stake start block height, stake end block height, staked amount, and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```bash
validators: 1
0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv
validator to get staking info for: 0 [auto-selected]
user stake:
StakeStartBlock=84 StakeEndBlock=216 StakedAmount=50000000000 RewardAddress=0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373 OwnerAddress=0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373
```

## Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```bash
emission info:
CurrentBlockHeight=228 TotalSupply=1706000000001022626 MaxSupply=10000000000000000000 TotalStaked=0 RewardsPerEpoch=0 NumBlocksInEpoch=10 EmissionAddress=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 EmissionAccumulatedReward=66950
```

## Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```bash
validator 0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv PublicKey=sVzz3/tJL/x04uwwTrlB39LN1iuV44e45Fi55Jwk8LN6FX6WWQ/zEhAbSuzD9uHC Active=false StakedAmount=100000000000 AccumulatedStakedReward=322445 DelegationFeeRate=0.500000 DelegatedAmount=0 AccumulatedDelegatedReward=231868
```

We can also retrieve all the validators(both staked and unstaked):

```bash
./build/nuklai-cli emission all-validators
```

If successful, the output should be something like:

```bash
validator 0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv PublicKey=sVzz3/tJL/x04uwwTrlB39LN1iuV44e45Fi55Jwk8LN6FX6WWQ/zEhAbSuzD9uHC StakedAmount=100000000000 AccumulatedStakedReward=322445 DelegationFeeRate=0.500000 DelegatedAmount=0 AccumulatedDelegatedReward=231868
validator 1: NodeID=NodeID-P5fcvSzDLZXYFJKARgjXsrSrWk17EfsQy PublicKey=pjA5ku2ABxcwuA5M0Ka9WpqZmS0NhfozpOnm+H7di9+qJXABlV6BHmfghscMTym+ StakedAmount=0 AccumulatedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 AccumulatedDelegatedReward=0
```

## Claim delegator staking reward

On NuklaiVM, you are able to claim your staking rewards at any point in time without undelegating the stake from a validator. Also, as a validator, you're able to also claim your validator staking rewards at any point in time without unstaking your validator node.

First, we need to make sure that we're on our account we used for delegating our stake.

Let's first check the balance on our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```bash
address: 0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373 balance: 149.999898200 NAI
```

Next, let's check out how many NAI delegator rewards have been accumulated with our validator we staked to:

```bash
./build/nuklai-cli emission staked-validators
```

Which gives an output:

```bash
validator 0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv PublicKey=sVzz3/tJL/x04uwwTrlB39LN1iuV44e45Fi55Jwk8LN6FX6WWQ/zEhAbSuzD9uHC Active=false StakedAmount=100000000000 AccumulatedStakedReward=322445 DelegationFeeRate=0.500000 DelegatedAmount=0 AccumulatedDelegatedReward=231868
```

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-user-stake-reward
```

Which should produce a result like:

```bash
validators: 1
0: NodeID=NodeID-3CEtvCmZBYKwGnewYkAst7KopGk25j1Kv
validator to claim staking rewards from: 0 [auto-selected]
continue (y/n): y
✅ txID: 2XZq5pqUBUzGakuCWTQ15zUv3YXRQbFbX8DZtmhRGpErF2Ty1K
fee consumed: 0.000033100 NAI
output:  &{StakeStartBlock:84 StakeEndBlock:216 StakedAmount:50000000000 BalanceBeforeClaim:149999865100 BalanceAfterClaim:150000096968 DistributedTo:0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373}
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```bash
address: 0213bab583ef405d7c12e3d0b0e0794a29a9d8c13f4b4aad2d7c382b4312fbe373 balance: 150.000096968 NAI
```

NOTE: The longer you stake, the more rewards you will get for delegating. Also, note that as soon as you claim your reward, this timestamp is recorded and your staking duration will be reset. This incentivizes users from not claiming their rewards for as long as they want to and they get rewarded for it.

## Undelegate user stake

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

## Claim validator staking reward

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

## Withdraw validator stake

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
