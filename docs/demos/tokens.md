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
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
✔ assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 0█
name: Kiran
symbol: KP1
decimals: 0
✔ metadata: test█
continue (y/n): y
✅ txID: tmQc4mxt5FC1Ts9hW6rvdDDxQAUw1sTJscVQF4mTfWk59SJXN
assetID: djTHcsCKqbbdNNt1JwafXfju8UKJ7oCkHY4nwWcbDF31S3MZv
```

### Create an asset of type "non-fungible"

```bash
./build/nuklai-cli asset create
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
✔ assetType(0 for fungible, 1 for non-fungible and 2 for dataset): 1█
name: Kiran
symbol: KP2
decimals: 0
✔ metadata: test█
continue (y/n): y
✅ txID: 24vUfjLCA674Qo2Paq24buu8nBY3YChFKHn8SfirgqWkRhcmx4
assetID: BK8Ghu86S6YkB4dHcQb7guJdbj5KZCXJdjPdF8XVxP7v4KHKW
```

### Create an asset of type "dataset"

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
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetID: djTHcsCKqbbdNNt1JwafXfju8UKJ7oCkHY4nwWcbDF31S3MZv
assetType: Fungible Token name: Kiran symbol: KP1 decimals: 0 metadata: test uri: https://nukl.ai totalSupply: 0 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
recipient: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
amount: 1010█
✔ amount: 10█
continue (y/n): y
✅ txID: 2iJfeGv5K6W4diY9nWcuMsARdePEs9rGHq1RUBAesqusiuhfqo
```

### Mint a non-fungible token

```bash
./build/nuklai-cli asset mint-nft
```

When you are done, the output should look something like this (usually easiest
just to mint to yourself).

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetID: BK8Ghu86S6YkB4dHcQb7guJdbj5KZCXJdjPdF8XVxP7v4KHKW
assetType: Non-Fungible Token name: Kiran symbol: KP2 decimals: 0 metadata: test uri: https://nukl.ai totalSupply: 0 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
recipient: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
unique nft #: 0
✔ uri: test█
continue (y/n): y
✅ txID: kJMvtujNWjmbzxdVbY5VL9Tnyj2PPYcvTawEFuydJLCSmXzoo
NFT ID: XbA9tWs3o1VUm3eMeKwZiWvsoAPdHXzyY4awZeAwvJ64e8VjH
```

## Step 3: View Your Balance

Now, let's check that the mint worked right by checking our balance. You can do
so by running the following command from this location:

### Check balance of our fungible token

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetID (use NAI for native token): djTHcsCKqbbdNNt1JwafXfju8UKJ7oCkHY4nwWcbDF31S3MZv
uri: http://127.0.0.1:9658/ext/bc/24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetType: {{/} Fungible Token name: Kiran symbol: KP1 decimals: 0 metadata: test uri: https://nukl.ai totalSupply: 10 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 10 KP1
```

### Check balance of our non-fungible token

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```bash
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetID (use NAI for native token): BK8Ghu86S6YkB4dHcQb7guJdbj5KZCXJdjPdF8XVxP7v4KHKW
uri: http://127.0.0.1:9658/ext/bc/24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
assetType: {{/} Non-Fungible Token name: Kiran symbol: KP2 decimals: 0 metadata: test uri: https://nukl.ai totalSupply: 1 maxSupply: 0 admin: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx mintActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx pauseUnpauseActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx freezeUnfreezeActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx enableDisableKYCAccountActor: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
balance: 1 KP2
```

### Check the NFT info

```bash
./build/nuklai-cli key balanceNFT
```

The output should be something like:

```bash
✔ address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx█
chainID: 24YK3t6WsiXeQTtTPEtPfSnZVoodZdN4wRRr1oMGdQHTKdF7bK
nftID: XbA9tWs3o1VUm3eMeKwZiWvsoAPdHXzyY4awZeAwvJ64e8VjH
collectionID: BK8Ghu86S6YkB4dHcQb7guJdbj5KZCXJdjPdF8XVxP7v4KHKW
uniqueID: 0
uri: test
ownerAddress: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
You own this NFT
```
