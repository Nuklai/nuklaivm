### Register a validator for staking

Even if there may be validators that are already taking part in the consensus of `nuklaivm` blocks, it doesn't mean they are automatically registered for the
staking mechanism. In order for a validator to register for staking on `nuklaivm`, they need to use the exact same account as they used while setting up the validator for
the Avalanche primary network.

```bash
./build/nuklai-cli action register-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: wkYgAYam1EB6Q7qPnRS8zMMaU13S27fYwgxS3fjGAH9JKHdUi
validators: 5
0: NodeID=NodeID-7jdCBX7cZkCTALUZVoCoAqLRhpa3jkw35 NodePublicKey=hcnVrQ12qQvH0Pb4IxNESVcYkztDn/kzn3WWrpRRZHvixmtIgatMjgDIJ8cRjOAB
1: NodeID=NodeID-JvQ7MZJd2M3xZFst86s2igP3jpyYRpDpD NodePublicKey=iz+92825d6rKcV7jGkEPAib28lySAHhI9X0Wqe6UM1Q2LBh5nnFKXCB0ak8c9b/S
2: NodeID=NodeID-5Eb8KR68tPHXEVK9gDXMvr7kfkUDuaZEz NodePublicKey=g2amgbfekrodLUDI08wOSqMnZ6jBZfVGFBL30AGmEVURDIpArmCdlOcbqbFftrIH
3: NodeID=NodeID-HQtHNjV4iabHpMaD7sEBTUrftHRsuZFGc NodePublicKey=jjnD22xVBzxwIDve4hDQfSSy79UGUG3zk/LRbemlzSrfPJ9YQYvb6dqYv3AHLnh3
4: NodeID=NodeID-6VH8eEDcUaaLDU3MeHPyQS78ctdSGwWEw NodePublicKey=if+D18Vt6DHScKdOnHyaWuegNFK7S12mIW3CCfWp6eia1jwI2Vsc+3xzByNtb7iY
validator to register for staking: 0
balance: 853000000.000000000 NAI
Staked amount: 100
✔ Staking Start Time(must be after 2024-02-15 15:53:44) [YYYY-MM-DD HH:MM:SS]: 2024-02-15 15:56:00█
✔ Staking End Time(must be after 2024-02-15 15:56:00) [YYYY-MM-DD HH:MM:SS]: 2024-02-15 15:57:00█
✔ Delegation Fee Rate(must be over 2): 50█
Reward Address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✔ continue (y/n): y█
✅ txID: FUQA9KV4z7JcPNNh6WiSHXdGYS6RSqMVNrxUVMhyfyrJgJJUT
```

### Get Validator stake info

You may want to check your staking info such as stake start time, stake end time, staked amount, delegation fee rate and reward address. To do so, you can do:

```bash
./build/nuklai-cli action get-validator-stake
```

If successful, the output should be something like:

```
database: .nuklai-cli
chainID: wkYgAYam1EB6Q7qPnRS8zMMaU13S27fYwgxS3fjGAH9JKHdUi
validators: 5
0: NodeID=NodeID-6VH8eEDcUaaLDU3MeHPyQS78ctdSGwWEw NodePublicKey=if+D18Vt6DHScKdOnHyaWuegNFK7S12mIW3CCfWp6eia1jwI2Vsc+3xzByNtb7iY
1: NodeID=NodeID-7jdCBX7cZkCTALUZVoCoAqLRhpa3jkw35 NodePublicKey=hcnVrQ12qQvH0Pb4IxNESVcYkztDn/kzn3WWrpRRZHvixmtIgatMjgDIJ8cRjOAB
2: NodeID=NodeID-JvQ7MZJd2M3xZFst86s2igP3jpyYRpDpD NodePublicKey=iz+92825d6rKcV7jGkEPAib28lySAHhI9X0Wqe6UM1Q2LBh5nnFKXCB0ak8c9b/S
3: NodeID=NodeID-5Eb8KR68tPHXEVK9gDXMvr7kfkUDuaZEz NodePublicKey=g2amgbfekrodLUDI08wOSqMnZ6jBZfVGFBL30AGmEVURDIpArmCdlOcbqbFftrIH
4: NodeID=NodeID-HQtHNjV4iabHpMaD7sEBTUrftHRsuZFGc NodePublicKey=jjnD22xVBzxwIDve4hDQfSSy79UGUG3zk/LRbemlzSrfPJ9YQYvb6dqYv3AHLnh3
validator to register for staking: 1
validator stake:
StakeStartTime=1708012560 StakeEndTime=1708012620 StakedAmount=100000000000 DelegationFeeRate=50 RewardAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx OwnerAddress=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
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
