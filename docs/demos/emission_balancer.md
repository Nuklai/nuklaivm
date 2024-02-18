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
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
Loading private key for node1
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
balance: 0.000000000 NAI
please send funds to nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc
exiting...
Balance of validator signer: 0.000000000
 You need a minimum of 100 NAI to register a validator
```

So, all we need to do is send at least 100 NAI to `nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc`

After sending some NAI to this account using `./build/nuklai-cli action transfer`, let's try this again:

```bash
./build/nuklai-cli action register-validator-stake auto node1
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
Loading private key for node1
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
balance: 101.000000000 NAI
Balance of validator signer: 101.000000000
Loading validator signer key : nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc
address: nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
Validator Signer Address: nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc Public Key: t7et4jcSplg6okXPLyWVaWmEMajENU82RREoqCZUNtzatC9AVIdfUkbrtPKyPDd3
Validator NodeID: NodeID-AySfsjEAiDZUZwW8XN8gFmvRU4PAbdPTQ
balance: 101.000000000 NAI
✔ Staked amount: 100
continue (y/n): y█
Register Validator Stake Info -  stakeStartTime: 2024-02-17 00:47:48 stakeEndTime: 2024-02-17 00:48:48 delegationFeeRate: 10 rewardAddress: nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc
✅ txID: 2kZWFPX242f7o6FJRGxnCZERvW2vH3f2jgbwuaW9w7xff95BGc
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
imported address: nuklai1qgzqljfa8zzrg9ne8vuu9txjxvt9n3ns58zq3kfry9pryw4kqknjkfkr30c
```

Let's make sure we have enough balance to send

```bash
./build/nuklai-cli key balance
```

Which outputs:

```
database: .nuklai-cli
address: nuklai1qgzqljfa8zzrg9ne8vuu9txjxvt9n3ns58zq3kfry9pryw4kqknjkfkr30c
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
assetID (use NAI for native token): NAI
uri: http://127.0.0.1:32911/ext/bc/2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
balance: 400.000000000 NAI
```

To register our validator for staking manually, we can do:

```bash
./build/nuklai-cli action register-validator-stake manual
```

If successful, you should see something like:

```
database: .nuklai-cli
address: nuklai1qgzqljfa8zzrg9ne8vuu9txjxvt9n3ns58zq3kfry9pryw4kqknjkfkr30c
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
Validator Signer Address: nuklai1qgzqljfa8zzrg9ne8vuu9txjxvt9n3ns58zq3kfry9pryw4kqknjkfkr30c Public Key: rxgK6sl8wgKLTNidfKrXgPRuwTDC8SqRy25ZXl8Sy20U6nEXv0nrVj0XYKbGynMd
Validator NodeID: NodeID-Lz8FCjSXyLwF8WUKAPWwFgmLYhwhEQzXe
balance: 400.000000000 NAI
✔ Staked amount: 200█
Staking Start Time(must be after 2024-02-17 01:02:48) [YYYY-MM-DD HH:MM:SS]:  2024-02-17 01:05:00
✔ Staking End Time(must be after 2024-02-17 01:05:00) [YYYY-MM-DD HH:MM:SS]:  2024-02-17 01:09:00█
✔ Delegation Fee Rate(must be over 2): 3█
✔ Reward Address: nuklai1qgzqljfa8zzrg9ne8vuu9txjxvt9n3ns58zq3kfry9pryw4kqknjkfkr30c█
✔ continue (y/n): y█
Register Validator Stake Info - stakeStartTime: 2024-02-17 01:05:00 stakeEndTime: 2024-02-17 01:09:00 delegationFeeRate: 3 rewardAddress: nuklai1qgzqljfa8zzrg9ne8vuu9txjxvt9n3ns58zq3kfry9pryw4kqknjkfkr30c

