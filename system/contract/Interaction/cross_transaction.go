package Interaction

import (
	"fmt"
	"math/big"
	"errors"
	atypes "github.com/DSiSc/apigateway/core/types"
	"github.com/DSiSc/craft/monitor"
	"github.com/DSiSc/craft/rlp"
	"github.com/DSiSc/craft/types"
	//"github.com/DSiSc/crypto-suite/crypto"
	cutil "github.com/DSiSc/crypto-suite/util"
	"github.com/DSiSc/evm-NG/system/contract/util"
	//sutil "github.com/DSiSc/statedb-NG/util"
	//"github.com/DSiSc/txpool"
	wtypes "github.com/DSiSc/wallet/core/types"
	wutils "github.com/DSiSc/wallet/utils"
	wcmn "github.com/DSiSc/web3go/common"
	craft "github.com/DSiSc/craft/types"
	sutil "github.com/DSiSc/statedb-NG/util"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/crypto-suite/crypto"
	eutil "github.com/DSiSc/evm-NG/system/contract/util"
)

const(
	FAILED = iota
	SUCCESS
	PENDING
)

type CrossChainPort string
// define specified type of system contract
const (
	Null = "Null"
	JustitiaChainA = "chainA"
	JustitiaChainB = "chainB"
)
const (
	InitialCrossChainPort CrossChainPort = "0"
	ChainACrossChainPort = "47768"
	ChainBCrossChainPort = "47769"
)

type Status uint64

var CrossChainAddr = cutil.HexToAddress("0000000000000000000000000000000000011100")
var (
	forwardFundsMethodHash = string(util.ExtractMethodHash(util.Hash([]byte("forwardFunds(string, uint64, string)"))))
	getTxStateMethodHash = string(util.ExtractMethodHash(util.Hash([]byte("getTxState(string,string)"))))
)

type CrossChainContract struct {
	records map[types.Address]Status
}

func NewCrossChainContract() *CrossChainContract {
	return new(CrossChainContract)
}

func CrossTargetChainPort(chainFlag string) CrossChainPort {
	var crossPort = InitialCrossChainPort
	if chainFlag == JustitiaChainA {
		crossPort = ChainACrossChainPort
	} else if chainFlag == JustitiaChainB {
		crossPort = ChainBCrossChainPort
	}
	return crossPort
}

func OppositeChainPort(chainFlag string) CrossChainPort {
	var crossPort = InitialCrossChainPort
	if chainFlag == JustitiaChainA {
		crossPort = ChainBCrossChainPort
	} else if chainFlag == JustitiaChainB {
		crossPort = ChainACrossChainPort
	}
	return crossPort
}

//如何获得合约的调用者？？？，保证资金安全性
func (this *CrossChainContract) forwardFunds(toAddr types.Address, amount uint64, payload string, chainFlag string) (types.Hash, bool) {
	//调用apigateway的receiveCrossTx交易
	from, err := GetPubliceAcccount()
	if err != nil {
		return types.Hash{}, false
	}

	hash, _, err := CallCrossRawTransactionReq(from, toAddr, amount, payload, chainFlag)
	if err != nil {
		return types.Hash{}, false
	}
	return hash, false
}

func (this *CrossChainContract) getTxState(address types.Address, chainFlag string) (uint64, bool) {
	switch chainFlag {
		case "chainA":
			//123
			fmt.Println()
		default:
	}

	return SUCCESS, true
}

func CallCrossRawTransactionReq(from types.Address, to types.Address, amount uint64, payload string, chainFlag string) (types.Hash, types.Hash, error) {
	monitor.JTMetrics.ApigatewayReceivedTx.Add(1)

	//call the broadcast the tx
	//port := CrossTargetChainPort(chainFlag)

	local := OppositeChainPort(chainFlag)
	web, err := wutils.NewWeb3("127.0.0.1", string(local), false)
	//web, err := wutils.NewWeb3("127.0.0.1", string(port), false)
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}

	// receive tx bytes, decode input
	input := wcmn.HexToBytes(payload)
	tx := new(craft.Transaction)
	if err := rlp.DecodeBytes(input, tx); err != nil {

		ethTx := new(craft.ETransaction)
		err = ethTx.DecodeBytes(input)
		if err != nil {
			log.Info("sendRawTransaction tx decode as ethereum error, err = %v", err)
			return types.Hash{}, types.Hash{}, err
		}
		ethTx.SetTxData(&tx.Data)
	}

	contractAddr := "0x47c5e40890bce4a473a49d7501808b9633f29782"
	txToAddr := sutil.AddressToHex(*tx.Data.Recipient)
	//TODO: verify args;add verify contract call addr is equal to payload from, or signer
	if amount != tx.Data.Amount.Uint64() || contractAddr != txToAddr {
		return types.Hash{}, types.Hash{}, errors.New("tx args not matched tx's")
	}

	//sendRawTransaction
	localHash, err := web.Eth.SendRawTransaction(input)
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}

	//switch payload to target chian's contract
	//TODO: construct a tx, which contract call
	argTo := to
	argPayload := payload
	argAmount := amount
	argChainId := uint64(5777)

	payload_, err := eutil.EncodeReturnValue(argTo, argPayload, argAmount, argChainId)
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}
	//funcSelector := wcmn.BytesToHex(util.ExtractMethodHash(util.Hash([]byte("ReceiveFunds(address,uint64,string,uint64)"))))
	funcSelector := "0xd0fe3c8b"
	//funcSelector := "0x5e678a44"
	payload_1 := wcmn.BytesToHex(payload_)

	input__ := funcSelector + payload_1[2:]
	input_ := wcmn.HexToBytes(input__)
	target := CrossTargetChainPort(chainFlag)
	//target := "47768"
	web_, err := wutils.NewWeb3("127.0.0.1", string(target), false)
	bigNonce , err := web_.Eth.GetTransactionCount(wcmn.Address(from), "latest")
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}

	tx_ := new(types.Transaction)
	addr, err := GetPubliceAcccount()
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}
	tx_.Data.From = &addr
	to_ := cutil.HexToAddress("47c5e40890bce4a473a49d7501808b9633f29782")
	private := "29ad43a4ebb4a65436d9fb116d471d96516b3d5cc153e045b384664bed5371b9"
	nonce := bigNonce.Uint64()
	tx_.Data.AccountNonce = nonce
	tx_.Data.Price = big.NewInt(0)
	tx_.Data.GasLimit = 6721975
	tx_.Data.Recipient = &to_
	tx_.Data.Amount = big.NewInt(int64(0))
	//payload填充问题，填充合约调用的参数
	tx_.Data.Payload = input_

	//sign tx
	priKey, err := crypto.HexToECDSA(private)
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}

	chainID := big.NewInt(int64(5777))
	tx_, err = wtypes.SignTx(tx_, wtypes.NewEIP155Signer(chainID), priKey)
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}

	txBytes, _ := rlp.EncodeToBytes(tx_)
	targetHash, err := web_.Eth.SendRawTransaction(txBytes)
	if err != nil {
		return types.Hash{}, types.Hash{}, err
	}

	log.Info("cross funds tx localHash, %s", sutil.HashToHex(craft.Hash(localHash)))
	log.Info("cross funds tx targetHash, %s", sutil.HashToHex(craft.Hash(targetHash)))
	return types.Hash(localHash), types.Hash(targetHash), nil
}

func GetPubliceAcccount() (types.Address, error){
	//get from config or genesis ?
	addr := "0x0fA3E9c7065Cf9b5f513Fb878284f902d167870c"
	address := atypes.HexToAddress(addr)

	return types.Address(address), nil
}

