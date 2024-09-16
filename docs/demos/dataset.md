# Datasets

## Step 1: Create Your Dataset

### Create Dataset automatically

We can create our own dataset on nuklaichain.

To do so, we do:

```bash
./build/nuklai-cli dataset create
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✔ name: Kiran█
✔ description: dataset for kiran█
✔ metadata: test metadata█
continue (y/n): y
✅ txID: 2gLhw1JskkvjBDyXHrpE78hhAczue25ap3QH6k9yL8mitU1MZ3
datasetID: 9sqNcSVxH9ZEDxwgpExLbLx8b6NGEtetk8RokmHnJMGyxua4F
assetID: 9sqNcSVxH9ZEDxwgpExLbLx8b6NGEtetk8RokmHnJMGyxua4F
nftID: NZvonnvW7yV8j2tpV7cMyrAPyFzXVSrQ8qauHfJBxoYaNxvB3
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

## Step 2: View Dataset Info

### View dataset details

Now, let's retrieve the info about our dataset.

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
datasetID: YywWmuyTGzDAPYPKNgXy4yk4gbjbNQU7qjyCD3RMuoAoWazsM
Retrieving dataset info for datasetID: YywWmuyTGzDAPYPKNgXy4yk4gbjbNQU7qjyCD3RMuoAoWazsM
dataset info:
Name=Kiran Description=desc for kiran Categories=Kiran LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test IsCommunityDataset=true OnSale=false BaseAsset=11111111111111111111111111111111LpoYY BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
Retrieving asset info for assetID: YywWmuyTGzDAPYPKNgXy4yk4gbjbNQU7qjyCD3RMuoAoWazsM
assetType:  Dataset Token name: Kiran symbol: KP4 decimals: 0 metadata: test uri: https://nukl.ai totalSupply: 1 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 1 KP4
```

### View balance of the assets

Since nuklai creates an asset while creating the dataset, let's check the balance of these assets.

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
✔ datasetID: 9sqNcSVxH9ZEDxwgpExLbLx8b6NGEtetk8RokmHnJMGyxua4F█
Retrieving dataset info for datasetID: 9sqNcSVxH9ZEDxwgpExLbLx8b6NGEtetk8RokmHnJMGyxua4F
dataset info:
Name=Kiran Description=dataset for kiran Categories=Kiran LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test metadata IsCommunityDataset=false OnSale=false BaseAsset=11111111111111111111111111111111LpoYY BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=100 RevenueModelMetadataOwnerCut=0 Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
Retrieving asset info for assetID: 9sqNcSVxH9ZEDxwgpExLbLx8b6NGEtetk8RokmHnJMGyxua4F
assetType: {{/} Dataset Token name: Kiran symbol: Kiran decimals: 0 metadata: dataset for kiran uri: dataset for kiran totalSupply: 1 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 1 Kiran
```

Since the dataset is also a fractionalized asset, we can check its balance doing:

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetID (use NAI for native token): 9sqNcSVxH9ZEDxwgpExLbLx8b6NGEtetk8RokmHnJMGyxua4F
uri: http://127.0.0.1:9658/ext/bc/24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetType: {{/} Dataset Token name: Kiran symbol: Kiran decimals: 0 metadata: dataset for kiran uri: dataset for kiran totalSupply: 1 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 1 Kiran
```

We can also check info about the parent NFT that was minted when creating the dataset.

```bash
./build/nuklai-cli key balanceNFT
```

The output should be something like:

```bash
✔ address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx█
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
✔ nftID: NZvonnvW7yV8j2tpV7cMyrAPyFzXVSrQ8qauHfJBxoYaNxvB3█
collectionID: 9sqNcSVxH9ZEDxwgpExLbLx8b6NGEtetk8RokmHnJMGyxua4F
uniqueID: 0
uri: dataset for kiran
ownerAddress: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
You own this NFT
```
