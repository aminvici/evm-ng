package rpc

import (
	"errors"
	"fmt"
	"reflect"
	"math/big"
	"github.com/DSiSc/craft/log"
	cutil "github.com/DSiSc/crypto-suite/util"
	"github.com/DSiSc/evm-NG/system/contract/Interaction"
	"github.com/DSiSc/evm-NG/system/contract/util"
	wutils "github.com/DSiSc/wallet/utils"
	wcmn "github.com/DSiSc/web3go/common"
	"github.com/DSiSc/craft/rlp"
	wtypes "github.com/DSiSc/wallet/core/types"
	sutil "github.com/DSiSc/statedb-NG/util"
	craft "github.com/DSiSc/craft/types"
	//ctypes "github.com/DSiSc/craft/types"
)

var RpcContractAddr = cutil.HexToAddress("0000000000000000000000000000000000011101")

// rpc routes
var routes = map[string]*RPCFunc{
	string(util.ExtractMethodHash(util.Hash([]byte("ForwardFunds(string,uint64,string,string)")))): NewRPCFunc(ForwardFunds),
	string(util.ExtractMethodHash(util.Hash([]byte("GetTxState(string,uint64,string,string)")))): NewRPCFunc(GetTxState),
	string(util.ExtractMethodHash(util.Hash([]byte("ReceiveFunds(address,uint64,string,uint64)")))): NewRPCFunc(ReceiveFunds),
}

// 0 means failed, 1 means success
func ForwardFunds(toAddr string, amount uint64, payload string, chainFlag string) (error, string, string, uint64) {
	from, _ := Interaction.GetPubliceAcccount()
	to := cutil.HexToAddress(toAddr)
	localHash, targetHash, err := Interaction.CallCrossRawTransactionReq(from, to, amount, payload, chainFlag)
	if err != nil {
		return err, "", "", 0
	}

	localBytes := cutil.HashToBytes(localHash)
	targetBytes := cutil.HashToBytes(targetHash)

	return err, wcmn.BytesToHex(localBytes), wcmn.BytesToHex(targetBytes), 1
}

// GetCross Tx state
func GetTxState(txHash string, amount uint64, tmp string, chainFlag string) (error, uint64){
	//call the broadcast the tx
	port := Interaction.CrossTargetChainPort(chainFlag)

	web, err := wutils.NewWeb3("127.0.0.1", string(port), false)
	if err != nil {
		return err, 0
	}

	hash := cutil.HexToHash(txHash)
	receipt, err := web.Eth.GetTransactionReceipt(wcmn.Hash(hash))
	if err != nil || receipt == nil {
		return err, 0
	}

	status := uint64(receipt.Status.Int64())
	return err, status
}

// Receipt funds
func ReceiveFunds(to string, amount uint64, payload string, srcChainId uint64) (error, uint64){
	//receive tx bytes, decode input
	input := wcmn.HexToBytes(payload)
	tx := new(craft.Transaction)
	if err := rlp.DecodeBytes(input, tx); err != nil {

		ethTx := new(craft.ETransaction)
		err = ethTx.DecodeBytes(input)
		if err != nil {
			log.Info("sendRawTransaction tx decode as ethereum error, err = %v", err)
			return err, 0
		}
		ethTx.SetTxData(&tx.Data)
	}

	from_, err := wtypes.Sender(wtypes.NewEIP155Signer(big.NewInt(int64(srcChainId))), tx)
	if err != nil {
		log.Error("get from address failed, err = %v", err)
		return err, 0
	}

	contractAddr := "0x47c5e40890bce4a473a49d7501808b9633f29782"
	targetToInput := wcmn.BytesToHex(tx.Data.Payload)
	txToAddr := sutil.AddressToHex(*tx.Data.Recipient)
	fromAddr := sutil.AddressToHex(craft.Address(from_))
	txFromAddr := sutil.AddressToHex(*tx.Data.From)
	//verify args
	if amount != tx.Data.Amount.Uint64() || contractAddr != txToAddr || fromAddr != txFromAddr || to != targetToInput{
		return errors.New("tx args not matched tx's"), 0
	}

	log.Info("ReceiveFunds_verify_success, targetToAddr=%s, amount=%d, srcChainId=%d", to, amount, srcChainId)
	return nil, 1
}

