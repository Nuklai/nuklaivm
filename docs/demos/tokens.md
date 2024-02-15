### Mint an asset

#### Step 1: Create Your Asset

We can create our own asset on nuklaichain. To do so, we do:

```bash
./build/nuklai-cli action create-asset
```

When you are done, the output should look something like this:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
symbol: TOKEN1
✔ decimals: 9
metadata: Example token1
✔ continue (y/n): y
✅ txID: ggMxHuutoobCLfuYyiLRYr1VMq7r9ULcBU621kvurtsqdifjN
```

_`txID` is the `assetID` of your new asset._

The "loaded address" here is the address of the default private key (`demo.pk`). We
use this key to authenticate all interactions with the `nuklaivm`.

#### Step 2: Mint Your Asset

After we've created our own asset, we can now mint some of it. You can do so by
running the following command from this location:

```bash
./build/nuklai-cli action mint-asset
```

When you are done, the output should look something like this (usually easiest
just to mint to yourself).

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
assetID: ggMxHuutoobCLfuYyiLRYr1VMq7r9ULcBU621kvurtsqdifjN
symbol: TOKEN1 decimals: 9 metadata: Example token1 supply: 0
recipient: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✔ amount: 1000█
continue (y/n): y
✅ txID: 2eG6m32NMgaxdz36W8omF1tQNBAkXr2JHojsU67X1AQzT73PSf
```

#### Step 3: View Your Balance

Now, let's check that the mint worked right by checking our balance. You can do
so by running the following command from this location:

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
assetID (use NAI for native token): ggMxHuutoobCLfuYyiLRYr1VMq7r9ULcBU621kvurtsqdifjN
uri: http://127.0.0.1:43689/ext/bc/277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
symbol: TOKEN1 decimals: 9 metadata: Example token1 supply: 1000000000000 warp: false
balance: 1000.000000000 TOKEN1
```

### Transfer Assets to Another Subnet

Unlike the mint demo, the AWM demo only requires running a single
command. You can kick off a transfer between the 2 Subnets you created by
running the following command from this location:

```bash
./build/nuklai-cli action export
```

When you are done, the output should look something like this:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX
✔ assetID (use NAI for native token): ggMxHuutoobCLfuYyiLRYr1VMq7r9ULcBU621kvurtsqdifjN█
symbol: TOKEN1 decimals: 9 metadata: Example token1 supply: 1000000000000 warp: false
balance: 1000.000000000 TOKEN1
recipient: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
✔ amount: 50█
✔ reward: 0█
available chains: 1 excluded: [277DehNDB9szuxsiMgAfQBaJhW7JuE9CP6bdW5Up9D3qNeipZX]
0) chainID: QXRg1vtkJe4rPF2ShjHBj3ocyrRhW8A24aSM3qCsK35DKV5iD
destination: 0 [auto-selected]
continue (y/n): y
✅ txID: 2LwezG85VjTktvfZpwnGigvqH3o74ZC9qBWmWUwjsXmqAwQMsV
perform import on destination (y/n): y
2RiJax29P6XVQzeXXApigqBBcsmrZjJ1vXKPBv1RaTKc5mqaxW to: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx source assetID: ggMxHuutoobCLfuYyiLRYr1VMq7r9ULcBU621kvurtsqdifjN source symbol: TOKEN1 output assetID: 7W44tf7zcvWW9MYLLvQdoL6Uf7L5MZ2vgNdGqYJ5caeEbMiuB value: 50.000000000 reward: 0.000000000 return: false
✅ txID: FLkYqXKMmE9HwKksdmiwyJberPzq4zEeEupwHDgDcAXkhjmGX
✔ switch default chain to destination (y/n): y█
```

_The `export` command will automatically run the `import` command on the
destination. If you wish to import the AWM message using a separate account,
you can run the `import` command after changing your key._
