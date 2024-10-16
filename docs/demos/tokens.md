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
assetType(0 for fungible, 1 for non-fungible and 2 for fractional): 0
name: Kiran1
symbol: KP1
decimals: 8
metadata: test1
✔ continue (y/n): y█
✅ txID: e9rcYqsGTjkn32jMedp37BbLZtji3iPeNfiiZVNQ4JN5NK2QX
fee consumed: 0.000075000 NAI
output:  &{AssetAddress:00b2f59484610fe1e1ef8a42fa25193e73994694edcd1b5fea3800ba6fc32383a1 AssetBalance:0 DatasetParentNftAddress:}
```

Note that an NFT is only generated when you create an asset of type "dataset".

### Create an asset of type "non-fungible"

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
✔ assetType(0 for fungible, 1 for non-fungible and 2 for fractional): 1█
name: Kiran2
symbol: KP2
metadata: test2
continue (y/n): y
✅ txID: 2e5MAUsXGMM2s8tmL6wTFsLcjUujCYU4XpCbivMyim62RjBzZ
fee consumed: 0.000075000 NAI
output:  &{AssetAddress:01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706 AssetBalance:0 DatasetParentNftAddress:}
```

Note that an NFT is only generated when you create an asset of type "dataset".

### Create an asset of type "dataset"

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
assetType(0 for fungible, 1 for non-fungible and 2 for fractional): 2
name: Kiran3
✔ metadata: test3█
continue (y/n): y
✅ txID: wn5FYH92RS6CQ9do4hww5Mw26gWAHrWGGFc83v8crhFpjt47c
fee consumed: 0.000096000 NAI
output:  &{AssetAddress:0211efdf03c7e941d9a74fa61c904156580d7e6c8689b56575de8dfaa8c26771dc AssetBalance:1 DatasetParentNftAddress:019a40ee0e2ca643fa9256c47247da7d0dff13f280a26cc8ef2c1fa5f3e8cffbe7}
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
assetAddress: 00b2f59484610fe1e1ef8a42fa25193e73994694edcd1b5fea3800ba6fc32383a1
assetType: Fungible Token name: Kiran1 symbol: KP1 decimals: 8 metadata: test1 uri: 00b2f59484610fe1e1ef8a42fa25193e73994694edcd1b5fea3800ba6fc32383a1 totalSupply: 0 maxSupply: 0 admin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
recipient: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
amount: 10
amount:  1000000000 8
continue (y/n): y
✅ txID: 2PcJhRGxdqkiDwUdb3Fhyba3vrHymgJ6UETLTTrBWSo7RVkToK
fee consumed: 0.000065200 NAI
output:  &{OldBalance:0 NewBalance:1000000000}
```

### Mint a non-fungible token

```bash
./build/nuklai-cli asset mint-nft
```

When you are done, the output should look something like this (usually easiest
just to mint to yourself).

```bash
assetAddress: 01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706
assetType: Non-Fungible Token name: Kiran2 symbol: KP2 decimals: 0 metadata: test2 uri: 01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706 totalSupply: 0 maxSupply: 0 admin: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 mintActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 pauseUnpauseActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 freezeUnfreezeActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 enableDisableKYCAccountActor: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
recipient: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
metadata: testtest
continue (y/n): y
✅ txID: 2mQX5YF4QdTjQiup5Z8CpGkJfFPqouWUCHQc9KwzyVqcuDLA5Q
fee consumed: 0.000086400 NAI
output:  &{AssetNftAddress:0188f06e048bb40c31d587d1dda5615b14c8cb55577747b142b2c77d69be5ba292 OldBalance:0 NewBalance:1}
```

## Step 3: View Your Balance

Now, let's check that the mint worked right by checking our balance. You can do
so by running the following command from this location:

### Check balance of our asset

```bash
./build/nuklai-cli key balance-asset
```

When you are done, the output should look something like this:

```bash
assetAddress: 00b2f59484610fe1e1ef8a42fa25193e73994694edcd1b5fea3800ba6fc32383a1
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
address: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9 balance: 10.00000000 KP1
```

### Check the NFT info

```bash
./build/nuklai-cli key nft
```

The output should be something like:

```bash
assetAddress: 019a40ee0e2ca643fa9256c47247da7d0dff13f280a26cc8ef2c1fa5f3e8cffbe7
uri: http://127.0.0.1:9650/ext/bc/nuklaivm
assetType:  Non-Fungible Token name: Kiran3 symbol: KP3-0 metadata: test3 collectionAssetAddress: 0211efdf03c7e941d9a74fa61c904156580d7e6c8689b56575de8dfaa8c26771dc owner: 00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9
collectionAssetAddress: 0211efdf03c7e941d9a74fa61c904156580d7e6c8689b56575de8dfaa8c26771dc balance: 1 KP3-0
You own this NFT
```
