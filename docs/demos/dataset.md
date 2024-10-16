# Datasets

## Create Your Dataset

### Create Dataset automatically

We can create our own dataset on nuklaichain.

First, let's create our asset.

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
assetType(0 for fungible, 1 for non-fungible and 2 for fractional): 2
): 2█
name: dataset1
symbol: ds1
metadata: test1
continue (y/n): y
✅ txID: sUBFBLbKbP8KuoUBsHKMB4qEu9efL1kKoUmxnj3pWKN1h7yDN
fee consumed: 0.000096200 NAI
output:  &{AssetAddress:02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 AssetBalance:1 DatasetParentNftAddress:016cf1cfff3f4c1aa8b081376c0119b1a16b82466c75d174d6ff88cf93444b50dd}
```

Since it's a fractionalized asset type, nuklai automatically mints the parent NFT at the same time.

Now, let's create our dataset using this asset.

```bash
./build/nuklai-cli dataset create
```

When you are done, the output should look something like this:

```bash
assetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7
name: dataset1
description: desc1
isCommunityDataset (y/n): y
metadata: test1
continue (y/n): y
✅ txID: GLFRqKmfKfP1pLo73fLVv7RuD9B1c9yNPXNWWqF6Z2Fd7tbER
fee consumed: 0.000176300 NAI
output:  &{DatasetAddress:02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 DatasetParentNftAddress:016cf1cfff3f4c1aa8b081376c0119b1a16b82466c75d174d6ff88cf93444b50dd}
```

Note that creating a dataset will automatically create an asset of type "dataset" and also mint the parent NFT for the dataset as this type of asset is a fractionalized asset and always has a parent NFT and corresponding child NFTs.

## View Dataset Info

### View dataset details

Now, let's retrieve the info about our dataset.

```bash
./build/nuklai-cli dataset info
```

When you are done, the output should look something like this:

```bash
datasetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7

Retrieving dataset info for datasetID: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7
dataset info:
Name=dataset1 Description=desc1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test1 IsCommunityDataset=true SaleID=000000000000000000000000000000000000000000000000000000000000000000 BaseAsset=000000000000000000000000000000000000000000000000000000000000000000 BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

Retrieving asset info for assetID: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7
assetType:  Fractional Token name: dataset1 symbol: ds1 decimals: 0 metadata: test1 uri: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 totalSupply: 1 maxSupply: 0 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 1 ds1
```

### View balance of the assets

Since nuklai creates an asset while creating the dataset, let's check the balance of these assets.

Since the dataset is also a fractionalized asset, we can check its balance doing:

```bash
./build/nuklai-cli key balance-asset
```

When you are done, the output should look something like this:

```bash
assetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 1 ds1
```

We can also check info about the parent NFT that was minted when creating the dataset.

```bash
./build/nuklai-cli key nft
```

The output should be something like:

```bash
assetAddress: 016cf1cfff3f4c1aa8b081376c0119b1a16b82466c75d174d6ff88cf93444b50dd
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
assetType:  Non-Fungible Token name: dataset1 symbol: ds1-0 metadata: test1 collectionAssetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
collectionAssetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 balance: 1 ds1-0
You own this NFT
```

## Contribute data to dataset

There are two steps involved when adding data to a dataset. Anyone can start the initiation process however, only the owner of the dataset can complete the contribution

### Step 1: Initiate Contribute

Let's switch to a different account and start the inititation process. You need to submit some amount of NAI(this is set in the VM config) as collateral when starting the contribution process so let's send some NAI to this new account first.

```bash
./build/nuklai-cli key set
```

Output:

```bash
stored keys: 2
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
0) address: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 balance: 100.000000000 NAI
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999899.999678969 NAI
set default key: 0
```

Now, let's start the contribution process.

```bash
 ./build/nuklai-cli dataset initiate-contribute
```

When you are done, the output should look something like this:

```bash
datasetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7
dataIdentifier: id1
continue (y/n): y
✅ txID: otMZkkq9TSbtjqAgj9FAy1VTTvVGuzEUVGYyPEJG6ezzRcqcq
fee consumed: 0.000153500 NAI
output:  &{DatasetContributionID:kUH3D98AGED3tYgbfHjzfnojZydsFAqUcQJ4wcQFvvxHVcRBt CollateralAssetAddress:00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb CollateralAmountTaken:1000000000}
```

Note that your balance may have decreased a bit. You will get it refunded once the dataset owner completes the contribution process.

```bash
0) address: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 balance: 98.999846500 NAI
```

### Step 2: View contribution info

We can now check more details about this pending contribution.

```bash
./build/nuklai-cli dataset contribute-info
```

Output:

```bash
✔ contributionID: kUH3D98AGED3tYgbfHjzfnojZydsFAqUcQJ4wcQFvvxHVcRBt
contribution info:
DatasetAddress=02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 DataLocation=default DataIdentifier=id1 Contributor=007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 ContributionAcceptedByDatasetOwner=false
```

### Step 3: Complete Contribute

Now, let's switch to the dataset owner account and complete this 2-step process.

```bash
./build/nuklai-cli dataset complete-contribute
```

Output:

```bash
✔ datasetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7█
contributionID: kUH3D98AGED3tYgbfHjzfnojZydsFAqUcQJ4wcQFvvxHVcRBt
contributor: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145
continue (y/n): y
✅ txID: 2NS6GM96GdiQyL9uur5PAXmpi4bgAhLKAnAuMRvw4YBYqDxweu
fee consumed: 0.000204100 NAI
output:  &{CollateralAssetAddress:00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb CollateralAmountRefunded:1000000000 DatasetChildNftAddress:0262b460a21ef30ea647b549077fa30382d053d6d9c1783c8e1058b076b853eb07 To:007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 DataLocation:default DataIdentifier:id1}
```

We should now have our collateral refunded back to us.

```bash
./build/nuklai-cli key set
```

Output:

```bash
stored keys: 2
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
0) address: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 balance: 99.999646900 NAI
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999899.999475002 NAI
set default key: 0
```

This should also issue an NFT to the contributor so let's switch back to our contributor account and check the dataset info.

```bash
./build/nuklai-cli dataset info
```

Output:

```bash
datasetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7

dataset info:
Name=dataset1 Description=desc1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test1 IsCommunityDataset=true SaleID=000000000000000000000000000000000000000000000000000000000000000000 BaseAsset=000000000000000000000000000000000000000000000000000000000000000000 BasePrice=0 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

Retrieving asset info for assetID: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7
assetType:  Fractional Token name: dataset1 symbol: ds1 decimals: 0 metadata: test1 uri: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 totalSupply: 2 maxSupply: 0 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
address: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 balance: 1 ds1
```

We can also check that we have been issued an NFT:

```bash
./build/nuklai-cli key nft
```

Output:

```bash
assetAddress: 0262b460a21ef30ea647b549077fa30382d053d6d9c1783c8e1058b076b853eb07
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
assetType:  Non-Fungible Token name: dataset1 symbol: ds1-1 metadata: {"dataIdentifier":"id1","dataLocation":"default"} collectionAssetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 owner: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145
collectionAssetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 balance: 1 ds1-1
You own this NFT
```
