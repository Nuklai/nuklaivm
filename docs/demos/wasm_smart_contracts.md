# WASM Smart Contracts

TODO: This is a work in progress.

## Pre-requisites

First install the following dependencies:

### Install Rust

You can install Rust by following these instructions:

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

For other platforms, follow directions at [https://www.rust-lang.org/tools/install](https://www.rust-lang.org/tools/install) and install Rust on your machine.

### Install Rust and WASM toolchain

You can do this by running the following command:

```bash
rustup target add wasm32-unknown-unknown
cargo install wasm-pack
```

## Examples

You can find the source code for the wasm smart contracts examples under [contracts](../../x/contracts/examples) directory.

### Counter

This is a smart contract that stores an Address -> Count mapping. It contains two functions `inc` which increments a value by a count and `get_value` a value at an Address.

#### Build the contract

Navigate to the project root directory and build the WASM binary:

```bash
cd x/contracts/examples/counter;
cargo build --release --target wasm32-unknown-unknown --target-dir=./target
```

The generated WASM binary will be found at [counter.wasm](./target/wasm32-unknown-unknown/release/counter.wasm)

#### Publish the contract to the blockchain

Provide the path to the WASM file when prompted.

```bash
./build/nuklai-cli action publishFile
```

Which should output something like:

```bash
contract file: ./x/contracts/examples/counter/target/wasm32-unknown-unknown/release/counter.wasm
continue (y/n): y
✅ txID: BfMeE6jBtvUXfQLLv9u61auFUTr9CtHY4c4wXnYXbWykjZo3J
fee consumed: 0.007489900
0300000023052025B7F3BEEDE2CB8514C1DE7B073CB1E13FE2DF06DF4CB0A1C5FE44A781EE6903EF
```

Note down the contract ID `0300000023052025B7F3BEEDE2CB8514C1DE7B073CB1E13FE2DF06DF4CB0A1C5FE44A781EE6903EF` as it will be required for deployment

#### Deploy the contract on the blockchain

```bash
./build/nuklai-cli action deploy
```

Which should output something like:

```bash
✔ contract id: 0300000023052025B7F3BEEDE2CB8514C1DE7B073CB1E13FE2DF06DF4CB0A1C5FE44A781EE6903EF█
✔ creation info: 00█
✔ continue (y/n): y█
✅ txID: 2KkfSyDAm9UJ5eZKekJ9BAtun2NbuwgstUV2ffPde46GyA1z4a
fee consumed: 0.000045800 NAI
output:  &{Address:006e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d}
```

#### Call a function from the deployed contract

```bash
./build/nuklai-cli action
```
