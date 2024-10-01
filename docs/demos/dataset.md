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
name: dataset1
description: desc1
isCommunityDataset (y/n): y
metadata: test1
✔ continue (y/n): y█
✅ txID: Z2UMDbY5U9g7CsFBcKcC2VTSL7Tqycb6H6ZeGJ58kH8AR6rVG
fee consumed: 0.000191100 NAI
output:  &{DatasetID:WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx DatasetParentNftID:2YkmGsJSiysEEa7PyXGZTknB3UQrygtginpDWs5MSeqWgt7iKt}
```

Note that creating a dataset will automatically create an asset of type "dataset" and also mint the parent NFT for the dataset as this type of asset is a fractionalized asset and always has a parent NFT and corresponding child NFTs.

### Create Dataset with existing asset

First, let's create our asset.

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 2
✔ name: dataset2█
symbol: ds2
decimals: 0
metadata: test2
✔ continue (y/n): y█
✅ txID: 2Q6mm4LCxR8dBxMRvrtsjTegEBvLWfCGZPA3TCH68RF2bGiakV
fee consumed: 0.000082200 NAI
output:  &{AssetID:2jZ1ajDTMKy6FNsA6GqQRJfRn9jr9YUx8bnpUJcqR1W1QYN4D3 AssetBalance:1 DatasetParentNftID:2Rdto817HWG4ii38WJvVWj67WqrUCQZkmyXyxMiiFiFvVxem7P}
```

Now, let's create our dataset using this asset.

```bash
./build/nuklai-cli dataset create-from-asset
```

When you are done, the output should look something like this:

```bash
assetID: 2jZ1ajDTMKy6FNsA6GqQRJfRn9jr9YUx8bnpUJcqR1W1QYN4D3
name: datset2
description: desc2
✔ isCommunityDataset (y/n): y█
metadata: test2
✔ continue (y/n): y█
✅ txID: 3dCN894iZoc1BJXPo4D9XPS3mWXmmPexT5rjWMQjxbyycGNN1
fee consumed: 0.000190900 NAI
output:  &{DatasetID:2jZ1ajDTMKy6FNsA6GqQRJfRn9jr9YUx8bnpUJcqR1W1QYN4D3 DatasetParentNftID:2Rdto817HWG4ii38WJvVWj67WqrUCQZkmyXyxMiiFiFvVxem7P}
```

## View Dataset Info

### View dataset details

Now, let's retrieve the info about our dataset.

```bash
./build/nuklai-cli dataset info
```

When you are done, the output should look something like this:

```bash
Retrieving dataset info for datasetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx█
dataset info:
Name=dataset1 Description=desc1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test1 IsCommunityDataset=true SaleID=11111111111111111111111111111111LpoYY BaseAsset=11111111111111111111111111111111LpoYY BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

Retrieving asset info for assetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
assetType:  Dataset Token name: dataset1 symbol: dataset1 decimals: 0 metadata: desc1 uri: desc1 totalSupply: 1 maxSupply: 0 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
balance: 1 dataset1
```

### View balance of the assets

Since nuklai creates an asset while creating the dataset, let's check the balance of these assets.

Since the dataset is also a fractionalized asset, we can check its balance doing:

```bash
./build/nuklai-cli key balance-ft
```

When you are done, the output should look something like this:

```bash
✔ assetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx█
assetType:  Dataset Token name: dataset1 symbol: dataset1 decimals: 0 metadata: desc1 uri: desc1 totalSupply: 1 maxSupply: 0 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
balance: 1 dataset1
```

We can also check info about the parent NFT that was minted when creating the dataset.

```bash
./build/nuklai-cli key balance-nft
```

The output should be something like:

```bash
collectionID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
uniqueID: 0
uri: desc1
metadata: desc1
ownerAddress: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
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
stored keys: 2
0) address: 00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4 balance: 0 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999999.999503732 NAI
set default key: 1
```

You need to submit some amount of NAI(this is set in the VM config) as collateral when starting the contribution process so let's send some NAI to this new account first.

```bash
0) address: 00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4 balance: 50.000000000 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999949.999776721 NAI
set default key: 0
```

Now, let's start the contribution process.

```bash
 ./build/nuklai-cli dataset initiate-contribute
```

When you are done, the output should look something like this:

```bash
datasetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
dataIdentifier: id1
✔ continue (y/n): y█
✅ txID: 2ReNPSR3GW9rWVaVg7joJ31rW7q5qkYwoUJPh4zTnJJYLCdPFc
fee consumed: 0.000132800 NAI
output:  &{CollateralAssetID:11111111111111111111111111111111LpoYY CollateralAmountTaken:1000000000}
```

Note that your balance may have decreased a bit. You will get it refunded once the dataset owner completes the contribution process.

```bash
address: 00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4 balance: 48.999867200 NAI
```

### Step 2: View contribution info

We can now check more details about this pending contribution.

```bash
./build/nuklai-cli dataset contribute-info
```

Output:

```bash
datasetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
Retrieving pending data contributions info for datasetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
Contribution 0: Contributor=00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4 DataLocation=default DataIdentifier=id1
```

### Step 3: Complete Contribute

Now, let's switch to the dataset owner account and complete this 2-step process.

```bash
./build/nuklai-cli dataset complete-contribute
```

Output:

```bash
✔ datasetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx█
contributor: 00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4
✔ unique nft #: 1█
continue (y/n): y
✅ txID: 2hpemTb16BqsNEh5iHGf25seGqJpmxRLZUnocRRCvthaQwczn
fee consumed: 0.000180600 NAI
output:  &{CollateralAssetID:11111111111111111111111111111111LpoYY CollateralAmountRefunded:1000000000 DatasetID:WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx DatasetChildNftID:2qdrhZS8WYiHtwZ8jw5kYS3YmL7PUmVv7kz3hYdHamupSQmoPv To:00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4 DataLocation:[100 101 102 97 117 108 116] DataIdentifier:[105 100 49]}
```

We should now have our collateral refunded back to us.

```bash
./build/nuklai-cli key set
```

Output:

```bash
0) address: 00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4 balance: 49.999867200 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999949.999596119 NAI
set default key: 0
```

This should also issue an NFT to the contributor so let's switch back to our contributor account and check the dataset info.

```bash
./build/nuklai-cli dataset info
```

Output:

```bash
Retrieving dataset info for datasetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
dataset info:
Name=dataset1 Description=desc1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test1 IsCommunityDataset=true SaleID=11111111111111111111111111111111LpoYY BaseAsset=11111111111111111111111111111111LpoYY BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

Retrieving asset info for assetID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
assetType:  Dataset Token name: dataset1 symbol: dataset1 decimals: 0 metadata: desc1 uri: desc1 totalSupply: 2 maxSupply: 0 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
balance: 1 dataset1
```

We can also check that we have been issued an NFT:

```bash
./build/nuklai-cli key balance-nft
```

Output:

```bash
collectionID: WRTzP6mQECmCwtLX7Gd72rkokWEcStRL9M89GJ4YHo1ZoLhFx
uniqueID: 1
uri: desc1
metadata: {"dataIdentifier":"id1","dataLocation":"default"}
ownerAddress: 00095ea19193bd18b48ea58137a6ec9bc0fbdedb7fd7c5b078df99d881f8d407c4
You own this NFT
```
