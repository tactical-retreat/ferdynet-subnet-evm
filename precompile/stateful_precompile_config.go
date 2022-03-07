// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package precompile

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ava-labs/subnet-evm/utils"
)

// StatefulPrecompileConfig defines the interface for a stateful precompile to
type StatefulPrecompileConfig interface {
	// Address returns the address where the stateful precompile is accessible.
	Address() common.Address
	// Timestamp returns the timestamp at which this stateful precompile should be enabled.
	// 1) 0 indicates that the precompile should be enabled from genesis.
	// 2) n indicates that the precompile should be enabled in the first block with timestamp >= [n].
	// 3) nil indicates that the precompile is never enabled.
	Timestamp() *big.Int
	// Configure is called on the first block where the stateful precompile should be enabled.
	// This allows the stateful precompile to configure its own state via [StateDB] as necessary.
	// This function must be deterministic since it will impact the EVM state. If a change to the
	// config causes a change to the state modifications made in Configure, then it cannot be safely
	// made to the config after the network upgrade has gone into effect.
	//
	// Configure is called on the first block where the stateful precompile should be enabled. This
	// provides the config the ability to set its initial state and should only modify the state within
	// its own address space.
	Configure(StateDB)
	// Contract returns a thread-safe singleton that can be used as the StatefulPrecompiledContract when
	// this config is enabled.
	Contract() StatefulPrecompiledContract
}

// CheckConfigure checks if [config] is activated by the transition from block at [parentTimestamp] to [currentTimestamp].
// If it does, then it calls Configure on [config] to make the necessary state update to enable the StatefulPrecompile.
// Note: this function is called within genesis to configure the starting state if it [config] specifies that it should be
// configured at genesis, or happens during block processing to update the state before processing the given block.
// TODO: add ability to call Configure at different timestamps, so that developers can easily re-configure by updating the
// stateful precompile config.
// Assumes that [config] is non-nil.
func CheckConfigure(parentTimestamp *big.Int, currentTimestamp *big.Int, config StatefulPrecompileConfig, state StateDB) {
	forkTimestamp := config.Timestamp()
	// If the network upgrade goes into effect within this transition, configure the stateful precompile
	if utils.IsForkTransition(forkTimestamp, parentTimestamp, currentTimestamp) {
		// Set the nonce of the precompile's address (as is done when a contract is created) to ensure
		// that it is marked as non-empty and will not be cleaned up when the statedb is finalized.
		state.SetNonce(config.Address(), 1)
		config.Configure(state)
	}
}