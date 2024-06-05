### Mint an asset

#### Step 1: Create Your Asset

We can create our own asset on nuklaichain. To do so, we do:

```bash
./build/nuklai-cli action create-asset
```

When you are done, the output should look something like this:

```
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2JJEMRDZ3Bw2qf62PHxVGP2NV2QVDMJvz33CY62vgrr3kuiGXb
✔ symbol: TOKEN1█
✔ decimals: 9█
✔ metadata: Example token1█
continue (y/n): y
✅ txID: 5Qd4vuWoxBa3qwq8Ktm19Q4PLGqKS5NjW3RDd2pSGbCYsSied
assetID: 2NgdoxJSTvUm6fSJP9C3tAB4k8XrLyK2bSULM1zbNQhMB7w19
```

#### Step 2: Mint Your Asset

After we've created our own asset, we can now mint some of it. You can do so by
running the following command from this location:

```bash
./build/nuklai-cli action mint-asset
```

When you are done, the output should look something like this (usually easiest
just to mint to yourself).

```
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2JJEMRDZ3Bw2qf62PHxVGP2NV2QVDMJvz33CY62vgrr3kuiGXb
assetID: 2NgdoxJSTvUm6fSJP9C3tAB4k8XrLyK2bSULM1zbNQhMB7w19
symbol: TOKEN1 decimals: 9 metadata: Example token1 supply: 0
recipient: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
amount: 10000
✔ continue (y/n): y█
✅ txID: 2bow3eEg8xW9Kn1q4X9xsy3uc4gvSG3waAvTc99ykWfVsJZaj7
```

#### Step 3: View Your Balance

Now, let's check that the mint worked right by checking our balance. You can do
so by running the following command from this location:

```bash
./build/nuklai-cli key balance
```

When you are done, the output should look something like this:

```
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: 2JJEMRDZ3Bw2qf62PHxVGP2NV2QVDMJvz33CY62vgrr3kuiGXb
assetID (use NAI for native token): 2NgdoxJSTvUm6fSJP9C3tAB4k8XrLyK2bSULM1zbNQhMB7w19
uri: http://127.0.0.1:9654/ext/bc/2JJEMRDZ3Bw2qf62PHxVGP2NV2QVDMJvz33CY62vgrr3kuiGXb
symbol: TOKEN1 decimals: 9 metadata: Example token1 supply: 10000000000000
balance: 10000.000000000 TOKEN1
```