// Register register a rpc route
func Register(methodName string, f *RPCFunc) error {
	paramStr := ""
	for _, arg := range f.args {
		switch arg.Kind() {
		case reflect.Uint64:
			paramStr += "uint64,"
		case reflect.String:
			paramStr += "string,"
		case reflect.Slice:
			if reflect.Uint8 != arg.Elem().Kind() {
				return errors.New("unsupported arg type")
			}
		}
	}
	if len(paramStr) > 0 {
		paramStr = paramStr[:len(paramStr)-1]
	}
	methodHash := util.Hash([]byte(methodName + "(" + paramStr + ")"))[:4]
	routes[string(methodHash)] = f
	return nil
}

func Handler(input []byte) ([]byte, error) {
	method := util.ExtractMethodHash(input)
	rpcFunc := routes[string(method)]
	if rpcFunc == nil {
		return nil, errors.New("routes not found")
	}

	args, err := inputParamsToArgs(rpcFunc, input[len(method):])
	if err != nil {
		return nil, err
	}

	log.Info("contract RPC method: %s", wcmn.BytesToHex(method))
	returns := rpcFunc.f.Call(args)
	return encodeResult(returns)
}

// Covert an http query to a list of properly typed values.
// To be properly decoded the arg must be a concrete type from tendermint (if its an interface).
func inputParamsToArgs(rpcFunc *RPCFunc, input []byte) ([]reflect.Value, error) {
	args := make([]interface{}, 0)
	for _, argT := range rpcFunc.args {
		args = append(args, reflect.New(argT).Interface())
	}
	err := util.ExtractParam(input, args...)
	if err != nil {
		return nil, err
	}

	argVs := make([]reflect.Value, 0)
	for _, arg := range args {
		argVs = append(argVs, reflect.ValueOf(arg).Elem())
	}
	return argVs, nil
}

// NOTE: assume returns is result struct and error. If error is not nil, return it
func encodeResult(returns []reflect.Value) ([]byte, error) {
	errV := returns[0]
	if errV.Interface() != nil {
		return nil, errors.New(fmt.Sprintf("%v", errV.Interface()))
	}
	returns = returns[1:]
	rvs := make([]interface{}, 0)
	for _, rv := range returns {
		// the result is a registered interface,
		// we need a pointer to it so we can marshal with type byte
		rvp := reflect.New(rv.Type())
		rvp.Elem().Set(rv)
		rvs = append(rvs, rvp.Elem().Interface())
	}
	return util.EncodeReturnValue(rvs...)
}

// RPCFunc contains the introspected type information for a function
type RPCFunc struct {
	f       reflect.Value  // underlying rpc function
	args    []reflect.Type // type of each function arg
	returns []reflect.Type // type of each return arg
}

// NewRPCFunc create a new RPCFunc instance
func NewRPCFunc(f interface{}) *RPCFunc {
	return &RPCFunc{
		f:       reflect.ValueOf(f),
		args:    funcArgTypes(f),
		returns: funcReturnTypes(f),
	}
}

// return a function's argument types
func funcArgTypes(f interface{}) []reflect.Type {
	t := reflect.TypeOf(f)
	n := t.NumIn()
	typez := make([]reflect.Type, n)
	for i := 0; i < n; i++ {
		typez[i] = t.In(i)
	}
	return typez
}

// return a function's return types
func funcReturnTypes(f interface{}) []reflect.Type {
	t := reflect.TypeOf(f)
	n := t.NumOut()
	typez := make([]reflect.Type, n)
	for i := 0; i < n; i++ {
		typez[i] = t.Out(i)
	}
	return typez
}

