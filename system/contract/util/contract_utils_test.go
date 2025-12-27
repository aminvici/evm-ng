package util

import (
	"fmt"
	"github.com/DSiSc/crypto-suite/util"
	"github.com/DSiSc/evm-NG/common/hexutil"
	"github.com/stretchr/testify/assert"
	"testing"
	util2 "github.com/DSiSc/statedb-NG/util"
	"github.com/DSiSc/web3go/common"
)

func TestHash(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello, World")
	exepectedHash := []byte{0xa0, 0x4a, 0x45, 0x10, 0x28, 0xd0, 0xf9, 0x28, 0x4c, 0xe8, 0x22, 0x43, 0x75, 0x5e, 0x24, 0x52, 0x38, 0xab, 0x1e, 0x4e, 0xcf, 0x7b, 0x9d, 0xd8, 0xbf, 0x47, 0x34, 0xd9, 0xec, 0xfd, 0x5, 0x29}
	actualHash := Hash(data)
	assert.Equal(exepectedHash, actualHash)
}

func TestExtractMethodHash(t *testing.T) {
	assert := assert.New(t)
	expectedHash := Hash([]byte("hello(string,string)"))[:4]
	input, _ := hexutil.Decode("0x939531c0000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")
	methodHash := ExtractMethodHash(input)
	assert.Equal(expectedHash, methodHash)
}

func TestExtractParam(t *testing.T) {
	assert := assert.New(t)
	arg1 := new(string)
	arg2 := new(string)
	expectedParam1 := "a"
	expectedParam2 := "b"
	input, _ := hexutil.Decode("0x939531c0000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")
	err := ExtractParam(input[4:], arg1, arg2)
	assert.Nil(err)
	assert.Equal(expectedParam1, *arg1)
	assert.Equal(expectedParam2, *arg2)
}

func TestExtractParam2(t *testing.T) {
	assert := assert.New(t)
	arg1 := make([]byte, 0)
	arg2 := make([]byte, 0)
	expectedParam1 := []byte("a")
	expectedParam2 := []byte("b")
	input, _ := hexutil.Decode("0x939531c0000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")
	err := ExtractParam(input[4:], &arg1, &arg2)
	assert.Nil(err)
	assert.Equal(expectedParam1, arg1)
	assert.Equal(expectedParam2, arg2)
}

func TestExtractParam3(t *testing.T) {
	assert := assert.New(t)
	arg1 := uint64(2)
	expectedParam1 := uint64(2)
	input, _ := hexutil.Decode("0xe05e91e00000000000000000000000000000000000000000000000000000000000000002")
	err := ExtractParam(input[4:], &arg1)
	assert.Nil(err)
	assert.Equal(expectedParam1, arg1)
}

func TestEncodeReturnValue(t *testing.T) {

	bytes1 := ExtractMethodHash(Hash([]byte("ForwardFunds(string,uint64,string,string)")))
	bytes2 := ExtractMethodHash(Hash([]byte("ReceiveFunds(address,uint64,string,uint64)")))
	bytes3 := ExtractMethodHash(Hash([]byte("GetTxState(string,uint64,string)")))

	fmt.Println(common.BytesToHex(bytes1))
	fmt.Println(common.BytesToHex(bytes2))
	fmt.Println(common.BytesToHex(bytes3))

	assert := assert.New(t)
	retVal1 := "a"
	retVal2 := "b"
	expect, _ := hexutil.Decode("0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")
	retB, err := EncodeReturnValue(retVal1, retVal2)
	//crossTx
	addr := util2.HexToAddress("0x932563314c6d88f23a4d7B026C14A3BD17144Be6")
	//0x47c5e40890bce4a473a49d7501808b9633f29782
	payload := "0xf87bf8760180809447c5e40890bce4a473a49d7501808b9633f29782949f026b8fec907c3747ecd8f167e41e724def98b18203e880822d45a08f16ef40c7b85927facc7eb41bc6d68ca39168f7df56808a9f880b46d742406ea02ea1239604fdf4dccb2fc0db88ef7889e736a5ea09b417bdba701ae3a966718ec0c0c0"

	//0x6295ee1b4f6dd65047762f924ecd367c17eabf8f
	//payload := "0xf87bf876808080946295ee1b4f6dd65047762f924ecd367c17eabf8f949f026b8fec907c3747ecd8f167e41e724def98b18203e880822d45a097bc0e0c52cb3799b41fb1cfe037d8ec8a4bf27800bd93c6b120b3651417fbdaa05109bc716f2a3090c6406e430d22cdfa64f159cfea15e066d1390a134d602496c0c0c0"
	chainFlag := "chainA"
	retA, err := EncodeReturnValue(addr, payload, chainFlag)
	str := common.BytesToHex(retA)
	fmt.Println("str = 0x68d4a18e", str)

	//query
	addr = util2.HexToAddress("0x9f026b8fec907c3747ecd8f167e41e724def98b1")
	chainFlag = "chainA"
	retC, err := EncodeReturnValue(addr, chainFlag)
	str = common.BytesToHex(retC)
	fmt.Println("str = 0x15508866", str)

	assert.Nil(err)
	assert.Equal(expect, retB)
}

func TestEncodeReturnValue2(t *testing.T) {
	assert := assert.New(t)
	retVal1 := []byte("a")
	retVal2 := []byte("b")
	expect, _ := hexutil.Decode("0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")
	retB, err := EncodeReturnValue(retVal1, retVal2)
	assert.Nil(err)
	assert.Equal(expect, retB)
}

func TestEncodeReturnValue3(t *testing.T) {
	assert := assert.New(t)
	retVal1 := uint64(2)
	expect, _ := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000002")
	retB, err := EncodeReturnValue(retVal1)
	assert.Nil(err)
	assert.Equal(expect, retB)
}

func TestEncodeReturnValue4(t *testing.T) {
	assert := assert.New(t)
	addr := util.HexToAddress("0000000000000000000000000000000000011110")
	expect, _ := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000011110")
	retB, err := EncodeReturnValue(addr)
	assert.Nil(err)
	assert.Equal(expect, retB)
}
