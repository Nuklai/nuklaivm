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
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
Loading private key for node1
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
balance: 0.000000000 NAI
please send funds to nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s
exiting...
Balance of validator signer: 0.000000000
 You need a minimum of 100 NAI to register a validator
```

So, all we need to do is send at least 100 NAI to `nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
Loading private key for node1
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
balance: 1000.000000000 NAI
Balance of validator signer: 1000.000000000
Loading validator signer key : nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s
address: nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
Validator Signer Address: nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s
Validator NodeID: NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP
balance: 1000.000000000 NAI
Staked amount: 100
✔ continue (y/n): y█
Register Validator Stake Info - stakeStartTime: 2024-03-08 14:07:15 stakeEndTime: 2024-03-08 14:08:15 delegationFeeRate: 10 rewardAddress: nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s
✅ txID: 2G1xXXDHM3EYvkDERMvayP9bM15M7jXCLMPUfXvyEQ8mK47po9
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
imported address: nuklai1q2ntkzcr0sgm7tp8lp2uutkp8y9yn5fvpwwpejf96a4ur47lsr3g53trcep
```

Let's make sure we have enough balance to send

```bash
./build/nuklai-cli key balance
```

Which outputs:

```
database: .nuklai-cli
address: nuklai1q2ntkzcr0sgm7tp8lp2uutkp8y9yn5fvpwwpejf96a4ur47lsr3g53trcep
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:37643/ext/bc/2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
balance: 1000.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1q2ntkzcr0sgm7tp8lp2uutkp8y9yn5fvpwwpejf96a4ur47lsr3g53trcep
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
Validator Signer Address: nuklai1q2ntkzcr0sgm7tp8lp2uutkp8y9yn5fvpwwpejf96a4ur47lsr3g53trcep
Validator NodeID: NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4
balance: 1000.000000000 NAI
Staked amount: 100
Staking Start Time(must be after 2024-03-08 14:07:05) [YYYY-MM-DD HH:MM:SS]:  2024-03-08 14:09:00
✔ Staking End Time(must be after 2024-03-08 14:09:00) [YYYY-MM-DD HH:MM:SS]:  2024-03-08 14:10:00█
Delegation Fee Rate(must be over 2): 90
✔ Reward Address: nuklai1q2ntkzcr0sgm7tp8lp2uutkp8y9yn5fvpwwpejf96a4ur47lsr3g53trcep█
continue (y/n): y
Register Validator Stake Info - stakeStartTime: 2024-03-08 14:09:00 stakeEndTime: 2024-03-08 14:10:00 delegationFeeRate: 90 rewardAddress:✔ continue (y/n): y█
✅ txID: mgqBn4jfiRgRzoG9fetQWQzimtHMfdXef2ptxBxyDkc81xR8X
```

### Get Validator stake info

You may want to check your validator staking info such as stake start time, stake end time, staked amount, delegation fee rate and reward address. To do so, you can do:

Let's check the validator staking info for node1 or `NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP` which we staked above.

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validators: 2
0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP
1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4
validator to get staking info for: 0
validator stake:
StakeStartTime=2024-03-08 14:07:15 +0000 UTC StakeEndTime=2024-03-08 14:08:15 +0000 UTC StakedAmount=100000000000 DelegationFeeRate=10 RewardAddress=nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s OwnerAddress=nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s
```

You can also retrieve other useful info from Emission Balancer by doing

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validator 0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP PublicKey=i6c8rIdLFylcm5K6NCWB4IyGjC58erzCF05NpY0/dcYyfeupuUVxVXz0InJmnM8W StakedAmount=100000000000 UnclaimedStakedReward=221599 DelegationFeeRate=0.100000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4 PublicKey=rx/7QHSm6oJ3THOMYMeH0B8HFWfDI8SFT+L8t853+tK7oXO+B2bx8Q4OJ8VyDle9 StakedAmount=100000000000 UnclaimedStakedReward=60989 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
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
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validators: 2
0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP
1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4
validator to delegate to: 0
balance: 10000001.000000000 NAI
✔ Staked amount: 10000000█
continue (y/n): y
Delegate User Stake Info - stakeStartTime: 2024-03-08 14:11:05 stakeEndTime: 2024-03-08 14:12:05 rewardAddress: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
✅ txID: 2cL77MTMnZv1dDweR4NyFYmaFFtJCdjuAp1ZkGnrMPZGUrduk5
```

### Get Delegated User stake info

You may want to check your delegated staking info such as stake start time, stake end time, staked amount, and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validators: 2
0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP
1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4
validator to get staking info for: 0
validator stake:
StakeStartTime=2024-03-08 14:11:05 +0000 UTC StakedAmount=10000000000000000 RewardAddress=nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032 OwnerAddress=nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
```

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
emission info:
TotalSupply=853000004757277384 MaxSupply=10000000000000000000 TotalStaked=10000200000000000 RewardsPerEpoch=2378281963 NumBlocksInEpoch=10 EmissionAddress=nuklai1qqmzlnnredketlj3cu20v56nt5ken6thchra7nylwcrmz77td654w2jmpt9 EmissionUnclaimedBalance=116850
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission staked-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validator 0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP PublicKey=i6c8rIdLFylcm5K6NCWB4IyGjC58erzCF05NpY0/dcYyfeupuUVxVXz0InJmnM8W StakedAmount=100000000000 UnclaimedStakedReward=4281181834 DelegationFeeRate=0.100000 DelegatedAmount=10000000000000000 UnclaimedDelegatedReward=475653495
validator 1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4 PublicKey=rx/7QHSm6oJ3THOMYMeH0B8HFWfDI8SFT+L8t853+tK7oXO+B2bx8Q4OJ8VyDle9 StakedAmount=100000000000 UnclaimedStakedReward=187324 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

