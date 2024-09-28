# Nuklai Marketplace

## Create Your Dataset

### Create Dataset automatically

We can create our own dataset on nuklaichain.

We are going to use address #2 for this exercise:

```bash
0) address: 0027e7ca083ae74508d9fc9a858f874663f235e5d4010052dbca9d23e86a3c2232 balance: 12.000000000 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999887.999903798 NAI
2) address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a balance: 100.000000000 NAI
```

To do so, we do:

```bash
./build/nuklai-cli dataset create
```

When you are done, the output should look something like this:

```bash
address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
✔ name: dataset1█
description: desc1
isCommunityDataset (y/n): y
✔ metadata: test1█
✔ metadata: test1█

✔ continue (y/n): y█
✅ txID: 2bfvP2xhhn6pnSFH5eEU89drjNusQ8ff5EZHJBMvxzi23B3bZs
fee consumed: 0.000191100 NAI
output:  &{DatasetID:2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN DatasetParentNftID:NFBhsTjGBFbQL3pcK91cEVkwPBbsmXWGXQVfggy8dBrSt3kV9}
```

Note that creating a dataset will automatically create an asset of type "dataset" and also mint the parent NFT for the dataset as this type of asset is a fractionalized asset and always has a parent NFT and corresponding child NFTs.

## Publish the dataset to Nuklai marketplace

We can put up our dataset up for sale on the native nuklai marketplace at anytime. Note that once it's on the marketplace, no new data can be added to the dataset. This is to accurately calculate the reward being generated from the subscription so it can be paid out to all the current NFT owners of the dataset.

```bash
./build/nuklai-cli marketplace publish
```

You can set the asset that subscribers must use for payment and you can also dictate the price of accessing your dataset per block as shown below.

Output:

```bash
address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
datasetID: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN
assetForPayment (use NAI for native token): NAI
balance: 99.999808900 NAI
✔ continue (y/n): y█
✅ txID: SwEPkKVUor6hnmyeSy5uxCBershPJe9LrbrkkfNYewdTb5LGp
fee consumed: 0.000173000 NAI
output:  &{MarketplaceAssetID:oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1 AssetForPayment:11111111111111111111111111111111LpoYY DatasetPricePerBlock:1000000000 Publisher:00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a}
```

Upon successful transaction, it also creates a unique marketplace NFT collection with assetID `oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1`.
This asset has no owner and cannot be updated by anyone. Subsequent NFTs are automatically generated anytime a user subscribes to your dataset. Holding this NFT proves that the user is subscribed to your dataset and can access the dataset at anytime. This is different from the dataset NFTs that the dataset contributors own which are used for claiming the payment from all the subscriptions.

## View dataset info from the marketplace

We can also look up more info about the dataset that is in the marketplace.

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
datasetID: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN
Retrieving dataset info from the marketplace: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN

dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetID=oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1 AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a

marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-datas AssetURI=2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN TotalSupply=0 MaxSupply=0 Owner=000000000000000000000000000000000000000000000000000000000000000000
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "datasetID":"2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"0", "marketplaceAssetID":"oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1", "paymentClaimed":"0", "paymentRemaining":"0", "publisher":"00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a", "subscriptions":"0"}
```

This displays summary of the dataset that includes the saleID in the marketplace along with the asset used for payment and price to access dataset per block. It also shows the unique metadata for this marketplace NFT collection with asset ID `oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1`

## Subscribe to the dataset in the marketplace

Let's switch to another account and subscribe to this dataset.

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
stored keys: 3
0) address: 0027e7ca083ae74508d9fc9a858f874663f235e5d4010052dbca9d23e86a3c2232 balance: 12.000000000 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999887.999903798 NAI
2) address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a balance: 99.999635900 NAI
✔ set default key: 0█
```

We are going to use our address #0 for this exercise.

Now, let's subscribe to this dataset for 5 blocks:

```bash
 ./build/nuklai-cli marketplace subscribe
```

Output:

```bash
Name=dataset1 Description=desc1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test1 IsCommunityDataset=true SaleID=oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1 BaseAsset=11111111111111111111111111111111LpoYY BasePrice=1000000000 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a
✔ assetForPayment (use NAI for native token): NAI█
numBlocksToSubscribe: 5
balance: 12.000000000 NAI
continue (y/n): y
✅ txID: 2BuWFzhtXs5sJj8jQqE4F1QLTR73GpaJF2Cg3wt6BHexCHSMoA
fee consumed: 0.000179200 NAI
output:  &{MarketplaceAssetID:oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1 MarketplaceAssetNumSubscriptions:1 SubscriptionNftID:12zPr21woaw1pLPNJatW7jNj4rF1k5NCoWAiaxsZk8Z9hpZZm AssetForPayment:11111111111111111111111111111111LpoYY DatasetPricePerBlock:1000000000 TotalCost:5000000000 NumBlocksToSubscribe:5 IssuanceBlock:337 ExpirationBlock:342}
```