✅ txID: 2NSaH96NYQuSk1FEppo1sxJPuceMzDRxQVE1dzrmHuLGEFy7FC
```

### Get Validator stake info

You may want to check your validator staking info such as stake start time, stake end time, staked amount, delegation fee rate and reward address. To do so, you can do:

Let's check the validator staking info for node1 or `NodeID=NodeID-AySfsjEAiDZUZwW8XN8gFmvRU4PAbdPTQ` which we staked above.

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
validators: 5
0: NodeID=NodeID-8toEuCFZ9E3jqee6X6ynWagzEXdYy6XZU NodePublicKey=pz2sY5ChOJljXCPcB2MJSTKGjJ7aipJ7N4TUon/k5kWTA+pCoxqAPpWiGNpM6VHK
1: NodeID=NodeID-PZofQ3NucKr43wxZN9vWzxiMnvW4k2Ygp NodePublicKey=q0EvKTkjfDjY4YzE7AX4hldyhe2HjXf6JsZMkmtcES99qaiZu5J4VPaZ5U4BWFPT
2: NodeID=NodeID-Lz8FCjSXyLwF8WUKAPWwFgmLYhwhEQzXe NodePublicKey=rxgK6sl8wgKLTNidfKrXgPRuwTDC8SqRy25ZXl8Sy20U6nEXv0nrVj0XYKbGynMd
3: NodeID=NodeID-AySfsjEAiDZUZwW8XN8gFmvRU4PAbdPTQ NodePublicKey=t7et4jcSplg6okXPLyWVaWmEMajENU82RREoqCZUNtzatC9AVIdfUkbrtPKyPDd3
4: NodeID=NodeID-F4fToB3WiUVZbBqTMEyNz7rZQ9h4Vp9YC NodePublicKey=s8KHv77JjvPkVi0xdLYnqA5XCgH2dKXzI2leBBSvbDaOcH0cCNQbCoEXz1oNmHes
✔ validator to register for staking: 3
validator stake:
StakeStartTime=1708130868 StakeEndTime=1708130928 StakedAmount=100000000000 DelegationFeeRate=10 RewardAddress=nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc OwnerAddress=nuklai1qfhh55jdfeg27l8w9acrq3ytcx45rcs26rskdaam58jcdtuuzpgsxsy8rsc
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
address: nuklai1q2w6mhxzvttqkkunrd25swy4n5crcld8e5v99cr2fy6h8vzxl739qcqnxqv
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
validators: 5
0: NodeID=NodeID-8toEuCFZ9E3jqee6X6ynWagzEXdYy6XZU NodePublicKey=pz2sY5ChOJljXCPcB2MJSTKGjJ7aipJ7N4TUon/k5kWTA+pCoxqAPpWiGNpM6VHK
1: NodeID=NodeID-PZofQ3NucKr43wxZN9vWzxiMnvW4k2Ygp NodePublicKey=q0EvKTkjfDjY4YzE7AX4hldyhe2HjXf6JsZMkmtcES99qaiZu5J4VPaZ5U4BWFPT
2: NodeID=NodeID-Lz8FCjSXyLwF8WUKAPWwFgmLYhwhEQzXe NodePublicKey=rxgK6sl8wgKLTNidfKrXgPRuwTDC8SqRy25ZXl8Sy20U6nEXv0nrVj0XYKbGynMd
3: NodeID=NodeID-AySfsjEAiDZUZwW8XN8gFmvRU4PAbdPTQ NodePublicKey=t7et4jcSplg6okXPLyWVaWmEMajENU82RREoqCZUNtzatC9AVIdfUkbrtPKyPDd3
4: NodeID=NodeID-F4fToB3WiUVZbBqTMEyNz7rZQ9h4Vp9YC NodePublicKey=s8KHv77JjvPkVi0xdLYnqA5XCgH2dKXzI2leBBSvbDaOcH0cCNQbCoEXz1oNmHes
✔ validator to delegate to: 3
balance: 897.999913400 NAI
Staked amount: 25
continue (y/n): y
Delegate User Stake Info - stakeStartTime: 2024-02-18 22:39:25 stakeEndTime: 2024-02-18 22:40:25 rewardAddress: nuklai1q2w6mhxzvttqkkunrd25swy4n5crcl✔ continue (y/n): y█
✅ txID: A1UZezFUnSK9zryupueQsBBMwRU6tULwLMnwi7W3FnThYdhbB
```

### Get Delegated User stake info

