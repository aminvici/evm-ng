package rpc

import (
	"fmt"
	"github.com/DSiSc/evm-NG/common"
	"github.com/DSiSc/evm-NG/common/hexutil"
	"github.com/DSiSc/evm-NG/system/contract/util"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var permDenyError = errors.New("permission deny")

func init() {
	routes[string(util.ExtractMethodHash(util.Hash([]byte("test1()"))))] = NewRPCFunc(func() error {
		return nil
	})

	routes[string(util.ExtractMethodHash(util.Hash([]byte("test2(string)"))))] = NewRPCFunc(func(name string) error {
		return nil
	})

	routes[string(util.ExtractMethodHash(util.Hash([]byte("test3(string,uint64)"))))] = NewRPCFunc(func(name string) error {
		return nil
	})

	routes[string(util.ExtractMethodHash(util.Hash([]byte("test4(string,uint64)"))))] = NewRPCFunc(func(name string, age uint64) (error, string, uint64) {
		return nil, "Hello " + name, age
	})

	routes[string(util.ExtractMethodHash(util.Hash([]byte("test5(string,uint64)"))))] = NewRPCFunc(func(name string, age uint64) (error, string) {
		return permDenyError, ""
	})
}

func Method1(name string, age uint64) {

}

func TestRegister(t *testing.T) {
	assert := assert.New(t)
	err := Register("Method1", NewRPCFunc(Method1))
	assert.Nil(err)
	methodHash := util.Hash([]byte("Method1(string,uint64)"))[:4]
	assert.NotNil(routes[string(methodHash)])

	optionFunc := util.ExtractMethodHash(util.Hash([]byte("ReceiveFunds(string, uint64)")))
	fmt.Println(common.Bytes2Hex(optionFunc))
}

func TestNewRPCFunc(t *testing.T) {
	assert := assert.New(t)
	f := func(name string, age uint64) (error, string) {
		return nil, ""
	}
	rpcFunc := NewRPCFunc(f)
	assert.Equal(2, len(rpcFunc.args))
	assert.Equal(reflect.String, rpcFunc.args[0].Kind())
	assert.Equal(reflect.Uint64, rpcFunc.args[1].Kind())
}

func TestHandler(t *testing.T) {
	assert := assert.New(t)
	input, _ := hexutil.Decode("0x6b59084d")
	_, err := Handler(input)
	assert.Nil(err)
}

func TestHandler1(t *testing.T) {
	assert := assert.New(t)
	input, _ := hexutil.Decode("0x30e738a700000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000003546f6d0000000000000000000000000000000000000000000000000000000000")
	_, err := Handler(input)
	assert.Nil(err)
}

func TestHandler2(t *testing.T) {
	assert := assert.New(t)
	input, _ := hexutil.Decode("0xfb61de2c0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000007b0000000000000000000000000000000000000000000000000000000000000003546f6d0000000000000000000000000000000000000000000000000000000000")
	_, err := Handler(input)
	assert.Nil(err)
}

func TestHandler3(t *testing.T) {
	assert := assert.New(t)
	expectRet, _ := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000007b000000000000000000000000000000000000000000000000000000000000000948656c6c6f20546f6d0000000000000000000000000000000000000000000000")
	input, _ := hexutil.Decode("0x4a1607800000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000007b0000000000000000000000000000000000000000000000000000000000000003546f6d0000000000000000000000000000000000000000000000000000000000")
	ret, err := Handler(input)
	assert.Nil(err)
	assert.Equal(expectRet, ret)
}

func TestHandler4(t *testing.T) {
	assert := assert.New(t)
	expectRet, _ := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000007b000000000000000000000000000000000000000000000000000000000000000948656c6c6f20546f6d0000000000000000000000000000000000000000000000")
	input, _ := hexutil.Decode("0x4a1607800000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000007b0000000000000000000000000000000000000000000000000000000000000003546f6d0000000000000000000000000000000000000000000000000000000000")
	ret, err := Handler(input)
	assert.Nil(err)
	assert.Equal(expectRet, ret)
}

func TestHandler5(t *testing.T) {
	assert := assert.New(t)
	input, _ := hexutil.Decode("0xd85a9b800000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000007b0000000000000000000000000000000000000000000000000000000000000003546f6d0000000000000000000000000000000000000000000000000000000000")
	_, err := Handler(input)
	assert.NotNil(err)
	assert.Equal(fmt.Sprintf("%v", permDenyError), fmt.Sprintf("%v", err))
}
