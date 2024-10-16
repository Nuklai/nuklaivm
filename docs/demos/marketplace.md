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
assetType(0 for fungible, 1 for non-fungible and 2 for fractional): 2
name: dataset1
symbol: ds1
✔ metadata: test1█
✔ continue (y/n): y█
✅ txID: XLouTNgX9mZkSzwjHfVdjW9wFQ6UejJXjTEPjeTzf43Pty7Ai
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
✅ txID: N1CFxkQA2MCGy3tQgJULYn8MK1avUcegokPRJGPMR3BX8UZ5f
fee consumed: 0.000176300 NAI
output:  &{DatasetAddress:02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7 DatasetParentNftAddress:016cf1cfff3f4c1aa8b081376c0119b1a16b82466c75d174d6ff88cf93444b50dd}
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
datasetAddress: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7
paymentAssetAddress (use NAI for native token): NAI
[0m: NAI█
assetType:  Fungible Token name: nuklaivm symbol: NAI decimals: 9 metadata: Nuklai uri: 00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb totalSupply: 852999999999727500 maxSupply: 10000000000000000000 owner: 000000000000000000000000000000000000000000000000000000000000000000 mintAdmin: 000000000000000000000000000000000000000000000000000000000000000000 pauseUnpauseAdmin: 000000000000000000000000000000000000000000000000000000000000000000 freezeUnfreezeAdmin: 000000000000000000000000000000000000000000000000000000000000000000 enableDisableKYCAccountAdmin: 000000000000000000000000000000000000000000000000000000000000000000
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999999.999727488 NAI
priceAmountPerBlock: 1
continue (y/n): y
✅ txID: 2RFFkKvpCA6GqNqi5VojfzwmLfqRWwG2ADywQ9t2jPbUQHJL5w
fee consumed: 0.000158200 NAI
output:  &{MarketplaceAssetAddress:0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee PaymentAssetAddress:00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb Publisher:00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 DatasetPricePerBlock:1000000000}
```

Upon successful transaction, it also creates a unique marketplace NFT collection with assetAddress `0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee`.
This asset has no owner and cannot be updated by anyone. Subsequent NFTs are automatically generated anytime a user subscribes to your dataset. Holding this NFT proves that the user is subscribed to your dataset and can access the dataset at anytime. This is different from the dataset NFTs that the dataset contributors own which are used for claiming the payment from all the subscriptions.

## View dataset info from the marketplace

We can also look up more info about the dataset that is in the marketplace.

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
Retrieving dataset info from the marketplace: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7

marketplace dataset info:
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetAddress=0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee PaymentAssetAddress=00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb DatasetPricePerBlock=1000000000 DatasetOwner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

marketplace asset info:
AssetType=Marketplace Token AssetName=NMAsset AssetSymbol=NMA AssetURI=0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee TotalSupply=0 MaxSupply=0 Owner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
AssetMetadata=map[string]string{"datasetAddress":"02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"0", "marketplaceAssetAddress":"0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee", "paymentAssetAddress":"00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb", "paymentClaimed":"0", "paymentRemaining":"0", "publisher":"00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9", "subscriptions":"0"}
```

This displays summary of the dataset that includes the marketplaceAssetAddress in the marketplace along with the asset used for payment and price to access dataset per block. It also shows the unique metadata for this marketplace NFT collection.

## Subscribe to the dataset in the marketplace

Let's switch to another account and subscribe to this dataset.

```bash
./build/nuklai-cli key set
```

Output:

```bash
stored keys: 2
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
0) address: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 balance: 100.000000000 NAI
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999899.999520779 NAI
set default key: 0
```

We are going to use our address #0 for this exercise.

Now, let's subscribe to this dataset for 5 blocks:

```bash
 ./build/nuklai-cli marketplace subscribe
```

Output:

```bash
marketplaceAddress: 0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee
✔ paymentAssetAddress (use NAI for native token): NAI█
✔ numBlocksToSubscribe: 10█
continue (y/n): y
✅ txID: nkw3AA2RRfHxQ8aaXtBSBxoGsYEVm1QHFg6ENMz2n1e2shxo8
fee consumed: 0.000086200 NAI
output:  &{MarketplaceAssetAddress:0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee MarketplaceAssetNumSubscriptions:1 SubscriptionNftAddress:019649803981d79513dc8e4425a4caa9a2cc755cd529a71dab52bd9ae02a00a836 PaymentAssetAddress:00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb DatasetPricePerBlock:1000000000 TotalCost:10000000000 NumBlocksToSubscribe:10 IssuanceBlock:219 ExpirationBlock:229}
```

Let's check the balance as it should automatically take 5 \* pricePerBlock from our account:

```bash
./build/nuklai-cli key balance
```

Output:

```bash
address: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 balance: 89.999913800 NAI
```

In addition, we also got issued an NFT for the above mentioned marketplace token which is unique to this dataset. This NFT is used to prove that the user is subscribed to the dataset in the marketplace.

```bash
./build/nuklai-cli key nft
```

Output:

```bash
assetAddress: 019649803981d79513dc8e4425a4caa9a2cc755cd529a71dab52bd9ae02a00a836
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
assetType:  Non-Fungible Token name: NMAsset symbol: NMA-0 metadata: {"datasetAddress":"02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7","datasetPricePerBlock":"1000000000","expirationBlock":"229","issuanceBlock":"219","marketplaceAssetAddress":"0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee","numBlocksToSubscribe":"10","paymentAssetAddress":"00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb","totalCost":"10000000000"} collectionAssetAddress: 0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee owner: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145
collectionAssetAddress: 0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee balance: 1 NMA-0
You own this NFT
```

This NFT contains details on how much the user paid to subscribe to the dataset and when does the access expire.

Now, let's check the info about the dataset from the marketplace again. The metadata should be updated:

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
Retrieving dataset info from the marketplace: 02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7

marketplace dataset info:
DatasetName=dataset1 DatasetDescription=desc1 IsCommunityDataset=true MarketplaceAssetAddress=0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee PaymentAssetAddress=00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb DatasetPricePerBlock=1000000000 DatasetOwner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9

marketplace asset info:
AssetType=Marketplace Token AssetName=NMAsset AssetSymbol=NMA AssetURI=0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee TotalSupply=1 MaxSupply=0 Owner=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
AssetMetadata=map[string]string{"datasetAddress":"02961eb5900643c5cd2b40812f12dcc6ff5db827a3d02271eaad16b96d5069cfb7", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"219", "marketplaceAssetAddress":"0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee", "paymentAssetAddress":"00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb", "paymentClaimed":"0", "paymentRemaining":"10000000000", "publisher":"00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9", "subscriptions":"1"}
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
stored keys: 2
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
0) address: 007677e11d0141fa64b15a7f834f81f2339679041c384cc87277483dbd20ef4145 balance: 89.999560700 NAI
chainID: 2ifYVzAcfsN8Bcf9g5p4beS2QkMPkNH5oGwyV3gqLwMfeXpwpZ
1) address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999899.999316692 NAI
✔ set default key: 1█
```

```bash
./build/nuklai-cli marketplace claim-payment
```

Output:

```bash
marketplaceAddress: 0206c96ae598a1ce3ce6433d595601261703c83ec2f1c481e726267a9f28cab0ee
✔ paymentAssetAddress (use NAI for native token): NAI█
continue (y/n): y
✅ txID: hCi5pnY9fst4nhgbHYDc2L352caP6w4WSNS6QP7yXudFTyHxF
fee consumed: 0.000059900 NAI
output:  &{LastClaimedBlock:827 PaymentClaimed:10000000000 PaymentRemaining:0 DistributedReward:10000000000 DistributedTo:00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9}
```

Let's ensure that we are rewarded with the payment:

```bash
./build/nuklai-cli key balance
```

Output:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 852999909.999460816 NAI
```

Looks like we got 10 NAI as expected. Previously, our balance was `852999899.999316692 NAI` and now, our new balance is `852999909.999460816 NAI`
