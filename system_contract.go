package evm

import (
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/evm-NG/system/contract/buffer"
	"github.com/DSiSc/evm-NG/system/contract/rpc"
	"github.com/DSiSc/evm-NG/system/contract/storage"
)

// SysContractExecutionFunc system contract execute function
type SysContractExecutionFunc func(interpreter *EVM, contract ContractRef, input []byte) ([]byte, error)

// system call routes
var routes = make(map[types.Address]SysContractExecutionFunc)

func init() {
	routes[buffer.SystemBufferAddr] = func(execEvm *EVM, contract ContractRef, input []byte) ([]byte, error) {
		systemBuffer := buffer.NewSystemBufferContract(execEvm.StateDB)
		return buffer.BufferExecute(systemBuffer, input)
	}

	routes[storage.TencentCosAddr] = func(execEvm *EVM, caller ContractRef, input []byte) ([]byte, error) {
		systemBuffer := buffer.NewSystemBufferContract(execEvm.StateDB)
		systemBufferReadWriter := buffer.NewSystemBufferReadWriterCloser(systemBuffer)
		tencentCos := storage.NewTencentCosContract(systemBufferReadWriter)
		return storage.CosExecute(tencentCos, input)
	}

	routes[rpc.RpcContractAddr] = func(execEvm *EVM, caller ContractRef, input []byte) ([]byte, error) {
		return rpc.Handler(input)
	}
}

//IsSystemContract check the contract with specified address is system contract
func IsSystemContract(addr types.Address) bool {
	return routes[addr] != nil
}

// GetSystemContractExecFunc get system contract execution function by address
func GetSystemContractExecFunc(addr types.Address) SysContractExecutionFunc {
	return routes[addr]
}