We can also retrieve all the validators(both staked and unstaked):

```bash
./build/nuklai-cli emission all-validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validator 0: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4 PublicKey=rx/7QHSm6oJ3THOMYMeH0B8HFWfDI8SFT+L8t853+tK7oXO+B2bx8Q4OJ8VyDle9 StakedAmount=100000000000 UnclaimedStakedReward=211106 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP PublicKey=i6c8rIdLFylcm5K6NCWB4IyGjC58erzCF05NpY0/dcYyfeupuUVxVXz0InJmnM8W StakedAmount=100000000000 UnclaimedStakedReward=6421614196 DelegationFeeRate=0.100000 DelegatedAmount=10000000000000000 UnclaimedDelegatedReward=713479313
validator 2: NodeID=NodeID-Ezr9kc5zHGbzQS8yYwLxMrzKug5qCu3NN PublicKey=pCBqcFOUALueEMdHp4lbzeml4B8shMh1BM2YzU/Y+pjGAHPZcfp7b3RvnMv2jWVX StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 3: NodeID=NodeID-AZk1eCy7uSvy8djBJPTQ2ZEXEQbditcdU PublicKey=ii/8gtx39tigvBduoD2+2VHpdIGAWK/DQrRaKK5J79w7GJlzvz2sf6FVsC6gMYpP StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 4: NodeID=NodeID-77g9bnJmXYSCK4pzVJeQC9qSdptYBkMCw PublicKey=qyjN2G1bHm2sz4/u8XIjjYse7mbmCFhW77t5QGiJv+K76Sd6r8/aUmlN9uGZav2X StakedAmount=0 UnclaimedStakedReward=0 DelegationFeeRate=0.000000 DelegatedAmount=0 UnclaimedDelegatedReward=0
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
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
assetID (use NAI for native token): NAI
✔ assetID (use NAI for native token): NAI█
balance: 0.999962800 NAI
```

Next, let's check out how many NAI delegator rewards have been accumulated with our validator we staked to:

```bash
./build/nuklai-cli emission staked-validators
```

Which gives an output:

```
database: .nuklai-cli
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validator 0: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4 PublicKey=rx/7QHSm6oJ3THOMYMeH0B8HFWfDI8SFT+L8t853+tK7oXO+B2bx8Q4OJ8VyDle9 StakedAmount=100000000000 UnclaimedStakedReward=282452 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
validator 1: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP PublicKey=i6c8rIdLFylcm5K6NCWB4IyGjC58erzCF05NpY0/dcYyfeupuUVxVXz0InJmnM8W StakedAmount=100000000000 UnclaimedStakedReward=12842911282 DelegationFeeRate=0.100000 DelegatedAmount=10000000000000000 UnclaimedDelegatedReward=1426956767
```

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-user-stake-reward
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validators: 2
0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP
1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4
validator to claim staking rewards from: 0
✔ continue (y/n): y█
✅ txID: smtojZG3amQthh7nzZsRHE9AphDN1ykcrnAHYmdMEsFCXvTT8
```

Let's check out balance again:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:37643/ext/bc/2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
balance: 2.189062190 NAI
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
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validators: 2
0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP
1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4
✔ validator to unstake from: 0█
✔ continue (y/n): y█
✅ txID: smWE7XzRLpcxrxfdWsxLdbKnHdohmPas1gJLe7DkGmgDBiSBD
```

Now, if we check the balance again, we should have our 1000000 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qgx6r6ttmxcjvdpjpchfwmyn4kaqz2ns0szx835egxxzd8y8ctef78dk032
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:37643/ext/bc/2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
balance: 10000002.902512344 NAI
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
address: nuklai1q2xrygad0vxqmu54e2dgm279a4ntaz6wvk3pxp6xtust4vnuymzmyknwt0s
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
✔ assetID (use NAI for native token): NAI█
uri: http://127.0.0.1:37643/ext/bc/2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
balance: 899.999946300 NAI
```

Next, let's check out how many NAI validator rewards have been accumulated with our validator:

```bash
./build/nuklai-cli emission staked-validators
```

Which gives an output:

```
database: .nuklai-cli
chainID: 2bdNEVxQ7w6TCjb9UE1pYkdcyqvb4xE8qTyJCsfm9Gy21LLTiE
validator 0: NodeID=NodeID-5f92nXVpfwLSRSgPZGmyi8rBiG88rsBCP PublicKey=i6c8rIdLFylcm5K6NCWB4IyGjC58erzCF05NpY0/dcYyfeupuUVxVXz0InJmnM8W StakedAmount=100000000000 UnclaimedStakedReward=19264490160 DelegationFeeRate=0.100000 DelegatedAmount=0 UnclaimedDelegatedReward=237829161
validator 1: NodeID=NodeID-59TagzsVYxnVbvqM8PeADJd7hBDYY5qT4 PublicKey=rx/7QHSm6oJ3THOMYMeH0B8HFWfDI8SFT+L8t853+tK7oXO+B2bx8Q4OJ8VyDle9 StakedAmount=100000000000 UnclaimedStakedReward=622225 DelegationFeeRate=0.900000 DelegatedAmount=0 UnclaimedDelegatedReward=0
```

Now, when we wanna claim our accumulated rewards, we do:

```bash
./build/nuklai-cli action claim-validator-stake-reward
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