You may want to check your delegated staking info such as stake start time, stake end time, staked amount, and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1q2w6mhxzvttqkkunrd25swy4n5crcld8e5v99cr2fy6h8vzxl739qcqnxqv
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
validators: 5
0: NodeID=NodeID-8toEuCFZ9E3jqee6X6ynWagzEXdYy6XZU NodePublicKey=pz2sY5ChOJljXCPcB2MJSTKGjJ7aipJ7N4TUon/k5kWTA+pCoxqAPpWiGNpM6VHK
1: NodeID=NodeID-PZofQ3NucKr43wxZN9vWzxiMnvW4k2Ygp NodePublicKey=q0EvKTkjfDjY4YzE7AX4hldyhe2HjXf6JsZMkmtcES99qaiZu5J4VPaZ5U4BWFPT
2: NodeID=NodeID-Lz8FCjSXyLwF8WUKAPWwFgmLYhwhEQzXe NodePublicKey=rxgK6sl8wgKLTNidfKrXgPRuwTDC8SqRy25ZXl8Sy20U6nEXv0nrVj0XYKbGynMd
3: NodeID=NodeID-AySfsjEAiDZUZwW8XN8gFmvRU4PAbdPTQ NodePublicKey=t7et4jcSplg6okXPLyWVaWmEMajENU82RREoqCZUNtzatC9AVIdfUkbrtPKyPDd3
4: NodeID=NodeID-F4fToB3WiUVZbBqTMEyNz7rZQ9h4Vp9YC NodePublicKey=s8KHv77JjvPkVi0xdLYnqA5XCgH2dKXzI2leBBSvbDaOcH0cCNQbCoEXz1oNmHes
✔ validator to get staking info for: 3
validator stake:
StakeStartTime=1708295965 StakeEndTime=1708296025 StakedAmount=25000000000 RewardAddress=nuklai1q2w6mhxzvttqkkunrd25swy4n5crcld8e5v99cr2fy6h8vzxl739qcqnxqv OwnerAddress=nuklai1q2w6mhxzvttqkkunrd25swy4n5crcld8e5v99cr2fy6h8vzxl739qcqnxqv
```

### Undelegate user stake

Once your delegated stake period has ended, perhaps you may want to undelegate your stake(i.e. you want to get back your NAI you have staked to a validator). When you undelegate,
you will also receive the accumulated rewards automatically so you do not need to call the action to claim your rewards separately.

```bash
./build/nuklai-cli action undelegate-user-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1q2w6mhxzvttqkkunrd25swy4n5crcld8e5v99cr2fy6h8vzxl739qcqnxqv
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
chainID: 2oRdajKFemnW1zg7QX5yRkHvAWkpJ1W9b5xHpCqJ5X43Qyih3H
validators: 5
0: NodeID=NodeID-8toEuCFZ9E3jqee6X6ynWagzEXdYy6XZU NodePublicKey=pz2sY5ChOJljXCPcB2MJSTKGjJ7aipJ7N4TUon/k5kWTA+pCoxqAPpWiGNpM6VHK
1: NodeID=NodeID-PZofQ3NucKr43wxZN9vWzxiMnvW4k2Ygp NodePublicKey=q0EvKTkjfDjY4YzE7AX4hldyhe2HjXf6JsZMkmtcES99qaiZu5J4VPaZ5U4BWFPT
2: NodeID=NodeID-Lz8FCjSXyLwF8WUKAPWwFgmLYhwhEQzXe NodePublicKey=rxgK6sl8wgKLTNidfKrXgPRuwTDC8SqRy25ZXl8Sy20U6nEXv0nrVj0XYKbGynMd
3: NodeID=NodeID-AySfsjEAiDZUZwW8XN8gFmvRU4PAbdPTQ NodePublicKey=t7et4jcSplg6okXPLyWVaWmEMajENU82RREoqCZUNtzatC9AVIdfUkbrtPKyPDd3
4: NodeID=NodeID-F4fToB3WiUVZbBqTMEyNz7rZQ9h4Vp9YC NodePublicKey=s8KHv77JjvPkVi0xdLYnqA5XCgH2dKXzI2leBBSvbDaOcH0cCNQbCoEXz1oNmHes
✔ validator to unstake from: 3
✔ continue (y/n): y█
✅ txID: 21JdRhtLJ4ra5QPixgdZ184iQ4sRzPzJS81SZ1AkaHBniLvggJ
```

### Get Emission Info

We can check info that emission is in charge of such as total supply, max supply, rewards per block and the validators' stake

```bash
./build/nuklai-cli emission info
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
emission info:
TotalSupply=853000594000000000 MaxSupply=10000000000000000000 RewardsPerBlock=2000000000 EmissionAddress=nuklai1qqmzlnnredketlj3cu20v56nt5ken6thchra7nylwcrmz77td654w2jmpt9 EmissionBalance=594000014850
```

### Get Validators

We can check the validators that have been staked

```bash
./build/nuklai-cli emission validators
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
validator 0: NodeID=NodeID-7NEx57dgetsQdmgAGw9VjxAmTDWgdqZkP NodePublicKey=st7C9uuKbCM71BkP97PKTn8Vzn+ToE3giO/2z2OyNsY+FwVvIS2qtm6jPChANOXk UserStake=map[] StakedAmount=0 StakedReward=0
validator 1: NodeID=NodeID-4xW5x8ekVnhqJr5UnPQGRf8emeRiPDfRr NodePublicKey=o3INAYDajQSR2jWTQzq6K8ut0i5/gVXhfEbnDTQB1r2Fi07pU/r+v5rGiEEECp/C UserStake=map[] StakedAmount=0 StakedReward=0
validator 2: NodeID=NodeID-BrkTkxtNzhxP8przzBUnzZTXST8qMKCMV NodePublicKey=jvv4m5Si1w0TJjzPAVHbk7bcJEOEsGoGu2mjXrsuiVj56PsvVORG843MYhTjE2uW UserStake=map[] StakedAmount=0 StakedReward=0
validator 3: NodeID=NodeID-4h8Ud3qTEksLBCMNjkKBFhYRKJ5NfJZ2w NodePublicKey=hzGTs30HIPili1lgXjO5BfGJTZw9oAE6bOcoAuIQEMhNbI70d4aUm3B80GMyUyLu UserStake=map[] StakedAmount=0 StakedReward=0
validator 4: NodeID=NodeID-B2BSzCnzhAM2WwKU5UbwZxpoMPmE9TnNf NodePublicKey=pXYMPMHBASy0Y5/PzNI1B4HGrqSkP1upFtuSiEI2HKAbc+8ZPdkhmCWearjYOWRE UserStake=map[] StakedAmount=0 StakedReward=0
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
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
validators: 5
0: NodeID=NodeID-4h8Ud3qTEksLBCMNjkKBFhYRKJ5NfJZ2w NodePublicKey=hzGTs30HIPili1lgXjO5BfGJTZw9oAE6bOcoAuIQEMhNbI70d4aUm3B80GMyUyLu
1: NodeID=NodeID-B2BSzCnzhAM2WwKU5UbwZxpoMPmE9TnNf NodePublicKey=pXYMPMHBASy0Y5/PzNI1B4HGrqSkP1upFtuSiEI2HKAbc+8ZPdkhmCWearjYOWRE
2: NodeID=NodeID-7NEx57dgetsQdmgAGw9VjxAmTDWgdqZkP NodePublicKey=st7C9uuKbCM71BkP97PKTn8Vzn+ToE3giO/2z2OyNsY+FwVvIS2qtm6jPChANOXk
3: NodeID=NodeID-4xW5x8ekVnhqJr5UnPQGRf8emeRiPDfRr NodePublicKey=o3INAYDajQSR2jWTQzq6K8ut0i5/gVXhfEbnDTQB1r2Fi07pU/r+v5rGiEEECp/C
4: NodeID=NodeID-BrkTkxtNzhxP8przzBUnzZTXST8qMKCMV NodePublicKey=jvv4m5Si1w0TJjzPAVHbk7bcJEOEsGoGu2mjXrsuiVj56PsvVORG843MYhTjE2uW
validator to stake to: 2
balance: 852999899.999970317 NAI
✔ Staked amount: 1000█
✔ End LockUp Height: 930█
✔ continue (y/n): y█

