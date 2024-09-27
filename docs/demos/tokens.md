# Digital Assets(tokens)

## Step 1: Create Your Asset

We can create our own asset on nuklaichain.

Note that there are 3 types of assets you can create:

- Fungible token (similar to erc20)
- Non-Fungible token (similar to erc721)
- Dataset token (fractionalized where there's 1 parent NFT and corresponding child NFTs)

To do so, we do:

### Create an asset of type "fungible"

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 0
name: Kiran1
symbol: KP1
decimals: 0
✔ metadata: test1█
continue (y/n): y
✅ txID: 26o6ti2ua2uBwkzGL66FPK2fUju1uoTDqy7uDUxC9mfaXmNvTG
fee consumed: 0.000064000 NAI
output:  &{AssetID:2Q7FsxpvScd1vf2UweTiE4iUrAPeQiwgjbUhq5B28Y6shTTPb3 AssetBalance:0 DatasetParentNftID:11111111111111111111111111111111LpoYY}
```

Note that the DatasetParentNftID value is `11111111111111111111111111111111LpoYY` which is an empty ID. An NFT is only generated when you create an asset of type "dataset".

### Create an asset of type "non-fungible"

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 1
name: Kiran2
✔ symbol: KP2█
✔ metadata: test2█
continue (y/n): y
✅ txID: 92yiBuztAEw7BRiQGcDnWYzkLJQDn9rCfAaNNawQKpRqPdeL1
fee consumed: 0.000082000 NAI
output:  &{AssetID:2m4LK7CNSpqZg35spCaLyYTbCdtpce7w4GMraLVXzjhNkS5WfX AssetBalance:0 DatasetParentNftID:11111111111111111111111111111111LpoYY}
```

Note that the DatasetParentNftID value is `11111111111111111111111111111111LpoYY` which is an empty ID. An NFT is only generated when you create an asset of type "dataset".

### Create an asset of type "dataset"

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 2
✔ name: Dataset0█
symbol: DS0
decimals: 0
metadata: test0
continue (y/n): y
✅ txID: JRxtYf9v5NQS7MyZM6MQjK7psghistzR89Av3E5AZoCabiZMJ
fee consumed: 0.000082200 NAI
output:  &{AssetID:np331YT7K3XdGKFriWNzwf9ieWeSMga6SZhZSaRCB9FU3QZq2 AssetBalance:1 DatasetParentNftID:res9ukBitVz9pkiy2nmZNQcYM8BYr1uTUvZCAbrjACHyPVCP2}
```

Since it's a fractionalized asset type, nuklai automatically mints the parent NFT at the same time.

## Step 2: Mint Your Asset

After we've created our own asset, we can now mint some of it.

You can do so by running the following command from this location:

### Mint a fungible token

```bash
./build/nuklai-cli asset mint-ft
```

When you are done, the output should look something like this (usually easiest
just to mint to yourself).

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
✔ assetID: 2Q7FsxpvScd1vf2UweTiE4iUrAPeQiwgjbUhq5B28Y6shTTPb3█
assetType: Fungible Token name: Kiran1 symbol: KP1 decimals: 0 metadata: test1 uri: https://nukl.ai totalSupply: 0 maxSupply: 0 admin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
recipient: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
amount: 100
✔ continue (y/n): y█
✅ txID: 2iaokSRAbwkc7Pcqia75gMZbPJfQCd7Qzh9kTH7dfYTULiRgpU
fee consumed: 0.000051600 NAI
```

### Mint a non-fungible token

```bash
./build/nuklai-cli asset mint-nft
```

When you are done, the output should look something like this (usually easiest
just to mint to yourself).

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetID: 2m4LK7CNSpqZg35spCaLyYTbCdtpce7w4GMraLVXzjhNkS5WfX
assetType: Non-Fungible Token name: Kiran2 symbol: KP2 decimals: 0 metadata: test2 uri: https://nukl.ai totalSupply: 0 maxSupply: 0 admin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
✔ recipient: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9█
unique nft #: 1
✔ metadata: https://metadata█
continue (y/n): y
✅ txID: MnL68d4bLpwQ2Z6qLwSUobceEL9Fs557Pgwn1JNaqDwRamAsk
fee consumed: 0.000073600 NAI
output:  &{NftID:ubza6ioqMZQ4ESuTtMbe6jcKTkYdaQWaEfEvYb39muKsNeS6v To:00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 OldBalance:0 NewBalance:1 AssetTotalSupply:1}
```

## Step 3: View Your Balance

Now, let's check that the mint worked right by checking our balance. You can do
so by running the following command from this location:

### Check balance of our fungible token

```bash
./build/nuklai-cli key balance-ft
```

When you are done, the output should look something like this:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetID: 2Q7FsxpvScd1vf2UweTiE4iUrAPeQiwgjbUhq5B28Y6shTTPb3
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
assetType:  Fungible Token name: Kiran1 symbol: KP1 decimals: 0 metadata: test1 uri: https://nukl.ai totalSupply: 100 maxSupply: 0 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
balance: 100 KP1
```

### Check balance of our non-fungible token

```bash
./build/nuklai-cli key balance-ft
```

When you are done, the output should look something like this:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetID: 2m4LK7CNSpqZg35spCaLyYTbCdtpce7w4GMraLVXzjhNkS5WfX
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
assetType:  Non-Fungible Token name: Kiran2 symbol: KP2 decimals: 0 metadata: test2 uri: https://nukl.ai totalSupply: 1 maxSupply: 0 owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountAdmin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
balance: 1 KP2
```

### Check the NFT info

```bash
./build/nuklai-cli key balance-nft
```

The output should be something like:

```bash
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
chainID: 2hKDi8QVgngBxCbakibVqQFa3EV8YzA957q7nPT5vrQRpx8Z9E
assetID: ubza6ioqMZQ4ESuTtMbe6jcKTkYdaQWaEfEvYb39muKsNeS6v
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
collectionID: 2m4LK7CNSpqZg35spCaLyYTbCdtpce7w4GMraLVXzjhNkS5WfX
uniqueID: 1
uri: https://metadata
metadata: https://metadata
ownerAddress: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
You own this NFT
```
