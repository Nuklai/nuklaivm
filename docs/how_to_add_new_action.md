# How to add a new action to a hypervm

Let's go through the process of adding a new action to a hypervm by implementing a "unstake_validator" action. We need to add functionality to both the core vm code and also include it as part of RPC API so external users can interact with the VM easily.

## HyperVM

Since we are going to define our action as part of the VM itself, we need to make changes to the core VM code for nuklaivm.

### 1. registry/registry.go

Register the new action to our registry

```go
consts.ActionRegistry.Register((&actions.UnstakeValidator{}).GetTypeID(), actions.UnmarshalUnstakeValidator, false),
```

### 2. actions/unstake_validator.go

- Create a new file called "unstake_validator.go". Here, we need to define some functions that complies with the Action interface defined at [https://github.com/ava-labs/hypersdk/blob/main/chain/dependencies.go#L171C1-L171C1](https://github.com/ava-labs/hypersdk/blob/main/chain/dependencies.go#L171C1-L171C1)

- We need to define the following functions:

```go
type Action interface {
	// GetTypeID uniquely identifies each supported [Action]. We use IDs to avoid
	// reflection.
	GetTypeID() uint8

	// ValidRange is the timestamp range (in ms) that this [Action] is considered valid.
	//
	// -1 means no start/end
	ValidRange(Rules) (start int64, end int64)

	// MaxComputeUnits is the maximum amount of compute a given [Action] could use. This is
	// used to determine whether the [Action] can be included in a given block and to compute
	// the required fee to execute.
	//
	// Developers should make every effort to bound this as tightly to the actual max so that
	// users don't need to have a large balance to call an [Action] (must prepay fee before execution).
	MaxComputeUnits(Rules) uint64

	// OutputsWarpMessage indicates whether an [Action] will produce a warp message. The max size
	// of any warp message is [MaxOutgoingWarpChunks].
	OutputsWarpMessage() bool

	// StateKeys is a full enumeration of all database keys that could be touched during execution
	// of an [Action]. This is used to prefetch state and will be used to parallelize execution (making
	// an execution tree is trivial).
	//
	// All keys specified must be suffixed with the number of chunks that could ever be read from that
	// key (formatted as a big-endian uint16). This is used to automatically calculate storage usage.
	//
	// If any key is removed and then re-created, this will count as a creation instead of a modification.
	StateKeys(auth Auth, txID ids.ID) []string

	// StateKeysMaxChunks is used to estimate the fee a transaction should pay. It includes the max
	// chunks each state key could use without requiring the state keys to actually be provided (may
	// not be known until execution).
	StateKeysMaxChunks() []uint16

	// Execute actually runs the [Action]. Any state changes that the [Action] performs should
	// be done here.
	//
	// If any keys are touched during [Execute] that are not specified in [StateKeys], the transaction
	// will revert and the max fee will be charged.
	//
	// An error should only be returned if a fatal error was encountered, otherwise [success] should
	// be marked as false and fees will still be charged.
	Execute(
		ctx context.Context,
		r Rules,
		mu state.Mutable,
		timestamp int64,
		auth Auth,
		txID ids.ID,
		warpVerified bool,
	) (success bool, computeUnits uint64, output []byte, warpMessage *warp.UnsignedMessage, err error)

	// Marshal encodes an [Action] as bytes.
	Marshal(p *codec.Packer)

	// Size is the number of bytes it takes to represent this [Action]. This is used to preallocate
	// memory during encoding and to charge bandwidth fees.
	Size() int
}
```

- Our actions/unstake_validator.go now looks like this:

```go
// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/storage"

	mconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*UnstakeValidator)(nil)

type UnstakeValidator struct {
	Stake  ids.ID `json:"stake"`
	NodeID []byte `json:"nodeID"`
}

func (*UnstakeValidator) GetTypeID() uint8 {
	return mconsts.UnstakeValidatorID
}

func (u *UnstakeValidator) StateKeys(auth chain.Auth, _ ids.ID) []string {
	return []string{
		string(storage.BalanceKey(auth.Actor())),
		string(storage.StakeKey(u.Stake)),
	}
}

func (*UnstakeValidator) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.StakeChunks}
}

func (*UnstakeValidator) OutputsWarpMessage() bool {
	return false
}

func (u *UnstakeValidator) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	auth chain.Auth,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	exists, nodeStaked, stakedAmount, _, owner, err := storage.GetStake(ctx, mu, u.Stake)
	if err != nil {
		return false, UnstakeValidatorComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if !exists {
		return false, UnstakeValidatorComputeUnits, OutputStakeMissing, nil, nil
	}
	if owner != auth.Actor() {
		return false, UnstakeValidatorComputeUnits, OutputUnauthorized, nil, nil
	}
	nodeID, err := ids.ToNodeID(u.NodeID)
	if err != nil {
		return false, UnstakeValidatorComputeUnits, OutputInvalidNodeID, nil, nil
	}
	if nodeStaked != nodeID {
		return false, UnstakeValidatorComputeUnits, OutputDifferentNodeIDThanStaked, nil, nil
	}
	if err := storage.DeleteStake(ctx, mu, u.Stake); err != nil {
		return false, UnstakeValidatorComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.AddBalance(ctx, mu, auth.Actor(), stakedAmount, true); err != nil {
		return false, UnstakeValidatorComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, UnstakeValidatorComputeUnits, nil, nil, nil
}

func (*UnstakeValidator) MaxComputeUnits(chain.Rules) uint64 {
	return UnstakeValidatorComputeUnits
}

func (*UnstakeValidator) Size() int {
	return consts.IDLen
}

func (u *UnstakeValidator) Marshal(p *codec.Packer) {
	p.PackID(u.Stake)
}

func UnmarshalUnstakeValidator(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var unstake UnstakeValidator
	p.UnpackID(true, &unstake.Stake)
	return &unstake, p.Err()
}

func (*UnstakeValidator) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
```

### 3. consts/types.go

We need to add a new ID for this new action which we are referencing on actions/unstake_validator.go. We can define this ID on consts/types.go:

```go
UnstakeValidatorID uint8 = 2
```

### 4. actions/consts.go

We need to add a new variable called "UnstakeValidatorComputeUnits" that we referenced on actions/unstake_validator.go that defines the compute units it's going to cost the user to perform this action. We can define this on actions/consts.go:

```go
UnstakeValidatorComputeUnits = 5
```

### 5. actions/outputs.go

We also need to add some error definitions which were referenced on actions/unstake_validator.go. We can define these on actions/outputs.go:

```go
	OutputStakeMissing              = []byte("stake is missing")
	OutputUnauthorized              = []byte("unauthorized")
	OutputInvalidNodeID             = []byte("invalid node ID")
	OutputDifferentNodeIDThanStaked = []byte("node ID is different than staked")
```

### 6. controller/controller.go

The Controller is the entry point of nuklaivm. It initializes the data structures utilized by the hypersdk and handles both Accepted and Rejected block callbacks.

Let's make sure to handle additional logic needed for our unstake validator action.

Under `Accepted` function right after tx is successful, let's call `UnstakeFromValidator` from our Emission Balancer so it calculates the staked amount from the validator accordingly.

```go
    case *actions.UnstakeValidator:
				c.metrics.unstake.Inc()
				// Check to make sure the unstake is valid
				_, _, _, endLockUp, _, _ := storage.GetStake(ctx, mu, action.Stake)
				if endLockUp < c.inner.LastAcceptedBlock().Height() {
					return fmt.Errorf("end lockup %d is less than current block height %d", endLockUp, c.inner.LastAcceptedBlock().Height())
				}
				err := c.emission.UnstakeFromValidator(tx.ID(), tx.Auth.Actor(), action)
				if err != nil {
					return err
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
		unstake: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "unstake",
			Help:      "number of unstake actions",
		}),
	}
	...
	errs.Add(
		...
		r.Register(m.unstake),
    ...
	)
	...
}
```

### 8. emission/emission.go

We now need to define a new function called `UnstakeFromValidator` that will unstake the NAI tokens from the given validator

```go
func (e *Emission) UnstakeFromValidator(txID ids.ID, actor codec.Address, action *actions.UnstakeValidator) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	nodeID, err := ids.ToNodeID(action.NodeID)
	if err != nil {
		return ErrInvalidNodeID // Invalid NodeID
	}

	stakeOwner := codec.MustAddressBech32(consts.HRP, actor)
	validator, ok := e.validators[nodeID.String()]
	if !ok {
		return ErrNotAValidator // Not a validator
	}
	userStake, ok := validator.UserStake[stakeOwner]
	if !ok {
		return ErrUserNotStaked // User is not staked
	}
	stakeInfo, ok := userStake.StakeInfo[txID.String()]
	if !ok {
		return ErrStakeNotFound // Stake not found
	}

	// Reduce the staked amount from the userstake
	userStake.StakedAmount -= stakeInfo.Amount
	// Reduce the staked amount from the validator
	validator.StakedAmount -= stakeInfo.Amount
	// Remove the stake info
	delete(userStake.StakeInfo, txID.String())
	// Remove the user stake if there are no more stakes
	if len(userStake.StakeInfo) == 0 {
		delete(validator.UserStake, stakeOwner)
	}
	// Remove the validator if the staked amount is 0
	if validator.StakedAmount == 0 {
		delete(e.validators, nodeID.String())
	}

	return nil
}
```

### 9. emission/errors.go

We need to add some error definitions which were referenced on emission/emission.go. We can define these on emission/errors.go:

```go
	ErrNotAValidatorOwner       = errors.New("not a validator owner")
	ErrUserNotStaked            = errors.New("user not staked")
```

## RPC API

We technically do not need to define any logic for our RPC API if all we want is for users to call this action we defined above however, if you want to add additional helper functions, we can define them easily via RPC API. An example could be if you wanted to define an RPC API to get the current user stake of the user or maybe you want to add an API to let validator owners to claim their rewards.

### 1. rpc/dependencies.go

Let's define the function definitions we want exposed to external users via our RPC API.

```go
type Controller interface {
  ...
  GetUserStake(nodeID string, owner string) (*emission.UserStake, error)
  ...
}
```

### 2. controller/resolutions.go

Now, it's time to implement the functions we defined on rpc/dependencies.go.

```go
func (c *Controller) GetUserStake(nodeID string, owner string) (*emission.UserStake, error) {
	return c.emission.GetUserStake(nodeID, owner), nil
}
```

### 3. emission/emission.go

Let's implement the function `GetUserStake` on our Emission Balancer.

```go
func (e *Emission) GetUserStake(nodeID, owner string) *UserStake {
	e.lock.RLock()
	defer e.lock.RUnlock()

	validator, ok := e.validators[nodeID]
	if !ok {
		return &UserStake{}
	}

	userStake, ok := validator.UserStake[owner]
	if !ok {
		return &UserStake{}
	}
	return userStake
}
```

### 4. rpc/jsonrpc_client.go

We need to define a new function on our RPC Client so users can call this API via external tools like curl, POSTMAN, or third party applications. We can do this on rpc/jsonrpc_client.go:

```go
func (cli *JSONRPCClient) Validators(ctx context.Context) ([]*emission.Validator, error) {
	resp := new(ValidatorsReply)
	err := cli.requester.SendRequest(
		ctx,
		"validators",
		nil,
		resp,
	)
	if err != nil {
		return []*emission.Validator{}, err
	}
	return resp.Validators, err
}
```

### 5. rpc/jsonrpc_server.go

We need to also define a corresponding function on our RPC server so whenever users interact with the API from their client, it talks to this server function which in turn calls the function defined in controller/resolutions.go. We can do this on rpc/jsonrpc_server.go:

```go
func (j *JSONRPCServer) UserStakeInfo(req *http.Request, args *StakeArgs, reply *StakeReply) (err error) {
	_, span := j.c.Tracer().Start(req.Context(), "Server.UserStakeInfo")
	defer span.End()

	userStake, err := j.c.GetUserStake(args.NodeID, args.Owner)
	if err != nil {
		return err
	}
	reply.UserStake = userStake
	return nil
}
```

## nuklai-cli

In order to easily test the capability of our new action and our new RPC API, we can integrate them in our nuklai-cli tool. This is not needed but highly encouraged because often times, external users will interact with our VM and giving developers the option to test their new actions via a command line tool is paramount. Think of `nuklai-cli` as a third party application that lets developers quickly interact with different `nuklaivm` features such as the new action we defined above or the `GetUserStake` function we defined above.

### 1. cmd/nuklai-cli/cmd/action.go

Let's define a new command to let users unstake their NAI tokens from the validator they have staked to in the past.

```go
var unstakeValidatorCmd = &cobra.Command{
	Use: "unstake-validator",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := bcli.Validators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s NodePublicKey=%s\n",
				i,
				validators[i].NodeID,
				validators[i].NodePublicKey,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to unstake from", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID, err := ids.NodeIDFromString(validatorChosen.NodeID)
		if err != nil {
			return err
		}

		// Get stake info
		owner, err := codec.AddressBech32(consts.HRP, priv.Address)
		if err != nil {
			return err
		}
		stake, err := bcli.UserStakeInfo(ctx, validatorChosen.NodeID, owner)
		if err != nil {
			return err
		}

		// Get current height
		_, currentHeight, _, err := cli.Accepted(ctx)
		if err != nil {
			return err
		}
		// Make sure to iterate over the stake info map in the same order every time
		keys := make([]string, 0, len(stake.StakeInfo))
		for k := range stake.StakeInfo {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// Show stake info to the user
		utils.Outf("{{cyan}}stake info:{{/}}\n")
		for index, txID := range keys {
			stakeInfo := stake.StakeInfo[txID]
			utils.Outf(
				"{{yellow}}%d:{{/}} TxID=%s StakedAmount=%d StartLockUpHeight=%d CurrentHeight=%d\n",
				index,
				txID,
				stakeInfo.Amount,
				stakeInfo.StartLockUp,
				currentHeight,
			)
		}

		// Select the stake Id to unstake
		stakeIndex, err := handler.Root().PromptChoice("stake to unstake", len(stake.StakeInfo))
		if err != nil {
			return err
		}
		stakeChosen := stake.StakeInfo[keys[stakeIndex]]
		stakeID, err := ids.FromString(stakeChosen.TxID)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.UnstakeValidator{
			Stake:  stakeID,
			NodeID: nodeID.Bytes(),
		}, cli, bcli, ws, factory, true)
		return err
	},
}
```

### 2. cmd/nuklai-cli/cmd/emission.go

Let's now define a new command to let users easily check their current stake on a chosen validator. We can do this on cmd/nuklai-cli/cmd/emission.go:

```go
var emissionStakeCmd = &cobra.Command{
	Use: "user-stake-info",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get clients
		clients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := clients[0].Validators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s NodePublicKey=%s\n",
				i,
				validators[i].NodeID,
				validators[i].NodePublicKey,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("choose validator whom you have staked to", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]

		// Get the address to look up
		stakeOwner, err := handler.Root().PromptAddress("address to get staking info for")
		if err != nil {
			return err
		}

		// Get user stake info
		_, err = handler.GetUserStake(ctx, clients[0], validatorChosen.NodeID, stakeOwner)
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
		unstakeValidatorCmd,
	)
...
emissionCmd.AddCommand(
		...
		emissionStakeCmd,
	)
```

### 4. cmd/nuklai-cli/cmd/handler.go

There is nothing left to do for our unstake validator action however, for the RPC API to get user stake, we need to define this function on handler.go so any other functions can call this function if need be. This is not needed but this is good practice so that multiple functions can reuse the same function.

```go
func (*Handler) GetUserStake(ctx context.Context,
	cli *brpc.JSONRPCClient, nodeID string, owner codec.Address,
) (*emission.UserStake, error) {
	saddr, err := codec.AddressBech32(consts.HRP, owner)
	if err != nil {
		return nil, err
	}
	userStake, err := cli.UserStakeInfo(ctx, nodeID, saddr)
	if err != nil {
		return nil, err
	}

	if userStake.Owner == "" {
		utils.Outf("{{yellow}}user stake: {{/}} Not staked yet\n")
	} else {
		utils.Outf(
			"{{yellow}}user stake: {{/}} Owner=%s StakedAmount=%d\n",
			userStake.Owner,
			userStake.StakedAmount,
		)
	}

	index := 1
	for txID, stakeInfo := range userStake.StakeInfo {
		utils.Outf(
			"{{yellow}}stake #%d:{{/}} TxID=%s Amount=%d StartLockUp=%d\n",
			index,
			txID,
			stakeInfo.Amount,
			stakeInfo.StartLockUp,
		)
		index++
	}
	return userStake, err
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

Note the working directory for run.sh is `/data/github/tmp/nuklaivm` so you will need to create it first and give it appropriate permissions.

```bash
sudo mkdir -p /data/github/tmp/nuklaivm;
sudo chown 777 /data/github/tmp/nuklaivm
```

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
./build/nuklai-cli action stake-validator
```

If successful, the output should be:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
validators: 5
0: NodeID=NodeID-wGVvYo7jBvtmnfUTaay2vcL8j8GJyokb NodePublicKey=rnp1CCGFvbni4bjFRFCJo7b3SxdnBIJ8qzOPPbB6HPR8VK8hvHaO37lZGNLJs30S
1: NodeID=NodeID-NuCwBadeuYbzntFJgBAS8Ut9pwo52XrT6 NodePublicKey=syBPqN0eU9nCBJDyFOVynneq/nia8lM0apG/DpboYtc7CJdm0hXlKGNZF5fwyjWp
2: NodeID=NodeID-3yxghtfwRdYcG69FjoxZrwjUkSXJAGhY9 NodePublicKey=kvFhrcEVW5Ooann3NaqqE2nANL/XS86AnCUFgrdyBQa2z+xAlCFcwuPHPDnvHyZp
3: NodeID=NodeID-6dvn9WTA4i7qG2pT3GKUXP46xa2SVY7Po NodePublicKey=oXMYzibvB7gHaGVAKVEB5z0+IFEPcWb0TxjrIz26p3eVjmaHkmKK41S64HDg8paD
4: NodeID=NodeID-423bGHFH5exxQfuNiRFUqxDquWD9svj6E NodePublicKey=rC9RaeHAUs4mSMw4YoAKBQaWecfrkLHEgKSq/JnfU2EnTXHjYuZu94aDQSTh1M7b
validator to stake to: 3
balance: 852999899.999972820 NAI
✔ Staked amount: 100█
End LockUp Height: 70
✔ continue (y/n): y█
✅ txID: EYnnHR9jtJvfAAE9UE5tkV9BDvajBhZ25YqJUD82LrRsJrZTo
```

### 5. Get user staking info

We can retrieve our staking info by passing in which validator we have staked to and the address to look up staking for using the new RPC API we defined as part of this exercise.

```bash
./build/nuklai-cli emission user-stake-info
```

If successful, the output should be:

```
database: .nuklai-cli
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
validators: 5
0: NodeID=NodeID-NuCwBadeuYbzntFJgBAS8Ut9pwo52XrT6 NodePublicKey=syBPqN0eU9nCBJDyFOVynneq/nia8lM0apG/DpboYtc7CJdm0hXlKGNZF5fwyjWp
1: NodeID=NodeID-3yxghtfwRdYcG69FjoxZrwjUkSXJAGhY9 NodePublicKey=kvFhrcEVW5Ooann3NaqqE2nANL/XS86AnCUFgrdyBQa2z+xAlCFcwuPHPDnvHyZp
2: NodeID=NodeID-6dvn9WTA4i7qG2pT3GKUXP46xa2SVY7Po NodePublicKey=oXMYzibvB7gHaGVAKVEB5z0+IFEPcWb0TxjrIz26p3eVjmaHkmKK41S64HDg8paD
3: NodeID=NodeID-423bGHFH5exxQfuNiRFUqxDquWD9svj6E NodePublicKey=rC9RaeHAUs4mSMw4YoAKBQaWecfrkLHEgKSq/JnfU2EnTXHjYuZu94aDQSTh1M7b
4: NodeID=NodeID-wGVvYo7jBvtmnfUTaay2vcL8j8GJyokb NodePublicKey=rnp1CCGFvbni4bjFRFCJo7b3SxdnBIJ8qzOPPbB6HPR8VK8hvHaO37lZGNLJs30S
✔ choose validator whom you have staked to: 2█
address to get staking info for: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
user stake:  Owner=nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx StakedAmount=100000000000
stake #1: TxID=EYnnHR9jtJvfAAE9UE5tkV9BDvajBhZ25YqJUD82LrRsJrZTo Amount=100000000000 StartLockUp=53
```

### 6. Unstake from a validator

Let's first check the balance on our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
uri: http://127.0.0.1:41743/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
balance: 852999799.999945203 NAI
```

We can unstake specific stake from a chosen validator.

```bash
./build/nuklai-cli action unstake-validator
```

Which produces result:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
validators: 5
0: NodeID=NodeID-wGVvYo7jBvtmnfUTaay2vcL8j8GJyokb NodePublicKey=rnp1CCGFvbni4bjFRFCJo7b3SxdnBIJ8qzOPPbB6HPR8VK8hvHaO37lZGNLJs30S
1: NodeID=NodeID-NuCwBadeuYbzntFJgBAS8Ut9pwo52XrT6 NodePublicKey=syBPqN0eU9nCBJDyFOVynneq/nia8lM0apG/DpboYtc7CJdm0hXlKGNZF5fwyjWp
2: NodeID=NodeID-3yxghtfwRdYcG69FjoxZrwjUkSXJAGhY9 NodePublicKey=kvFhrcEVW5Ooann3NaqqE2nANL/XS86AnCUFgrdyBQa2z+xAlCFcwuPHPDnvHyZp
3: NodeID=NodeID-6dvn9WTA4i7qG2pT3GKUXP46xa2SVY7Po NodePublicKey=oXMYzibvB7gHaGVAKVEB5z0+IFEPcWb0TxjrIz26p3eVjmaHkmKK41S64HDg8paD
4: NodeID=NodeID-423bGHFH5exxQfuNiRFUqxDquWD9svj6E NodePublicKey=rC9RaeHAUs4mSMw4YoAKBQaWecfrkLHEgKSq/JnfU2EnTXHjYuZu94aDQSTh1M7b
validator to unstake from: 3
stake info:
0: TxID=EYnnHR9jtJvfAAE9UE5tkV9BDvajBhZ25YqJUD82LrRsJrZTo StakedAmount=100000000000 StartLockUpHeight=53 CurrentHeight=200
stake ID to unstake: 0 [auto-selected]
continue (y/n): y
✅ txID: 2eSkTRQa4KqHDXidoeoQ8XSsjeSbga3x5B52hetwhGw68enbHt
```

Now, if we check the balance again, we should have our 100 NAI back to our account:

```bash
./build/nuklai-cli key balance
```

Which should produce a result like:

```
database: .nuklai-cli
address: nuklai1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjss0gwx
chainID: GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
uri: http://127.0.0.1:41743/ext/bc/GgbXLiBzd8j98CkrcEfsf13sbCTfwonTVMuFKgVVu4GpDNwJF
balance: 852999899.999896407 NAI
```
