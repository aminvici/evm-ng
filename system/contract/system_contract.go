package contract

import (
	"github.com/DSiSc/craft/types"
)

// SystemContract represent system internal contract
type SystemContract interface {
	// Address return contract address
	Address() types.Address
}
