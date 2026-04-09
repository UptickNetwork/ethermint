package types

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
)

// Options for the EVM module
type Options struct {
	CanTransfer vm.CanTransferFunc
	Transfer    vm.TransferFunc
}

// DefaultOptions for the EVM module
func DefaultOptions() Options {
	return Options{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
	}
}

// Message wraps core.Message and carries fee payer metadata.
type Message struct {
	core.Message
	FeePayer string
}
