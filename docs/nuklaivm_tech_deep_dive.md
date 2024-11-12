# NuklaiVM: A Comprehensive Technical Deep-Dive

NuklaiVM is a high-performance blockchain Virtual Machine (VM) built using HyperSDK on the Avalanche network. It provides extensive capabilities for creating, managing, and utilizing datasets on-chain, and supports a variety of asset interactions. This document is a complete reference and technical deep dive into NuklaiVM, describing its architecture, features, actions, and interactions with the blockchain, as well as its efficient state management and execution methodologies.

## Overview

NuklaiVM, constructed using [HyperSDK](https://github.com/ava-labs/hypersdk), is designed to offer an efficient blockchain environment that supports complex operations such as asset creation, dataset creation, contributions and marketplace interactions. Unlike traditional blockchain systems that rely on external smart contracts for functionality, NuklaiVM's key operations are implemented as native actions directly within the VM. This approach provides an optimized runtime with minimized latency and improved throughput.

The core functionality of NuklaiVM revolves around creating and managing datasets, enabling community contributions, issuing tokenized representations of ownership, and supporting marketplace-style interactions for dataset assets. Additionally, the VM supports various types of tokens, including fungible tokens (FTs), non-fungible tokens (NFTs), fractionalized-fungible tokens, and fractionalized-non-fungible tokens.

## Key Features

### Efficient State Management

NuklaiVM relies on HyperSDK's integration with `x/merkledb` for managing blockchain state, utilizing a path-based merkelized radix tree. This approach minimizes the on-disk footprint by removing unnecessary data. The state is maintained efficiently using:

- **Dynamic State Sync**: Allows nodes to sync the most recent state from the network, rather than processing all historical transactions.
- **Block Pruning**: Stores only a limited number of accepted blocks to keep disk requirements low. Block storage is configured based on an AcceptedBlockWindow, allowing room for disaster recovery and ensuring nodes can help new participants sync state.

### Parallel Transaction Execution

NuklaiVM can execute transactions in parallel by leveraging the HyperSDK executor package. The actions specify keys they will access, allowing non-conflicting transactions to be executed concurrently. This parallel execution capability significantly improves throughput, allowing thousands of transactions per second.

### WASM-Based Smart Contracts

NuklaiVM supports WASM-based smart contracts, providing flexibility for custom operations beyond the predefined actions. WASM binaries allow contracts to be written in languages like Rust, C, or C++ and be compiled for execution within the VM.

### Account Abstraction and Multidimensional Fee Pricing

NuklaiVM implements account abstraction, allowing flexible authorization schemes (referred to as Auth). This abstraction enables separation between transaction actors and sponsors, improving usability and extensibility.

The multidimensional fee pricing model measures resource usage independently for compute, bandwidth, storage (read, write, allocate), and other factors. This enables accurate fee calculation and prevents overcharging, while still ensuring fair resource distribution.

### Proposer-Aware Gossip and Continuous Block Production

- **Proposer-Aware Gossip**: Transactions are gossiped primarily to the next block proposers, reducing unnecessary network load and improving efficiency.
- **Continuous Block Production**: Blocks are produced continuously, even if empty, with a minimum gap to prevent issues during recovery scenarios. This enables efficient tracking and synchronization within the blockchain.

## Features Overview

The features of NuklaiVM can be divided into the following categories: Assets (Tokens), Emission Balancer, Dataset, and Marketplace. Below, each feature is described in detail to provide a comprehensive understanding of how NuklaiVM functions.

### Assets (Tokens)

NuklaiVM supports a wide variety of token types, including:

- **Fungible Tokens (FTs)**: Represent standard assets that can be split into smaller units.
- **Non-Fungible Tokens (NFTs)**: Represent unique items, often used for dataset ownership or row contributions.
- **Fractionalized Tokens**: Supports fractional ownership of non-fungible items. These are the native assets utilized by datasets.

### Asset Creation and Minting

Assets can be created using the `create_asset` action, specifying the supply, owner, and metadata. Minting of tokens, especially NFTs, can be integrated into dataset actions, such as CompleteContributeDataset. This ensures tokens are minted upon significant events like dataset contributions.

The state for each asset is represented by a unique key that contains metadata and permissions. The mint_asset action allows additional units to be minted as per the supply and permissions defined during creation.

### Token Transfers

Transfers are managed via the **Transfer** action. The logic prevents users from transferring tokens to themselves and ensures that appropriate checks are made based on whether the token is fungible or non-fungible. This action involves reading, updating, and allocating state for both the sender and recipient.

- **Balance Management**: Fungible tokens update the sender's and receiver's balances, while NFTs update ownership records.
- **Atomicity**: Transfers are processed atomically to ensure that either both the sender's balance is decremented and the recipient's balance is incremented, or neither change occurs.

#### 1. Create Asset

The **create_asset** action allows users to create new assets on the blockchain, specifying details such as name, total supply, and type (fungible or non-fungible).

- **Inputs**: Name, supply, owner address, asset type (FT, NFT, fractionalized).
- **Process**: Initializes the asset state, sets ownership details, and assigns an asset ID.
- **Output**: Asset address and Transaction ID.

#### 2. Mint Asset

The **mint_asset** action allows the minting of additional units of an asset, applicable to both fungible and non-fungible tokens.

- **Inputs**: Asset address, amount, recipient address.
- **Process**: Verifies ownership and minting permissions, then increments the asset supply and updates the recipient's balance or ownership.
- **Output**: Updated asset supply and Transaction ID.

#### 3. Burn Asset

The **burn_asset** action allows holders to destroy a specified number of tokens, reducing the total supply.

- **Inputs**: Asset address, amount, holder address.
- **Process**: Verifies the holder's balance, then decrements the asset supply and the holder's balance.
- **Output**: Updated asset supply and Transaction ID.

#### 4. Transfer Asset

The **Transfer** action allows the transfer of assets between accounts, including both fungible and non-fungible tokens.

- **Inputs**: Recipient address, asset address, value, memo.
- **Validation**: Prevents self-transfers, checks sufficient balance, ensures correct NFT value.
- **Output**: Updated balances of sender and receiver.

#### Under the Hood: Asset Actions

- **Asset Verification**: Each action verifies whether the specified asset exists and determines its type (fungible or non-fungible). This determines the validation logic for the action.
- **State Representation**: Assets are represented in the state by unique keys that include metadata such as the asset type, supply, owner, and permissions.
- **Balance Management**: For fungible tokens, the sender's and receiver's balances are updated accordingly. For NFTs, ownership metadata is updated to reflect the new owner.

### Emission Balancer

The **Emission Balancer** and **Staking Mechanism** are critical components within the NuklaiVM that help maintain network security, incentivize participation, and ensure fair reward distribution.

- **Emission Balancer**: The emission balancer manages the supply of the native token NAI and distributes emissions to stakers and validators. It tracks active and inactive validators, delegators, and the associated rewards over epochs. The emissions are calculated based on the Annual Percentage Rate (APR) and distributed proportionally to the stake.
- **Staking Rewards**: Validators and delegators receive rewards through actions like **ClaimDelegationStakeRewards** and **ClaimValidatorStakeRewards**. These actions validate inputs, ensure authorized actors claim rewards, and manage the state to distribute rewards efficiently.
- **Epoch-Based System**: The reward distribution is handled in epochs, allowing consistent emissions and ensuring that participants are rewarded for their contributions over time.

#### Validator and Delegator Reward Distribution

The **Validator and Delegator Reward Distribution** mechanism ensures that rewards are split fairly between validators and delegators based on their stake and the defined delegation fee rate.

- **Validator Rewards**: Validators earn rewards for proposing and validating blocks. These rewards are distributed using the **distributeValidatorRewardsOrFees** function, which calculates the rewards based on the staked amount and delegation fee rate.
- **Delegator Rewards**: Delegators receive a portion of the rewards generated by their delegated stake, incentivizing them to support validators and contribute to network security.

#### Under the Hood: Reward Distribution

- **Reward Calculation**: The rewards per epoch are calculated by considering the total staked amount, APR, and delegation fee rate. The **Emission** struct orchestrates this reward distribution, ensuring fairness and consistency.
- **State Updates and Consistency**: The state is updated atomically during reward distribution to prevent inconsistencies. Validators and delegators can claim their rewards through dedicated actions, which also handle event notifications for successful claims.

### Dataset

The dataset functionalities of NuklaiVM provide tools for users to create datasets, publish them for community use, contribute to datasets, and subscribe to them.

#### Create, Update, and Contribute to Datasets

The create_dataset action allows for creating datasets on-chain, issuing NFTs representing ownership. This process includes setting up initial metadata and revenue-sharing parameters.

Users can contribute to datasets by adding new rows of data through a two-step process: `initiate_contribute_dataset`t and `complete_contribute_dataset`. The latter action issues an NFT to the contributor, representing their contribution.

Contributions are stored in-memory until approved, after which they are persisted on-chain, ensuring both performance and integrity of the data contribution process.

#### 1. Create Dataset

The **create_dataset** action creates a new dataset on-chain, issuing an NFT representing ownership.

- **Inputs**: Name, description, categories, license, metadata, community dataset flag.
- **Process**: Validates inputs, generates a dataset address, issues an NFT to the owner, and sets initial revenue-sharing parameters.
- **Revenue Model**: Includes configurable shares for data contributions.
- **Output**: Dataset address and Dataset parent NFT address.

- **Under the Hood: Dataset Creation**

  - **Validation**: All input fields are validated, including verifying the uniqueness of the dataset name and ensuring the categories and license information are consistent with predefined standards.
  - **State Initialization**: A new dataset address is generated, and the corresponding state keys are initialized. This includes creating the dataset info storage entry and setting the ownership details.
  - **NFT Issuance**: An NFT representing ownership is issued to the dataset creator. This NFT is stored in the AssetInfo state, with metadata linking it to the dataset properties.

#### 2. Update Dataset

The **update_dataset** action allows modification of an existing dataset's details but does not allow changes to metadata, pricing, or sale status. These can be modified using dedicated actions.

- **Inputs**: Dataset address, updated fields (name, description, etc.).
- **Process**: Updates dataset details and validates ownership.
- **Output**: Updated fields (name, description, etc).

- **Under the Hood: Dataset Update**
  - **State Access**: The update action retrieves the current dataset state using the dataset address key. The state is locked for writing to prevent concurrent modifications.
  - **Modification**: Only the specified fields are updated, and the modified dataset state is written back to the blockchain.
  - **Ownership Check**: The action includes an ownership validation step to ensure that only the dataset owner can perform updates.

#### 3. Data Contribution

The data contribution workflow allows users to contribute new rows to an existing dataset. This is a two-step process:

- **Initiate Contribution (initiate_contribute_dataset)**:

  - **Inputs**: Dataset address, data location, data identifier.
  - **Process**: Stores contribution information in-memory, generates a dataset contribution ID, deducts a certain amount of collateral asset(eg. NAI) to prevent abuse. This collateral amount is refunded during the complete_contribute_dataset action.
  - **Output**: Dataset contribution ID, Collateral Asset address and Collateral Amount.

- **Complete Contribution (complete_contribute_dataset)**:

  - **Inputs**: Dataset contribution ID, dataset address and dataset contributor.
  - **Process**: Stores data properties on-chain, issues an NFT to the contributor, representing their contribution. The collateral is refunded back to the contributor.
  - **Output**: Dataset child NFT address.

- **Under the Hood: Data Contribution**
  - **In-Memory State Handling**: The initial contribution is stored in-memory within the NuklaiVM to allow the dataset owner to approve or reject it. This temporary state is managed to ensure it does not persist beyond a certain block limit if the contribution is not approved.
  - **NFT Issuance for Contributions**: Upon approval, an NFT is issued to the contributor. This NFT is linked to the contributed data, ensuring that provenance and contribution details are permanently recorded on-chain.

### Marketplace

The marketplace within NuklaiVM allows datasets to be published for subscription, allowing owners to generate revenue. The economics of the marketplace are governed by the following:

- **Dataset Ownership and Contributions**: Owners retain a percentage of revenues, while contributors receive the rest. The exact percentages are configurable within each dataset's revenue model.
- **Subscription Fees**: Users pay to subscribe to datasets for a specified duration. The fee is calculated based on the number of blocks and is paid using the specified base asset.

#### 1. Publish Dataset to Marketplace

The **publish_dataset_marketplace** action makes the dataset available for subscription by other users. It includes parameters for specifying a payment asset address and price per block for access.

- **Inputs**: Dataset address, payment asset address and dataset price per block.
- **Process**: Updates the dataset to "on sale" status and creates a marketplace asset.
- **Output**: Marketplace asset address and the publisher of the dataset.

- **Under the Hood: Publishing Dataset**
  - **Marketplace Asset Creation**: When a dataset is published, a marketplace-specific asset is created to represent its availability. This asset holds metadata about the pricing, payment asset, and subscription terms.
  - **State Update**: The dataset state is updated to mark it as "on sale" and link it with the marketplace asset. This allows future actions like subscriptions to interact directly with the marketplace representation of the dataset.

#### 2. Subscribe to Dataset on the Marketplace

The **subscribe_dataset_marketplace** action allows users to subscribe to datasets for a specified number of blocks, receiving an NFT representing their subscription.

- **Inputs**: Marketplace asset address, payment asset and number of blocks to subscribe.
- **Process**: Issues a subscription NFT with an expiration, charges payment, and tracks usage.
- **Output**: Subscription asset address.

- **Under the Hood: Subscription Handling**
  - **Subscription NFT**: A non-fungible token is issued to the subscriber, representing their right to access the dataset for the specified duration. The NFT contains metadata detailing the subscription period and expiration block.
  - **Payment Processing**: The subscription fee is deducted from the subscriber's balance, and the dataset owner's revenue is updated accordingly. This fee processing is handled atomically to ensure consistency.

#### 3. Claim Marketplace Payment

The **claim_marketplace_payment** action allows dataset owners to claim accumulated payments for subscriptions.

- **Inputs**: Marketplace asset address, payment asset address.
- **Process**: Calculates rewards, updates payment details, and distributes rewards.
- **Output**: Payment claim details.

- **Under the Hood: Marketplace Asset Management**
  - **Marketplace Asset Creation**: When a dataset is published, a marketplace-specific asset is created to represent its availability. This asset holds metadata about the pricing, payment asset, and subscription terms.
  - **Revenue Tracking**: The marketplace asset tracks accumulated payments, and its metadata is updated each time a new subscription is initiated.

## Technical Architecture

### Core VM Logic

The core logic of NuklaiVM is implemented in the vm/ directory, which defines the core components and execution mechanisms of the virtual machine. Below is an overview of the core VM structure:

- **Initialization and Configuration**: The VM initializes by loading configuration parameters, including state storage paths, block settings, and connection details for the external subscriber and indexer. The configuration is organized to be modular, allowing easy customization.
- **Action Registry**: NuklaiVM maintains a registry of actions, allowing the dynamic addition of new actions. Each action is associated with an action ID and includes the definition of the action's inputs, processing logic, and output structure.
- **State Management**: The VM uses the x/merkledb integration to handle state changes. The state is managed using a merkelized radix tree, enabling efficient storage and fast lookup operations. The state update logic is handled centrally to ensure consistency.
- **Indexer Integration**: The VM supports an optional indexer to track blockchain events and provide historical data. Indexer settings, including the block window, can be configured through the indexerBlockWindow parameter in the configuration.
- **External Subscriber Integration**: The VM integrates with an external subscriber service to handle extended functionalities like event streaming and historical data retrieval. The externalSubscriberServerAddress is set based on configuration, enabling or disabling external subscriber services.
- **Execution Flow**: Actions are executed based on incoming transactions, with a key emphasis on maintaining state integrity. Non-conflicting actions are executed concurrently to optimize throughput, while conflicting actions are executed sequentially.

### RPC API

The RPC layer of NuklaiVM is implemented via a JSON-RPC interface, allowing clients to interact with the blockchain efficiently. Below are some of the key RPC methods supported by NuklaiVM, implemented in vm/client.go:

#### 1. Genesis: Retrieves the genesis state of the blockchain

- **Endpoint**: genesis
- **Description**: Returns the initial configuration and chain parameters.
- **Output**: Genesis information including rules and chain settings.

#### 2. Balance: Fetches the balance of a specific asset for a given address

- **Endpoint**: balance
- **Inputs**: Address, Asset address.
- **Output**: Balance of the specified asset for the address.

#### 3. Asset: Retrieves information about a specific asset

- **Endpoint**: asset
- **Inputs**: Asset address, Cache preference.
- **Output**: Asset details including name, symbol, supply, and permissions.

#### 4. Dataset : Fetches details of a dataset on the blockchain

- **Endpoint**: dataset
- **Inputs**: Dataset address, Cache preference.
- **Output**: Dataset metadata, owner, revenue model, and marketplace details.

#### 5. DatasetContribution: Fetches information about a specific dataset contribution

- **Endpoint**: datasetContribution
- **Inputs**: Contribution ID.
- **Output**: Contribution details including data location, contributor, and status.

#### 6. EmissionInfo: Retrieves current emission metrics for the native token NAI

- **Endpoint**: emissionInfo
- **Description**: Returns information about emissions, current supply, staking, and rewards.
- **Output**: Emission details including total supply, rewards per epoch, and staking information.

#### 7. AllValidators: Lists all validators currently registered on the network

- **Endpoint**: allValidators
- **Output**: List of all validators, including their staking status.

#### 8. StakedValidators: Lists validators with active stakes

- **Endpoint**: stakedValidators
- **Output**: List of staked validators.

#### 9. ValidatorStake: Retrieves staking details for a specific validator

- **Endpoint**: validatorStake
- **Inputs**: Validator Node ID.
- **Output**: Staking details such as staked amount, reward address, and delegation fee.

#### 10. UserStake: Retrieves the staking details of a user for a specific validator

- **Endpoint**: userStake
- **Inputs**: Owner address, Validator Node ID.
- **Output**: Stake details, including start and end blocks, staked amount, and reward information.

The RPC client is implemented using requester.EndpointRequester to manage communication with the VM, and the state is cached when appropriate to improve efficiency and reduce repeated requests.

### Storage Backends

HyperSDK allows flexibility in choosing storage backends. In NuklaiVM, state storage is typically managed on-disk using Pebble (CockroachDB's storage engine), but blocks and metadata can be stored in different backends, such as S3 or PostgreSQL, depending on the developer's requirements. This can be done by connecting the VM to an externalSubscriberServer.

### Storage Operations

The storage implementation of NuklaiVM focuses on managing assets, datasets, emission details, and programs. Each component is efficiently stored using HyperSDK's state management capabilities, ensuring high performance for both reads and writes.

- **Asset Storage**: Assets are managed with unique keys that include metadata like name, symbol, owner, and supply details. Functions like AssetInfoKey and SetAssetInfo are used to create and update asset data efficiently. For balance management, functions like SetAssetAccountBalance and GetAssetAccountBalanceFromState handle balance operations across multiple accounts.
- **Dataset Storage**: Datasets are represented with keys generated using DatasetInfoKey, which encapsulates all relevant details, including metadata, owner, and pricing. Dataset contributions are similarly managed using DatasetContributionInfoKey and related functions, enabling efficient contribution tracking and management.
- **Emission Storage**: Validator and delegator stake data are stored using unique keys like ValidatorStakeKey and DelegatorStakeKey, which manage staking and reward details efficiently. Functions like SetValidatorStake and GetValidatorStakeNoController handle staking operations, ensuring validator and delegator data is accurately recorded.
- **Program Storage**: Programs, including WASM-based contracts, are stored and managed via functions like StoreContract and

  GetContractBytes. This allows for efficient retrieval and storage of

  contract information, supporting the extensibility of NuklaiVM.

### VM and Execution

The VM wraps all components into a cohesive runtime that integrates action execution, state management, and account abstraction. Using the defaultvm package, the VM is constructed with a default set of services provided by HyperSDK, while also allowing custom RPC APIs.

### VM Lifecycle and Execution Flow

- **Transaction Parsing**: Incoming transactions are parsed to determine the action type. Each action is then routed to its respective handler for processing.
- **Concurrency Handling**: Actions are analyzed for potential conflicts using their StateKeys method. Non-conflicting actions are executed in parallel

  to maximize throughput.

- **State Updates**: After successful execution, state changes are committed. Failed transactions trigger a rollback, ensuring the state remains consistent.

### Transaction Handling

NuklaiVM uses nonce-less transactions that specify an expiration time. Transactions are processed in **FIFO (First In, First Out)** order, and if they cannot be executed within their expiration window, they are dropped. Transactions may also include **action batches**, which allow for multiple related actions to be executed atomically.

### Action Batches and Atomic Execution

- **Batch Execution**: Transactions can include multiple actions that must all succeed for the transaction to be valid. This ensures that complex operations, such as transferring assets followed by subscribing to a dataset, are completed without intermediate failure states.
- **Outputs and Events**: Each action within a batch can generate outputs or trigger events. These outputs are used for further processing or integration with off-chain systems, such as notifying users of successful actions.

## Conclusion

NuklaiVM is a powerful blockchain Virtual Machine that integrates complex features like dataset management, marketplace economics, staking, and tokenized contributions directly within its VM environment. By leveraging HyperSDK, NuklaiVM provides a high-performance, low-latency blockchain solution with extensive capabilities for data-driven applications and community interactions. The emission balancer and staking mechanism ensure network stability and fair reward distribution, making NuklaiVM a robust and sustainable platform for decentralized applications.
