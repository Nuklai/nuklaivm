# Datasets

## Create Your Dataset

### Create Dataset automatically

We can create our own dataset on nuklaichain.

To do so, we do:

```bash
./build/nuklai-cli dataset create
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
name: dataset1
description: desc for dataset1
✔ isCommunityDataset (y/n): y█
metadata: metadata for dataset1
continue (y/n): y
✅ txID: 2vi4DzscWo7NJ5ZNMrxyj58qgGPNyPXHxhdkuYcSnPfUCjLhpn
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
assetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
nftID: Aauzfwi9o3wAE8JuxhBvrry1XqeAqE3ysZcHqyha8WA2b1s9N
```

Note that creating a dataset will automatically create an asset of type "dataset" and also mint the parent NFT for the dataset as this type of asset is a fractionalized asset and always has a parent NFT and corresponding child NFTs.

### Create Dataset with existing asset

First, let's create our asset.

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 2✔ assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 2█
name: Kiran
symbol: KP4
✔ decimals: 0█
metadata: test
continue (y/n): y
✅ txID: 2pwpuuz1fbAbXndHbYzY4oczVDRkKcP69TdKc8bc1c6xQrFZQF
assetID: YywWmuyTGzDAPYPKNgXy4yk4gbjbNQU7qjyCD3RMuoAoWazsM
nftID: ZXvW9q7sFS3XzTmK2uQ4T6M6dngdGG3engYBJH9rrrki4HMpu
```

Now, let's create our dataset using this asset.

```bash
./build/nuklai-cli dataset create-from-asset
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetID (use NAI for native token): YywWmuyTGzDAPYPKNgXy4yk4gbjbNQU7qjyCD3RMuoAoWazsM
name: Kiran
description: desc for kiran
isCommunityDataset (y/n): y
metadata: test
continue (y/n): y
✅ txID: 2ahkDeeJGDJd3ptQXGwQtC5KQ7iFjGUCfNkjqUi6ayyee7uwLL
datasetID: YywWmuyTGzDAPYPKNgXy4yk4gbjbNQU7qjyCD3RMuoAoWazsM
assetID: YywWmuyTGzDAPYPKNgXy4yk4gbjbNQU7qjyCD3RMuoAoWazsM
nftID: ZXvW9q7sFS3XzTmK2uQ4T6M6dngdGG3engYBJH9rrrki4HMpu
```

## View Dataset Info

### View dataset details

Now, let's retrieve the info about our dataset.

```bash
./build/nuklai-cli dataset info
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
Retrieving dataset info for datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
dataset info:
Name=dataset1 Description=desc for dataset1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=metadata for dataset1 IsCommunityDataset=true SaleID=11111111111111111111111111111111LpoYY BaseAsset=11111111111111111111111111111111LpoYY BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
Retrieving asset info for assetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
assetType:  Dataset Token name: dataset1 symbol: dataset1 decimals: 0 metadata: desc for dataset1 uri: desc for dataset1 totalSupply: 1 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 1 dataset1
```

### View balance of the assets

Since nuklai creates an asset while creating the dataset, let's check the balance of these assets.

Since the dataset is also a fractionalized asset, we can check its balance doing:

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
assetID (use NAI for native token): 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
✔ assetID (use NAI for native token): 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH█
assetType:  Dataset Token name: dataset1 symbol: dataset1 decimals: 0 metadata: desc for dataset1 uri: desc for dataset1 totalSupply: 1 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 1 dataset1
```

We can also check info about the parent NFT that was minted when creating the dataset.

```bash
./build/nuklai-cli key balanceNFT
```

The output should be something like:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
nftID: Aauzfwi9o3wAE8JuxhBvrry1XqeAqE3ysZcHqyha8WA2b1s9N
collectionID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
uniqueID: 0
uri: desc for dataset1
metadata: desc for dataset1
ownerAddress: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
You own this NFT
```

## Contribute data to dataset

There are two steps involved when adding data to a dataset. Anyone can start the initiation process however, only the owner of the dataset can complete the contribution

### Step 1: Initiate Contribute

Let's switch to a different account and start the inititation process.

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999949.999773979 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 0.000000000 NAI
✗ set default key:
```

You need to submit some amount of NAI(this is set in the VM config) as collateral when starting the contribution process so let's send some NAI to this new account first.

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999939.999741912 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 10.000000000 NAI
set default key: 1
```

Now, let's start the contribution process.

```bash
 ./build/nuklai-cli dataset initiate-contribute
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
✔ datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH█
dataIdentifier: idrow1
continue (y/n): y
✅ txID: 2AD8PjZx8ZYdU2ENN8q5Nnn7bnG6tUQTEvSEF2FEdWxydmWDPo
```

Note that your balance may have decreased a bit. You will get it refunded once the dataset owner completes the contribution process.

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999939.999741912 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 8.999867900 NAI
```

### Step 2: View contribution info

We can now check more details about this pending contribution.

```bash
./build/nuklai-cli dataset info
```

Output:

```bash
address: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
Retrieving dataset info for datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
dataset info:
Name=dataset1 Description=desc for dataset1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=metadata for dataset1 IsCommunityDataset=true SaleID=11111111111111111111111111111111LpoYY BaseAsset=11111111111111111111111111111111LpoYY BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
Retrieving asset info for assetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
assetType:  Dataset Token name: dataset1 symbol: dataset1 decimals: 0 metadata: desc for dataset1 uri: desc for dataset1 totalSupply: 1 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
assetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
name: dataset1
symbol: dataset1
balance: 0
please send funds to nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
exiting...
```

### Step 3: Complete Contribute

Now, let's switch to the dataset owner account and complete this 2-step process.

```bash
./build/nuklai-cli dataset complete-contribute
```

Output:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
contributor: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
unique nft #: 1
✔ continue (y/n): y█
✅ txID: 2UMmEhyp1oV2f6nFBab9LRNaBXFsj7HLseWb32bMtnsqbYvp2U
nftID: A4tQnLCsyafbQPqLQdL8c7sQUfB4Wj36ZX6Nh3HdyUfSs4ABR
```

We should now have our collateral refunded back to us.

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999939.999561310 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 9.999867900 NAI
set default key: 1
```

This should also issue an NFT to the contributor so let's switch back to our contributor account and check the dataset info.

```bash
./build/nuklai-cli dataset info
```

Output:

```bash
address: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
✔ datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH█
dataset info:
Name=dataset1 Description=desc for dataset1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=metadata for dataset1 IsCommunityDataset=true SaleID=11111111111111111111111111111111LpoYY BaseAsset=11111111111111111111111111111111LpoYY BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
Retrieving asset info for assetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
assetType:  Dataset Token name: dataset1 symbol: dataset1 decimals: 0 metadata: desc for dataset1 uri: desc for dataset1 totalSupply: 2 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 1 dataset1
```

We can also check that we have been issued an NFT:

```bash
./build/nuklai-cli key balanceNFT
```

Output:

```bash
address: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
nftID: A4tQnLCsyafbQPqLQdL8c7sQUfB4Wj36ZX6Nh3HdyUfSs4ABR
collectionID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
uniqueID: 1
uri: desc for dataset1
metadata: {"dataLocation":"default","dataIdentifier":"idrow1"}
ownerAddress: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
You own this NFT
```