✅ txID: xheUBZvoYNm2fWd8xEwvKKC2jwnr7FsrAAvNLTkcouM2LTwcD
```

### Get user staking info

We can retrieve our staking info by passing in which validator we have staked to and the address to look up staking for using the new RPC API we defined as part of this exercise.

```bash
./build/nuklai-cli emission user-stake-info
```

If successful, the output should be:

```
database: .nuklai-cli
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
validators: 5
0: NodeID=NodeID-B2BSzCnzhAM2WwKU5UbwZxpoMPmE9TnNf NodePublicKey=pXYMPMHBASy0Y5/PzNI1B4HGrqSkP1upFtuSiEI2HKAbc+8ZPdkhmCWearjYOWRE
1: NodeID=NodeID-7NEx57dgetsQdmgAGw9VjxAmTDWgdqZkP NodePublicKey=st7C9uuKbCM71BkP97PKTn8Vzn+ToE3giO/2z2OyNsY+FwVvIS2qtm6jPChANOXk
2: NodeID=NodeID-4xW5x8ekVnhqJr5UnPQGRf8emeRiPDfRr NodePublicKey=o3INAYDajQSR2jWTQzq6K8ut0i5/gVXhfEbnDTQB1r2Fi07pU/r+v5rGiEEECp/C
3: NodeID=NodeID-BrkTkxtNzhxP8przzBUnzZTXST8qMKCMV NodePublicKey=jvv4m5Si1w0TJjzPAVHbk7bcJEOEsGoGu2mjXrsuiVj56PsvVORG843MYhTjE2uW
4: NodeID=NodeID-4h8Ud3qTEksLBCMNjkKBFhYRKJ5NfJZ2w NodePublicKey=hzGTs30HIPili1lgXjO5BfGJTZw9oAE6bOcoAuIQEMhNbI70d4aUm3B80GMyUyLu
choose validator whom you have staked to: 1
address to get staking info for: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
user stake:  Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx StakedAmount=1000000000000
stake #1: TxID=xheUBZvoYNm2fWd8xEwvKKC2jwnr7FsrAAvNLTkcouM2LTwcD Amount=1000000000000 StartLockUp=315
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
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
assetID (use NAI for native token): NAI
✔ assetID (use NAI for native token): NAI█
balance: 852998899.999943018 NAI
```

We can unstake specific stake from a chosen validator.

```bash
./build/nuklai-cli action unstake-validator
```

Which produces result:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
validators: 5
0: NodeID=NodeID-4h8Ud3qTEksLBCMNjkKBFhYRKJ5NfJZ2w NodePublicKey=hzGTs30HIPili1lgXjO5BfGJTZw9oAE6bOcoAuIQEMhNbI70d4aUm3B80GMyUyLu
1: NodeID=NodeID-B2BSzCnzhAM2WwKU5UbwZxpoMPmE9TnNf NodePublicKey=pXYMPMHBASy0Y5/PzNI1B4HGrqSkP1upFtuSiEI2HKAbc+8ZPdkhmCWearjYOWRE
2: NodeID=NodeID-7NEx57dgetsQdmgAGw9VjxAmTDWgdqZkP NodePublicKey=st7C9uuKbCM71BkP97PKTn8Vzn+ToE3giO/2z2OyNsY+FwVvIS2qtm6jPChANOXk
3: NodeID=NodeID-4xW5x8ekVnhqJr5UnPQGRf8emeRiPDfRr NodePublicKey=o3INAYDajQSR2jWTQzq6K8ut0i5/gVXhfEbnDTQB1r2Fi07pU/r+v5rGiEEECp/C
4: NodeID=NodeID-BrkTkxtNzhxP8przzBUnzZTXST8qMKCMV NodePublicKey=jvv4m5Si1w0TJjzPAVHbk7bcJEOEsGoGu2mjXrsuiVj56PsvVORG843MYhTjE2uW
✔ validator to unstake from: 2█
stake info:
0: TxID=xheUBZvoYNm2fWd8xEwvKKC2jwnr7FsrAAvNLTkcouM2LTwcD StakedAmount=1000000000000 StartLockUpHeight=315 CurrentHeight=403
stake ID to unstake: 0 [auto-selected]
continue (y/n): y
✅ txID: 7zMSUymBJiEWFViMfGSGS2yWFyovDq1FJ5nPNthQmnoRc9G6w
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
