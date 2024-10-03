# Nuklai Marketplace

## Create Your Dataset

### Create Dataset automatically

We can create our own dataset on nuklaichain.

First, let's create our asset.

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 2
✔ name: dataset1
symbol: ds1
decimals: 0
metadata: test
✔ continue (y/n): y█
✅ txID: 2jguHxmfowHxhwoZUyF1deoN12aHvcbDeBZTfgDAWVabksHRYQ
fee consumed: 0.000085300 NAI
output:  &{AssetID:2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE AssetBalance:1 DatasetParentNftID:242a2eZpoSBKLcj4waVyueCBQFXrf5aCpEqt3YvrfbLLAFpNEi}
```

Now, let's create our dataset using this asset.

```bash
./build/nuklai-cli dataset create
```

When you are done, the output should look something like this:

```bash
assetID: 2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE
name: dataset1
✔ description: desc1█
✔ isCommunityDataset (y/n): y█
✔ metadata: test1█
continue (y/n): y
): y█
✅ txID: 41cK1RXoAe2F72NnoyHujvBAxvthGs8t86CLMuNRpFJrn1PJD
fee consumed: 0.000173100 NAI
output:  &{DatasetID:2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE DatasetParentNftID:242a2eZpoSBKLcj4waVyueCBQFXrf5aCpEqt3YvrfbLLAFpNEi}
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
datasetID: 2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE
assetForPayment (use NAI for native token): NAI
balance: 852999899.999471903 NAI
priceAmountPerBlock: 1
continue (y/n): y
✅ txID: 2uN1xEPbDvsS8NaW4tfX7JkVFwxEQTYTphqMRppvGb66kjUp6R
fee consumed: 0.000176200 NAI
output:  &{MarketplaceAssetID:Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ AssetForPayment:11111111111111111111111111111111LpoYY DatasetPricePerBlock:1000000000 Publisher:00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9}
```

Upon successful transaction, it also creates a unique marketplace NFT collection with assetID `Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ`.
This asset has no owner and cannot be updated by anyone. Subsequent NFTs are automatically generated anytime a user subscribes to your dataset. Holding this NFT proves that the user is subscribed to your dataset and can access the dataset at anytime. This is different from the dataset NFTs that the dataset contributors own which are used for claiming the payment from all the subscriptions.

## View dataset info from the marketplace

We can also look up more info about the dataset that is in the marketplace.

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
Retrieving dataset info from the marketplace: 2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetID=Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-ds1 AssetURI=2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE TotalSupply=0 MaxSupply=0 Owner=000000000000000000000000000000000000000000000000000000000000000000
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "datasetID":"2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"0", "marketplaceAssetID":"Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ", "paymentClaimed":"0", "paymentRemaining":"0", "publisher":"00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9", "subscriptions":"0"}
```

This displays summary of the dataset that includes the saleID in the marketplace along with the asset used for payment and price to access dataset per block. It also shows the unique metadata for this marketplace NFT collection with asset ID `Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ`

## Subscribe to the dataset in the marketplace

Let's switch to another account and subscribe to this dataset.

```bash
./build/nuklai-cli key set
```

Output:

```bash
0) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999899.999295712 NAI
1) address: 00cc39cb91242b1ff0968b6e6184f349c903c37f1462b90809c5a536dc1720ef0e balance: 100.000000000 NAI
set default key: 1
```

We are going to use our address #0 for this exercise.

Now, let's subscribe to this dataset for 5 blocks:

```bash
 ./build/nuklai-cli marketplace subscribe
```

Output:

