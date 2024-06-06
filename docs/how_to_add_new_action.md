# How to add a new action to a hypervm

![How actions are executed](./how_actions_are_executed.png)

Let's go through the process of adding a new action to a hypervm by implementing a "Withdraw Validator Stake" action. We need to add functionality to both the core vm code and also include it as part of RPC API so external users can interact with the VM easily.

## HyperVM

Since we are going to define our action as part of the VM itself, we need to make changes to the core VM code for nuklaivm.

### 1. registry/registry.go

Register the new action to our registry

```go
  consts.ActionRegistry.Register((&actions.UndelegateUserStake{}).GetTypeID(), actions.UnmarshalUndelegateUserStake, false),
```

### 2. actions/undelegate_user_stake.go

- Create a new file called `undelegate_user_stake.go`. Here, we need to define some functions that complies with the Action interface defined at [https://github.com/ava-labs/hypersdk/blob/main/chain/dependencies.go#L206](https://github.com/ava-labs/hypersdk/blob/main/chain/dependencies.go#L206)

- We need to define the following functions:

```go
type Action interface {
 Object

 // ComputeUnits is the amount of compute required to call [Execute]. This is used to determine
 // whether the [Action] can be included in a given block and to compute the required fee to execute.
 ComputeUnits(Rules) uint64

 // StateKeysMaxChunks is used to estimate the fee a transaction should pay. It includes the max
 // chunks each state key could use without requiring the state keys to actually be provided (may
 // not be known until execution).
 StateKeysMaxChunks() []uint16

 // StateKeys is a full enumeration of all database keys that could be touched during execution
 // of an [Action]. This is used to prefetch state and will be used to parallelize execution (making
 // an execution tree is trivial).
 //
 // All keys specified must be suffixed with the number of chunks that could ever be read from that
 // key (formatted as a big-endian uint16). This is used to automatically calculate storage usage.
 //
 // If any key is removed and then re-created, this will count as a creation instead of a modification.
 StateKeys(actor codec.Address, actionID ids.ID) state.Keys

 // Execute actually runs the [Action]. Any state changes that the [Action] performs should
 // be done here.
 //
 // If any keys are touched during [Execute] that are not specified in [StateKeys], the transaction
 // will revert and the max fee will be charged.
 //
 // If [Execute] returns an error, execution will halt and any state changes will revert.
 Execute(
  ctx context.Context,
  r Rules,
  mu state.Mutable,
  timestamp int64,
  actor codec.Address,
  actionID ids.ID,
 ) (outputs [][]byte, err error)
}
```

- Our actions/undelegate_user_stake.go now looks like this:

```go
// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
 "context"

 "github.com/ava-labs/avalanchego/ids"
 "github.com/ava-labs/hypersdk/chain"
 "github.com/ava-labs/hypersdk/codec"
 "github.com/ava-labs/hypersdk/consts"
 "github.com/ava-labs/hypersdk/state"

 nconsts "github.com/nuklai/nuklaivm/consts"
 "github.com/nuklai/nuklaivm/emission"
 "github.com/nuklai/nuklaivm/storage"
)

var _ chain.Action = (*UndelegateUserStake)(nil)

type UndelegateUserStake struct {
 NodeID        []byte        `json:"nodeID"`        // Node ID of the validator where NAI is staked
 RewardAddress codec.Address `json:"rewardAddress"` // Address to receive rewards

 // TODO: add boolean to indicate whether sender will
 // create recipient account
}

func (*UndelegateUserStake) GetTypeID() uint8 {
 return nconsts.UndelegateUserStakeID
}

func (u *UndelegateUserStake) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
 // TODO: How to better handle a case where the NodeID is invalid?
 nodeID, _ := ids.ToNodeID(u.NodeID)
 return state.Keys{
  string(storage.BalanceKey(actor, ids.Empty)):           state.Read | state.Write,
  string(storage.BalanceKey(u.RewardAddress, ids.Empty)): state.All,
  string(storage.DelegateUserStakeKey(actor, nodeID)):    state.Read | state.Write,
 }
}

func (*UndelegateUserStake) StateKeysMaxChunks() []uint16 {
 return []uint16{storage.BalanceChunks, storage.DelegateUserStakeChunks}
}

func (*UndelegateUserStake) OutputsWarpMessage() bool {
 return false
}

func (u *UndelegateUserStake) Execute(
 ctx context.Context,
 _ chain.Rules,
 mu state.Mutable,
 _ int64,
 actor codec.Address,
 _ ids.ID,
) ([][]byte, error) {
 nodeID, err := ids.ToNodeID(u.NodeID)
 if err != nil {
  return nil, ErrOutputInvalidNodeID
 }

 exists, _, stakeEndBlock, stakedAmount, _, ownerAddress, _ := storage.GetDelegateUserStake(ctx, mu, actor, nodeID)
 if !exists {
  return nil, ErrOutputStakeMissing
 }
 if ownerAddress != actor {
  return nil, ErrOutputUnauthorized
 }

 // Get the emission instance
 emissionInstance := emission.GetEmission()

 // Check that lastBlockHeight is after stakeEndBlock
 if emissionInstance.GetLastAcceptedBlockHeight() < stakeEndBlock {
  return nil, ErrOutputStakeNotEnded
 }

 // Undelegate in Emission Balancer
 rewardAmount, err := emissionInstance.UndelegateUserStake(nodeID, actor)
 if err != nil {
  return nil, err
 }
 if err := storage.AddBalance(ctx, mu, u.RewardAddress, ids.Empty, rewardAmount, true); err != nil {
  return nil, err
 }

 if err := storage.DeleteDelegateUserStake(ctx, mu, ownerAddress, nodeID); err != nil {
  return nil, err
 }
 if err := storage.AddBalance(ctx, mu, ownerAddress, ids.Empty, stakedAmount, true); err != nil {
  return nil, err
 }

 sr := &UndelegateUserStakeResult{stakedAmount, rewardAmount}
 output, err := sr.Marshal()
 if err != nil {
  return nil, err
 }
 return [][]byte{output}, nil
}

func (*UndelegateUserStake) ComputeUnits(chain.Rules) uint64 {
 return UndelegateUserStakeComputeUnits
}

func (*UndelegateUserStake) Size() int {
 return ids.NodeIDLen + codec.AddressLen
}

func (u *UndelegateUserStake) Marshal(p *codec.Packer) {
 p.PackBytes(u.NodeID)
 p.PackAddress(u.RewardAddress)
}

func UnmarshalUndelegateUserStake(p *codec.Packer) (chain.Action, error) {
 var unstake UndelegateUserStake
 p.UnpackBytes(ids.NodeIDLen, true, &unstake.NodeID)
 p.UnpackAddress(&unstake.RewardAddress)
 return &unstake, p.Err()
}

func (*UndelegateUserStake) ValidRange(chain.Rules) (int64, int64) {
 // Returning -1, -1 means that the action is always valid.
 return -1, -1
}

type UndelegateUserStakeResult struct {
 StakedAmount uint64
 RewardAmount uint64
}

func UnmarshalUndelegateUserStakeResult(b []byte) (*UndelegateUserStakeResult, error) {
 p := codec.NewReader(b, 2*consts.Uint64Len)
 var result UndelegateUserStakeResult
 result.StakedAmount = p.UnpackUint64(true)
 result.RewardAmount = p.UnpackUint64(false)
 return &result, p.Err()
}

func (s *UndelegateUserStakeResult) Marshal() ([]byte, error) {
 p := codec.NewWriter(2*consts.Uint64Len, 2*consts.Uint64Len)
 p.PackUint64(s.StakedAmount)
 p.PackUint64(s.RewardAmount)
 return p.Bytes(), p.Err()
}
```

### 3. consts/types.go

We need to add a new ID for this new action which we are referencing on actions/undelegate_user_stake.go. We can define this ID on consts/types.go:

```go
UndelegateUserStakeID        uint8 = 11
```

### 4. actions/consts.go

We need to add a new variable called "UnstakeValidatorComputeUnits" that we referenced on actions/undelegate_user_stake.go that defines the compute units it's going to cost the user to perform this action. We can define this on actions/consts.go:

```go
UndelegateUserStakeComputeUnits    = 1
```

### 5. actions/outputs.go

We also need to add some error definitions which were referenced on actions/undelegate_user_stake.go. We can define these on actions/outputs.go:

```go
ErrOutputInvalidNodeID = errors.New("invalid node ID")
ErrOutputStakeMissing  = errors.New("stake is missing")
ErrOutputUnauthorized       = errors.New("unauthorized")
ErrOutputStakeNotEnded = errors.New("stake not ended")
```

### 6. controller/controller.go

The Controller is the entry point of nuklaivm. It initializes the data structures utilized by the hypersdk and handles both Accepted and Rejected block callbacks.

Let's make sure to handle additional logic needed for our unstake validator action.

Under `Accepted` function right after tx is successful, let's call `UnstakeFromValidator` from our Emission Balancer so it calculates the staked amount from the validator accordingly.

```go
    case *actions.UndelegateUserStake:
     c.metrics.claimStakingRewards.Inc()
     c.metrics.undelegateUserStake.Inc()
     outputs := result.Outputs[i]
     for _, output := range outputs {
      stakeResult, err := actions.UnmarshalUndelegateUserStakeResult(output)
      if err != nil {
       // This should never happen
       return err
      }
      c.metrics.delegatorStakeAmount.Sub(float64(stakeResult.StakedAmount))
      c.metrics.mintedNAI.Add(float64(stakeResult.RewardAmount))
      c.metrics.rewardAmount.Add(float64(stakeResult.RewardAmount))
     }
```

### 7. controller/metrics.go

We need to add a new metric for our unstake validator action that we referenced on controller/controller.go that defines the number of unstake actions.

```go
type metrics struct {
 ...
 unstake prometheus.Counter
  ...
}

func newMetrics(gatherer ametrics.MultiGatherer) (*metrics, error) {
  m := &metrics{
    ...
  undelegateUserStake: prometheus.NewCounter(prometheus.CounterOpts{
   Namespace: "actions",
   Name:      "undelegate_user_stake",
   Help:      "number of undelegate user stake actions",
  }),
 }
 ...
 errs.Add(
  ...
  r.Register(m.undelegateUserStake),
    ...
 )
 ...
}
```

### 8. emission/emission.go

We now need to define a new function called `UnstakeFromValidator` that will unstake the NAI tokens from the given validator

```go
// UndelegateUserStake decreases the delegated stake for a validator and rebalances the heap.
func (e *Emission) UndelegateUserStake(nodeID ids.NodeID, actor codec.Address) (uint64, error) {
 e.lock.Lock()
 defer e.lock.Unlock()

 e.c.Logger().Info("undelegating user stake",
  zap.String("nodeID", nodeID.String()))

 // Find the validator
 validator, exists := e.validators[nodeID]
 if !exists {
  return 0, ErrValidatorNotFound
 }

 // Check if the delegator exists
 _, exists = validator.delegators[actor]
 if !exists {
  e.c.Logger().Error("delegator not found")
  return 0, ErrDelegatorNotFound
 }

 // Calculate rewards while undelegating
 rewardAmount, err := e.CalculateUserDelegationRewards(nodeID, actor)
 if err != nil {
  e.c.Logger().Error("error calculating rewards", zap.Error(err))
  return 0, err
 }
 // Ensure AccumulatedDelegatedReward does not become negative
 if rewardAmount > validator.AccumulatedDelegatedReward {
  rewardAmount = validator.AccumulatedDelegatedReward
 }
 validator.AccumulatedDelegatedReward -= rewardAmount

 // Remove the delegator from the list
 delete(validator.delegators, actor)

 // If the validator is inactive and has withdrawn and has no more delegators, remove the validator
 if !validator.IsActive && validator.StakedAmount == 0 && len(validator.delegators) == 0 {
  e.c.Logger().Info("removing validator",
   zap.String("nodeID", nodeID.String()))
  delete(e.validators, nodeID)
 }

 e.c.Logger().Info("undelegated user stake",
  zap.String("nodeID", nodeID.String()),
  zap.Uint64("rewardAmount", rewardAmount))

 return rewardAmount, nil
}
```

### 9. emission/errors.go

We need to add some error definitions which were referenced on emission/emission.go. We can define these on emission/errors.go:

```go
 ErrValidatorNotFound          = errors.New("validator not found")
 ErrDelegatorNotFound          = errors.New("delegator not found")
```

## RPC API

We technically do not need to define any logic for our RPC API if all we want is for users to call this action we defined above however, if you want to add additional helper functions, we can define them easily via RPC API. An example could be if you wanted to define an RPC API to get the currently staked validators info such as the total amount staked, delegated amount, etc. and also to check out the stake of a delegator.

### 1. rpc/dependencies.go

Let's define the function definitions we want exposed to external users via our RPC API.

```go
type Controller interface {
  ...
  GetStakedValidatorInfo(nodeID ids.NodeID) (*emission.Validator, error)
  GetDelegatedUserStakeFromState(ctx context.Context, owner codec.Address, nodeID ids.NodeID) (
  bool, // exists
  uint64, // StakeStartBlock
  uint64, // StakeEndBlock
  uint64, // StakedAmount
  codec.Address, // RewardAddress
  codec.Address, // OwnerAddress
  error,
 )
  ...
}
```

### 2. controller/resolutions.go

Now, it's time to implement the functions we defined on rpc/dependencies.go.

```go
func (c *Controller) GetStakedValidatorInfo(nodeID ids.NodeID) (*emission.Validator, error) {
 validators := c.emission.GetStakedValidator(nodeID)
 return validators[0], nil
}
func (c *Controller) GetDelegatedUserStakeFromState(ctx context.Context, owner codec.Address, nodeID ids.NodeID) (
 bool, // exists
 uint64, // StakeStartBlock
 uint64, // StakeEndBlock
 uint64, // StakedAmount
 codec.Address, // RewardAddress
 codec.Address, // OwnerAddress
 error,
) {
 return storage.GetDelegateUserStakeFromState(ctx, c.inner.ReadState, owner, nodeID)
}
```

### 3. emission/emission.go

Let's implement the function `GetUserStake` on our Emission Balancer.

```go
// GetStakedValidator retrieves the details of a specific validator by their NodeID.
func (e *Emission) GetStakedValidator(nodeID ids.NodeID) []*Validator {
 e.c.Logger().Info("fetching staked validator")

 if nodeID == ids.EmptyNodeID {
  validators := make([]*Validator, 0, len(e.validators))
  for _, validator := range e.validators {
   validators = append(validators, validator)
  }
  return validators
 }

 // Find the validator
 if validator, exists := e.validators[nodeID]; exists {
  return []*Validator{validator}
 }
 return []*Validator{}
}
```

### 4. rpc/jsonrpc_client.go

We need to define a new function on our RPC Client so users can call this API via external tools like curl, POSTMAN, or third party applications. We can do this on rpc/jsonrpc_client.go:

```go
func (cli *JSONRPCClient) StakedValidators(ctx context.Context) ([]*emission.Validator, error) {
 resp := new(ValidatorsReply)
 err := cli.requester.SendRequest(
  ctx,
  "stakedValidators",
  nil,
  resp,
 )
 if err != nil {
  return []*emission.Validator{}, err
 }
 return resp.Validators, err
}

func (cli *JSONRPCClient) UserStake(ctx context.Context, owner codec.Address, nodeID ids.NodeID) (uint64, uint64, uint64, codec.Address, codec.Address, error) {
 resp := new(UserStakeReply)
 err := cli.requester.SendRequest(
  ctx,
  "userStake",
  &UserStakeArgs{
   Owner:  owner,
   NodeID: nodeID,
  },
  resp,
 )
 if err != nil {
  return 0, 0, 0, codec.EmptyAddress, codec.EmptyAddress, err
 }
 return resp.StakeStartBlock, resp.StakeEndBlock, resp.StakedAmount, resp.RewardAddress, resp.OwnerAddress, err
}
```

### 5. rpc/jsonrpc_server.go

We need to also define a corresponding function on our RPC server so whenever users interact with the API from their client, it talks to this server function which in turn calls the function defined in controller/resolutions.go. We can do this on rpc/jsonrpc_server.go:

```go
type ValidatorsReply struct {
 Validators []*emission.Validator `json:"validators"`
}

func (j *JSONRPCServer) StakedValidators(req *http.Request, _ *struct{}, reply *ValidatorsReply) (err error) {
 ctx, span := j.c.Tracer().Start(req.Context(), "Server.StakedValidators")
 defer span.End()

 validators, err := j.c.GetValidators(ctx, true)
 if err != nil {
  return err
 }
 reply.Validators = validators
 return nil
}

type UserStakeArgs struct {
 Owner  codec.Address `json:"owner"`
 NodeID ids.NodeID    `json:"nodeID"`
}

type UserStakeReply struct {
 StakeStartBlock uint64        `json:"stakeStartBlock"` // Start block of the stake
 StakeEndBlock   uint64        `json:"stakeEndBlock"`   // End block of the stake
 StakedAmount    uint64        `json:"stakedAmount"`    // Amount of NAI staked
 RewardAddress   codec.Address `json:"rewardAddress"`   // Address to receive rewards
 OwnerAddress    codec.Address `json:"ownerAddress"`    // Address of the owner who delegated
}

func (j *JSONRPCServer) UserStake(req *http.Request, args *UserStakeArgs, reply *UserStakeReply) (err error) {
 ctx, span := j.c.Tracer().Start(req.Context(), "Server.UserStake")
 defer span.End()

 exists, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, ownerAddress, err := j.c.GetDelegatedUserStakeFromState(ctx, args.Owner, args.NodeID)
 if err != nil {
  return err
 }
 if !exists {
  return ErrUserStakeNotFound
 }

 reply.StakeStartBlock = stakeStartBlock
 reply.StakeEndBlock = stakeEndBlock
 reply.StakedAmount = stakedAmount
 reply.RewardAddress = rewardAddress
 reply.OwnerAddress = ownerAddress
 return nil
}
```

## nuklai-cli

In order to easily test the capability of our new action and our new RPC API, we can integrate them in our nuklai-cli tool. This is not needed but highly encouraged because often times, external users will interact with our VM and giving developers the option to test their new actions via a command line tool is paramount. Think of `nuklai-cli` as a third party application that lets developers quickly interact with different `nuklaivm` features such as the new action we defined above or the `GetUserStake` function we defined above.

### 1. cmd/nuklai-cli/cmd/action.go

Let's define a new command to let users unstake their NAI tokens from the validator they have staked to in the past.

```go
var undelegateUserStakeCmd = &cobra.Command{
 Use: "undelegate-user-stake",
 RunE: func(*cobra.Command, []string) error {
  ctx := context.Background()
  _, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
  if err != nil {
   return err
  }

  // Get current list of validators
  validators, err := ncli.StakedValidators(ctx)
  if err != nil {
   return err
  }
  if len(validators) == 0 {
   hutils.Outf("{{red}}no validators{{/}}\n")
   return nil
  }

  // Show validators to the user
  hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
  for i := 0; i < len(validators); i++ {
   hutils.Outf(
    "{{yellow}}%d:{{/}} NodeID=%s\n",
    i,
    validators[i].NodeID,
   )
  }
  // Select validator
  keyIndex, err := handler.Root().PromptChoice("validator to unstake from", len(validators))
  if err != nil {
   return err
  }
  validatorChosen := validators[keyIndex]
  nodeID := validatorChosen.NodeID

  // Get stake info
  _, _, stakedAmount, _, _, err := ncli.UserStake(ctx, priv.Address, nodeID)
  if err != nil {
   return err
  }

  if stakedAmount == 0 {
   hutils.Outf("{{red}}user has not yet delegated to this validator{{/}}\n")
   return nil
  }

  // Confirm action
  cont, err := handler.Root().PromptContinue()
  if !cont || err != nil {
   return err
  }

  // Generate transaction
  _, err = sendAndWait(ctx, []chain.Action{&actions.UndelegateUserStake{
   NodeID:        nodeID.Bytes(),
   RewardAddress: priv.Address,
  }}, hcli, hws, ncli, factory, true)
  return err
 },
}
```

### 2. cmd/nuklai-cli/cmd/emission.go

Let's now define a new command to let users easily check their current stake on a chosen validator. We can do this on cmd/nuklai-cli/cmd/emission.go:

```go
var emissionStakedValidatorsCmd = &cobra.Command{
 Use: "staked-validators",
 RunE: func(_ *cobra.Command, args []string) error {
  ctx := context.Background()

  // Get clients
  nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
  if err != nil {
   return err
  }
  ncli := nclients[0]

  // Get validators info
  _, err = handler.GetStakedValidators(ctx, ncli)
  if err != nil {
   return err
  }

  return nil
 },
}
```

### 3. cmd/nuklai-cli/cmd/root.go

We need to add these two new commands to root.go so it's available when users interact with `nuklai-cli`

```go
actionCmd.AddCommand(
  ...
  undelegateUserStakeCmd,
 )
...
emissionCmd.AddCommand(
  ...
  emissionStakedValidatorsCmd,
 )
```

### 4. cmd/nuklai-cli/cmd/handler.go

There is nothing left to do for our unstake validator action however, for the RPC API to get user stake, we need to define this function on handler.go so any other functions can call this function if need be. This is not needed but this is good practice so that multiple functions can reuse the same function.

```go
func (*Handler) GetStakedValidators(
 ctx context.Context,
 cli *nrpc.JSONRPCClient,
) ([]*emission.Validator, error) {
 validators, err := cli.StakedValidators(ctx)
 if err != nil {
  return nil, err
 }
 for index, validator := range validators {
  publicKey, err := bls.PublicKeyFromBytes(validator.PublicKey)
  if err != nil {
   return nil, err
  }
  hutils.Outf(
   "{{yellow}}validator %d:{{/}} NodeID=%s PublicKey=%s Active=%t StakedAmount=%d AccumulatedStakedReward=%d DelegationFeeRate=%f DelegatedAmount=%d AccumulatedDelegatedReward=%d\n",
   index,
   validator.NodeID,
   base64.StdEncoding.EncodeToString(publicKey.Compress()),
   validator.IsActive,
   validator.StakedAmount,
   validator.AccumulatedStakedReward,
   validator.DelegationFeeRate,
   validator.DelegatedAmount,
   validator.AccumulatedDelegatedReward,
  )
 }
 return validators, nil
}

func (*Handler) GetUserStake(ctx context.Context,
 cli *nrpc.JSONRPCClient, owner codec.Address, nodeID ids.NodeID,
) (uint64, uint64, uint64, string, string, error) {
 stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, ownerAddress, err := cli.UserStake(ctx, owner, nodeID)
 if err != nil {
  return 0, 0, 0, "", "", err
 }

 rewardAddressString, err := codec.AddressBech32(nconsts.HRP, rewardAddress)
 if err != nil {
  return 0, 0, 0, "", "", err
 }
 ownerAddressString, err := codec.AddressBech32(nconsts.HRP, ownerAddress)
 if err != nil {
  return 0, 0, 0, "", "", err
 }

 hutils.Outf(
  "{{yellow}}user stake: {{/}}\nStakeStartBlock=%d StakeEndBlock=%d StakedAmount=%d RewardAddress=%s OwnerAddress=%s\n",
  stakeStartBlock,
  stakeEndBlock,
  stakedAmount,
  rewardAddressString,
  ownerAddressString,
 )
 return stakeStartBlock,
  stakeEndBlock,
  stakedAmount,
  rewardAddressString,
  ownerAddressString, err
}
```

This basically prints the user stake to the screen. This is especially useful on our `nuklai-cli` because we can quickly call the RPC API for getting user stake this way.

## Conclusion

That's it! Now, let's see this in action by building `nuklai-cli` and `nuklaivm` and running the vm in our subnet.

The following info is also available as part of the main README.md in the repository.

### 1. Launch Subnet

The first step to running this demo is to launch your own `nuklaivm` Subnet. You
can do so by running the following command from this location (may take a few
minutes):

```bash
./scripts/run.sh;
```

### 2. Build `nuklai-cli`

To make it easy to interact with the `nuklaivm`, we implemented the `nuklai-cli`.
Next, you'll need to build this tool. You can use the following command:

```bash
./scripts/build.sh
```

### 3. Configure `nuklai-cli`

Next, you'll need to add the chains you created and the default key to the
`nuklai-cli`. You can use the following commands from this location to do so:

```bash
./build/nuklai-cli key import ed25519 demo.pk
```

If the key is added correctly, you'll see the following log:

```
database: .nuklai-cli
imported address: created address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
```

Next, you'll need to store the URLs of the nodes running on your Subnet:

```bash
./build/nuklai-cli chain import-anr
```

If `nuklai-cli` is able to connect to ANR, it will emit the following logs:

```
database: .nuklai-cli
stored chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF uri: http://127.0.0.1:43383/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
stored chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF uri: http://127.0.0.1:41523/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
stored chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF uri: http://127.0.0.1:38419/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
stored chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF uri: http://127.0.0.1:39943/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
stored chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF uri: http://127.0.0.1:41743/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
```

_`./build/nuklai-cli chain import-anr` connects to the Avalanche Network Runner server running in
the background and pulls the URIs of all nodes tracking each chain you
created._

### 4. Stake to a validator

We can stake to a validator of our choice

```bash
./build/nuklai-cli action delegate-user-stake [auto | manual]
```

If successful, the output should be something like:

```
validators: 2
0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
validator to delegate to: 0
balance: 1000000.000000000 NAI
✔ continue (y/n): y█
✅ txID: 2RosZKQY3K74Q9i1yJfm6DdvMPMsSbZCLFke9v2oFQEEBFAMT5
```

### 5. Get user staking info

We can retrieve our staking info by passing in which validator we have staked to and the address to look up staking for using the new RPC API we defined as part of this exercise.

```bash
./build/nuklai-cli action get-user-stake
```

If successful, the output should be something like:

```
validators: 2
0: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
1: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
validator to get staking info for: 0
validator stake:
StakeStartBlock=100 StakedAmount=100000000000000 RewardAddress=nuklai1qgtvmjhh5xkjh5tf993s05fptc2l0mzn6j8yw72pmrqpa947xsp5scqsrma OwnerAddress=nuklai1qgtvmjhh5xkjh5tf993s05fptc2l0mzn6j8yw72pmrqpa947xsp5scqsrma
```

### 6. Unstake from a validator

We can unstake specific stake from a chosen validator.

```bash
./build/nuklai-cli action undelegate-user-stake
```

If successful, the output should be something like:

```
validators: 2
0: NodeID=NodeID-JV548bkici8bBx1SzvSCUKZdgP3RY3iXs
1: NodeID=NodeID-9aaVYT33M2GPAws7eSjHor5c3zLhkngy9
✔ validator to unstake from: 1█
continue (y/n): y
✅ txID: bTFRsFwyMJT4sESishFHeAihL9SnBFvirrSExp95eJTFVLyLz
```
