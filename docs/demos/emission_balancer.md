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
please send funds to 02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754
exiting...
Balance of validator signer: 0.000000000
```

So, all we need to do is send at least 100 NAI to `02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```bash
Loading private key for node1
Validator Signer Address: 02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754
chainID: 2VeYXe3bBXbxaxgW1ME46CFHLwgQ5ejdJcgtejyW6FGkfvduHY
balance: 300.000000000 NAI
Balance of validator signer: 300.000000000
Loading validator signer key : 02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754
address: 02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754
chainID: 2VeYXe3bBXbxaxgW1ME46CFHLwgQ5ejdJcgtejyW6FGkfvduHY
Validator Signer Address: 02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754
Validator NodeID: NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
balance: 300.000000000 NAI
✔ Staked amount: 100█
continue (y/n): y
Register Validator Stake Info - stakeStartBlock: 173 stakeEndBlock: 473 delegationFeeRate: 50 rewardAddress: 02da6ac369af5431778f509dc71e5continue (y/n): y
): y█
✅ txID: 2cZFW5xWKMBgx4L3WKKFjiaZAP5hgaqdEsGPBFfcFEhM4U5tkn
fee consumed: 0.000058800 NAI
output:  &{NodeID:NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm StakeStartBlock:173 StakeEndBlock:473 StakedAmount:100000000000 DelegationFeeRate:50 RewardAddress:02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754}
```

### Manual run

Here, we can be granular and set our own values for stakeStartBlock, stakeEndBlock, delegationFeeRate and rewardAddress.

First, let's import the key manually:

```bash
./build/nuklai-cli key import bls /tmp/nuklaivm/nodes/node2/signer.key
```

Which should output:

```bash
imported address: 02622d52de28b306fd25d4899cbfbfd496554a5de37f59a0c42e9dd6ebe52ea183
```

Let's make sure we have enough balance to send

```bash
./build/nuklai-cli key balance
```

Which outputs:

```bash
address: 02622d52de28b306fd25d4899cbfbfd496554a5de37f59a0c42e9dd6ebe52ea183 balance: 200.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```bash
Validator Signer Address: 02622d52de28b306fd25d4899cbfbfd496554a5de37f59a0c42e9dd6ebe52ea183
Validator NodeID: NodeID-HmCZrj6qRQattGzZf4t1PH6vQ6Lii3qeD
balance: 200.000000000 NAI
✔ Staked amount: 100█
✔ Staking Start Block(must be after 614): 700█
✔ Staking End Block(must be after 700): 1000█
Delegation Fee Rate(must be over 2): 90
Reward Address: 02622d52de28b306fd25d4899cbfbfd496554a5de37f59a0c42e9dd6ebe52ea183
continue (y/n): y
Register Validator Stake Info - stakeStartBlock: 700 stakeEndBlock: 1000 delegationFeeRate: 90 rewardAddress: 02622d52de28b306fd25d4899cbfbfd496554a5de37f59a0c42e9dd6ebe52ea183
✅ txID: 26int5SNG713mjzGGF6FK7VSeYbfgiCTYj4FJG9kzianqvzSJD
fee consumed: 0.000058800 NAI
output:  &{NodeID:NodeID-HmCZrj6qRQattGzZf4t1PH6vQ6Lii3qeD StakeStartBlock:700 StakeEndBlock:1000 StakedAmount:100000000000 DelegationFeeRate:90 RewardAddress:02622d52de28b306fd25d4899cbfbfd496554a5de37f59a0c42e9dd6ebe52ea183}
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
0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
validator to get staking info for: 0 [auto-selected]
validator stake:
StakeStartBlock=173 StakeEndBlock=473 StakedAmount=100000000000 DelegationFeeRate=50 RewardAddress=02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754 OwnerAddress=02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754
```

You can also retrieve other useful info from Emission Balancer by doing

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```bash
validator 0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm PublicKey=g5qscI5lmMiu3KyO9brJiizJcdXJ6d0BXEFsg16qsdIF/Qkgy3hO2qXbGhRmSk2y Active=true StakedAmount=100000000000 AccumulatedStakedReward=201492 DelegationFeeRate=0.500000 DelegatedAmount=0 AccumulatedDelegatedReward=0
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
0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
validator to delegate to: 0 [auto-selected]
balance: 199.999882400 NAI
✔ Staked amount: 50█
Staking Start Block(must be after 261): 300
[0m: 300█
Staking End Block(must be after 300): 1000
✔ continue (y/n): y█
✅ txID: BabSBcRUp2g59uEMTrFa8Aw4GSS1WnJj3aZ7teZ5zjei2Frsk
fee consumed: 0.000043000 NAI
output:  &{StakedAmount:50000000000 BalanceBeforeStake:199999839400 BalanceAfterStake:149999839400}
```

## Get Delegated User stake info

You may want to check your delegated staking info such as stake start block height, stake end block height, staked amount, and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```bash
validators: 1
0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
validator to get staking info for: 0 [auto-selected]
user stake:
StakeStartBlock=300 StakeEndBlock=473 StakedAmount=50000000000 RewardAddress=02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754 OwnerAddress=02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754
```

## Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```bash
emission info:
CurrentBlockHeight=288 TotalSupply=1706000000000523204 MaxSupply=10000000000000000000 TotalStaked=100000000000 RewardsPerEpoch=23782 NumBlocksInEpoch=10 EmissionAddress=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 EmissionAccumulatedReward=112400
```

## Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```bash
validator 0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm PublicKey=g5qscI5lmMiu3KyO9brJiizJcdXJ6d0BXEFsg16qsdIF/Qkgy3hO2qXbGhRmSk2y Active=true StakedAmount=100000000000 AccumulatedStakedReward=387184 DelegationFeeRate=0.500000 DelegatedAmount=0 AccumulatedDelegatedReward=0
```

We can also retrieve all the validators(both staked and unstaked):

```bash
./build/nuklai-cli emission all-validators
```

If successful, the output should be something like:

```bash
validator 0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm PublicKey=g5qscI5lmMiu3KyO9brJiizJcdXJ6d0BXEFsg16qsdIF/Qkgy3hO2qXbGhRmSk2y StakedAmount=100000000000 AccumulatedStakedReward=405021 DelegationFeeRate=0.500000 DelegatedAmount=50000000000 AccumulatedDelegatedReward=17836
```

## Claim delegator staking reward

On NuklaiVM, you are able to claim your staking rewards at any point in time without undelegating the stake from a validator. Also, as a validator, you're able to also claim your validator staking rewards at any point in time without unstaking your validator node.

First, we need to make sure that we're on our account we used for delegating our stake.

When we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-user-stake-reward
```

Which should produce a result like:

```bash
validators: 1
0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
validator to claim staking rewards from: 0 [auto-selected]
continue (y/n): y
✅ txID: ADyqY8bfDmXFELjXAo249hTmteGk9t2ij5Vp6GcdakazG78hB
fee consumed: 0.000033100 NAI
output:  &{StakeStartBlock:300 StakeEndBlock:473 StakedAmount:50000000000 BalanceBeforeClaim:149999806300 BalanceAfterClaim:149999841972 DistributedTo:02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754}
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

```bash
validators: 1
0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
validator to unstake from: 0 [auto-selected]
✔ continue (y/n): y█
✅ txID: 255GRo16jVCjXN4jJ176is25WC2ykr1rULM8ixPzHQRp29iMFg
fee consumed: 0.000033100 NAI
output:  &{StakeStartBlock:300 StakeEndBlock:473 UnstakedAmount:50000000000 DelegationFeeRate:0 RewardAmount:285376 BalanceBeforeUnstake:150000489797 BalanceAfterUnstake:200000775173 DistributedTo:02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754}
```

Note that if you check your balance, you will find that the delegator staking rewards were automatically claimed along with the original staked amount.

## Claim validator staking reward

On NuklaiVM, you are able to claim your validator staking rewards at any point in time without withdrawing your stake.

First, we need to make sure that we're on our account we used for registrating our validator stake.

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-validator-stake-reward
```

Which should produce a result like:

```bash
validators: 1
0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
validator to claim staking rewards for: 0 [auto-selected]
continue (y/n): y
✅ txID: 2pv4ajsNX4CPC3PDU7TjZGuirdyqBGtv66xdFKXBas2VYPtv28
fee consumed: 0.000035100 NAI
output:  &{StakeStartBlock:173 StakeEndBlock:473 StakedAmount:100000000000 DelegationFeeRate:50 BalanceBeforeClaim:149999738672 BalanceAfterClaim:150000522897 DistributedTo:02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754}
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

```bash
validators: 1
0: NodeID=NodeID-Ak5rbpMogUSVT5EAoHRJBxoFW8CstfSEm
validator to withdraw from staking: 0 [auto-selected]
continue (y/n): y
✅ txID: 3zxMhsdYhGK3wdbvafd8b8C1uvLZuT9jHpMy8Sqk4GUrk1757
fee consumed: 0.000035100 NAI
output:  &{StakeStartBlock:173 StakeEndBlock:473 UnstakedAmount:100000000000 DelegationFeeRate:50 RewardAmount:0 BalanceBeforeUnstake:0 BalanceAfterUnstake:100000000000 DistributedTo:02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754}
```

Now, if we check the balance again, we should have our 1000000 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```bash
address: 02da6ac369af5431778f509dc71e52318eb91cb7824d60167202f4f772e3dd1754 balance: 300.000740073 NAI
```

We got back our original staked amount and the validator staking rewards.
