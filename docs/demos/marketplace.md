# Nuklai Marketplace

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

## Publish the dataset to Nuklai marketplace

We can put up our dataset up for sale on the native nuklai marketplace at anytime. Note that once it's on the marketplace, no new data can be added to the dataset. This is to accurately calculate the reward being generated from the subscription so it can be paid out to all the current NFT owners of the dataset.

```bash
./build/nuklai-cli marketplace publish
```

You can set the asset that subscribers must use for payment and you can also dictate the price of accessing your dataset per block as shown below.

Output:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
assetForPayment (use NAI for native token): NAI
balance: 852999939.999561310 NAI
priceAmountPerBlock: 1
continue (y/n): y
✅ txID: 4Z61pHJnbFixtX11u4pb7DMAuDbFj2GSuNogCN5ntQZjUsi2k
assetID: 2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW
```

Upon successful transaction, it also creates a unique marketplace NFT collection with assetID `2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW`.
This asset has no owner and cannot be updated by anyone. Subsequent NFTs are automatically generated anytime a user subscribes to your dataset. Holding this NFT proves that
the user is subscribed to your dataset and can access the dataset at anytime. This is different from the dataset NFTs that the dataset contributors own which are used for claiming the payment from all the subscriptions.

## View dataset info from the marketplace

We can also look up more info about the dataset that is in the marketplace.

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
Retrieving dataset info from the marketplace: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc for dataset1 IsCommunityDataset=true SaleID=2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-datas AssetURI=2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH TotalSupply=0 MaxSupply=0 Owner=nuklai1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqvn638
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "dataset":"2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"0", "marketplaceID":"2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW", "paymentClaimed":"0", "paymentRemaining":"0", "publisher":"nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx", "subscriptions":"0"}
```

This displays summary of the dataset that includes the saleID in the marketplace along with the asset used for payment and price to access dataset per block. It also shows the unique metadata for this marketplace NFT collection with asset ID `2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW`

## Subscribe to the dataset in the marketplace

Let's switch to another account and subscribe to this dataset.

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999939.999388337 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 9.999867900 NAI
set default key: 1
```

Now, let's subscribe to this dataset for 5 blocks:

```bash
 ./build/nuklai-cli marketplace subscribe
```

Output:

```bash
address: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
Retrieving dataset info for datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
dataset info:
Name=dataset1 Description=desc for dataset1 Categories=dataset1 LicenseName=MIT LicenseSymbol=MIT LicenseURL=https://opensource.org/licenses/MIT Metadata=metadata for dataset1 IsCommunityDataset=true SaleID=2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW BaseAsset=11111111111111111111111111111111LpoYY BasePrice=1000000000 RevenueModelDataShare=100 RevenueModelMetadataShare=0 RevenueModelDataOwnerCut=10 RevenueModelMetadataOwnerCut=0 Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✔ assetForPayment (use NAI for native token): NAI█
numBlocksToSubscribe: 5
balance: 9.999867900 NAI
✔ continue (y/n): y█
✅ txID: 2swrbYWRSGyS1o3sXpBup8vDBNryRJ51dnduFX2xY9kPziT3dw
nftID: 2dmzqK7XgrhVAWqKBjae142docPFxsm5tgk4YCFNC4QvY4VLHh
```

Let's check the balance as it should automatically take 5 \* pricePerBlock from our account:

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999939.999388337 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 4.999688700 NAI
set default key: 1
```

In addition, we also got issued an NFT for the above mentioned marketplace token which is unique to this dataset. This NFT is used to prove that the user is subscribed to the dataset in the marketplace.

```bash
./build/nuklai-cli key balanceNFT
```

Output:

```bash
✔ address: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp█
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
nftID: 2dmzqK7XgrhVAWqKBjae142docPFxsm5tgk4YCFNC4QvY4VLHh
collectionID: 2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW
uniqueID: 1
uri: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
metadata: {"dataset":"2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH","marketplaceID":"2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW","datasetPricePerBlock":"1000000000","totalCost":"5000000000","assetForPayment":"11111111111111111111111111111111LpoYY","issuanceBlock":"746","expirationBlock":"751","numBlocksToSubscribe":"5"}
ownerAddress: nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp
You own this NFT
```

This NFT contains details on how much the user paid to subscribe to the dataset and when does the access expire.

Now, let's check the info about the dataset from the marketplace again. The metadata should be updated:

```bash
./build/nuklai-cli marketplace info
```

Output:

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
Retrieving dataset info from the marketplace: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc for dataset1 IsCommunityDataset=true SaleID=2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-datas AssetURI=2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH TotalSupply=1 MaxSupply=0 Owner=nuklai1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqvn638
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "dataset":"2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"746", "marketplaceID":"2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW", "paymentClaimed":"0", "paymentRemaining":"5000000000", "publisher":"nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx", "subscriptions":"1"}
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
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999939.999388337 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 4.999688700 NAI
```

```bash
./build/nuklai-cli marketplace claim-payment
```

Output:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
✔ datasetID: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH█
Retrieving dataset info from the marketplace: 2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH
dataset info from marketplace:
DatasetName=dataset1 DatasetDescription=desc for dataset1 IsCommunityDataset=true SaleID=2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW AssetForPayment=11111111111111111111111111111111LpoYY PricePerBlock=1000000000 DatasetOwner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
marketplace asset info:
AssetType=Marketplace Token AssetName=Dataset-Marketplace-dataset1 AssetSymbol=DM-datas AssetURI=2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH TotalSupply=1 MaxSupply=0 Owner=nuklai1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqvn638
AssetMetadata=map[string]string{"assetForPayment":"11111111111111111111111111111111LpoYY", "dataset":"2vfdfyQwZ8X5xCUMsxJi8Ve9SE4AL2eDRmXmDUfJDYGKmaZqRH", "datasetPricePerBlock":"1000000000", "lastClaimedBlock":"746", "marketplaceID":"2ZtQKoP45x2PYWfu4Fmu9nayUzcG3fcVbskMgAKoPdWqWHtHLW", "paymentClaimed":"0", "paymentRemaining":"5000000000", "publisher":"nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx", "subscriptions":"1"}
✔ continue (y/n): y█
✅ txID: RbvEZfWWkUYb2wcQ4AdVV86EFggwKCa6QvQ2isCMu13Ky52YE
```

Let's ensure that we are rewarded with the payment:

```bash
./build/nuklai-cli key set
```

Output:

```bash
chainID: WPfnFTkhMN7iR2B5AfSjSa1ntFBT3RkkjwougcrH6Ck1kjLMU
stored keys: 2
0) address (ed25519): nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx balance: 852999944.999212980 NAI
1) address (ed25519): nuklai1qrnqpl6p2nx8gtuw9e9esgr7suufp7qdjk9h96rc3k9yqwla4tfc5lrcqmp balance: 4.999688700 NAI
```

Looks like we got 5 NAI as expected.