Let's check the balance as it should automatically take 5 \* pricePerBlock from our account:

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
stored keys: 3
0) address: 0027e7ca083ae74508d9fc9a858f874663f235e5d4010052dbca9d23e86a3c2232 balance: 6.999820800 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999887.999903798 NAI
2) address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a balance: 99.999635900 NAI
set default key: 0
```

In addition, we also got issued an NFT for the above mentioned marketplace token which is unique to this dataset. This NFT is used to prove that the user is subscribed to the dataset in the marketplace.

```bash
./build/nuklai-cli key balance-nft
```

Output:

```bash
address: 0027e7ca083ae74508d9fc9a858f874663f235e5d4010052dbca9d23e86a3c2232
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetID: 12zPr21woaw1pLPNJatW7jNj4rF1k5NCoWAiaxsZk8Z9hpZZm
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
collectionID: oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1
uniqueID: 1
uri: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN
metadata: {"assetForPayment":"11111111111111111111111111111111LpoYY","datasetID":"2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN","datasetPricePerBlock":"1000000000","expirationBlock":"342","issuanceBlock":"337","marketplaceAssetID":"oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1","numBlocksToSubscribe":"5","totalCost":"5000000000"}
ownerAddress: 0027e7ca083ae74508d9fc9a858f874663f235e5d4010052dbca9d23e86a3c2232
You own this NFT
```

This NFT contains details on how much the user paid to subscribe to the dataset and when does the access expire.

Now, let's check the info about the dataset from the marketplace again. The metadata should be updated:

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
✔ datasetID: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN█
Retrieving dataset info from the marketplace: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN

dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetID=oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1 AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a

marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-datas AssetURI=2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN TotalSupply=1 MaxSupply=0 Owner=000000000000000000000000000000000000000000000000000000000000000000
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "datasetID":"2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"337", "marketplaceAssetID":"oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1", "paymentClaimed":"0", "paymentRemaining":"5000000000", "publisher":"00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a", "subscriptions":"1"}
```

Now, it shows the number of subscriptions is 1 instead of 0 and various other details have also been updated.

## Claim accumulated subscription rewards

As an owner, whenever a user subscribes to your dataset, they pay a certain amount based on however many blocks they subscribe to. This payment is not done to the owner instantly but rather the blockchain holds the money and releases the payment slowly over time. This is to prevent a case whereby a user subscribes to the dataset but the dataset is not available to the subscribed user. The payment is accumulated every epoch(around x blocks) based on the liveliness of the dataset data(Currently not implemented).

Let's check our balance before claiming the payment

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
stored keys: 3
0) address: 0027e7ca083ae74508d9fc9a858f874663f235e5d4010052dbca9d23e86a3c2232 balance: 6.999820800 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999887.999903798 NAI
2) address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a balance: 99.999635900 NAI
set default key: 2
```

```bash
./build/nuklai-cli marketplace claim-payment
```

Output:

```bash
address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
datasetID: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN
Retrieving dataset info from the marketplace: 2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN
dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetID=oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1 AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a
marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-datas AssetURI=2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN TotalSupply=1 MaxSupply=0 Owner=000000000000000000000000000000000000000000000000000000000000000000
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "datasetID":"2bovg98jMSdt9XGb2CedUr8uHZ56cVZaMLPxdn1HtQ8usgaJHN", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"337", "marketplaceAssetID":"oYmC6Ep2w9NQaNzXN4oLNPMVB3CVr1NmsRsHaV34Z14T7BLT1", "paymentClaimed":"0", "paymentRemaining":"5000000000", "publisher":"00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a", "subscriptions":"1"}
✔ continue (y/n): y█
✅ txID: 25rKXbBvLjC61YPkffneh9rDeVgK6mRuyEM3XgKMcRbBj3hYnF
fee consumed: 0.000175400 NAI
output:  &{LastClaimedBlock:482 PaymentClaimed:5000000000 PaymentRemaining:0 DistributedReward:5000000000 DistributedTo:00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a}
```

Let's ensure that we are rewarded with the payment:

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
stored keys: 3
0) address: 0027e7ca083ae74508d9fc9a858f874663f235e5d4010052dbca9d23e86a3c2232 balance: 6.999820800 NAI
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999887.999903798 NAI
2) address: 00fa92500595699234176c32afbf5c6558df21deb10ba4d2d691e5e5148658c64a balance: 104.999460500 NAI
set default key: 2
```

Looks like we got 5 NAI as expected. Previously, our balance was `99.999635900 NAI` and now, our new balance is `104.999460500 NAI`