```bash
✔ datasetID: 2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE█
dataset info:
Name=dataset1 Description=desc1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=test1 IsCommunityDataset=true SaleID=Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ BaseAsset=11111111111111111111111111111111LpoYY BasePrice=1000000000 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Own✔ assetForPayment (use NAI for native token): NAI█
numBlocksToSubscribe: 10
balance: 100.000000000 NAI
✔ continue (y/n): y█
✅ txID: 2nZccHHanyu2QLJtnhF46NmfHvh5sSK2H6jyr19zkupj83AuxF
fee consumed: 0.000179200 NAI
output:  &{MarketplaceAssetID:Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ MarketplaceAssetNumSubscriptions:1 SubscriptionNftID:2aU4VENsDqD3jvfPz1vHBVF3siZmxAQXtg9w7XCvzzpkBjxh36 AssetForPayment:11111111111111111111111111111111LpoYY DatasetPricePerBlock:1000000000 TotalCost:10000000000 NumBlocksToSubscribe:10 IssuanceBlock:890 ExpirationBlock:900}
```

Let's check the balance as it should automatically take 5 \* pricePerBlock from our account:

```bash
./build/nuklai-cli key balance
```

Output:

```bash
address: 00cc39cb91242b1ff0968b6e6184f349c903c37f1462b90809c5a536dc1720ef0e balance: 89.999820800 NAI
```

In addition, we also got issued an NFT for the above mentioned marketplace token which is unique to this dataset. This NFT is used to prove that the user is subscribed to the dataset in the marketplace.

```bash
./build/nuklai-cli key balance-nft
```

Output:

```bash
collectionID: Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ
uniqueID: 1
uri: 2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE
metadata: {"assetForPayment":"11111111111111111111111111111111LpoYY","datasetID":"2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE","datasetPricePerBlock":"1000000000","expirationBlock":"900","issuanceBlock":"890","marketplaceAssetID":"Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ","numBlocksToSubscribe":"10","totalCost":"10000000000"}
ownerAddress: 00cc39cb91242b1ff0968b6e6184f349c903c37f1462b90809c5a536dc1720ef0e
You own this NFT
```

This NFT contains details on how much the user paid to subscribe to the dataset and when does the access expire.

Now, let's check the info about the dataset from the marketplace again. The metadata should be updated:

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
Retrieving dataset info from the marketplace: 2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE
dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetID=Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-ds1 AssetURI=2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE TotalSupply=1 MaxSupply=0 Owner=000000000000000000000000000000000000000000000000000000000000000000
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "datasetID":"2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"890", "marketplaceAssetID":"Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ", "paymentClaimed":"0", "paymentRemaining":"10000000000", "publisher":"00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9", "subscriptions":"1"}
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
0) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999899.999295712 NAI
1) address: 00cc39cb91242b1ff0968b6e6184f349c903c37f1462b90809c5a536dc1720ef0e balance: 89.999820800 NAI
set default key: 0
```

```bash
./build/nuklai-cli marketplace claim-payment
```

Output:

```bash
Retrieving dataset info from the marketplace: 2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE
dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetID=Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-ds1 AssetURI=2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE TotalSupply=1 MaxSupply=0 Owner=000000000000000000000000000000000000000000000000000000000000000000
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "datasetID":"2TpSHwR82k4tqAhgaFMpVLryKrLdpUHdwGaibycPe13HqmYoPE", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"890", "marketplaceAssetID":"Up1AiTUQP8bMbffodMu5qK3ryXWneiPeoBboREeAtVFdUAcsQ", "paymentClaimed":"0", "paymentRemaining":"10000000000", "publisher":"00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9", "subscriptions":"1"}
✔ continue (y/n): y█
✅ txID: 7hY1HUarjFqxvaAF3a3dZPsHUQBykz6Ht6z5WtCPEfamCxmKw
fee consumed: 0.000175400 NAI
output:  &{LastClaimedBlock:1092 PaymentClaimed:10000000000 PaymentRemaining:0 DistributedReward:10000000000 DistributedTo:00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9}
```

Let's ensure that we are rewarded with the payment:

```bash
./build/nuklai-cli key balance
```

Output:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999909.999120235 NAI
```

Looks like we got 10 NAI as expected. Previously, our balance was `852999899.999295712 NAI` and now, our new balance is `852999909.999120235 NAI`
