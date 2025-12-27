package storage

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/DSiSc/craft/types"
	cutil "github.com/DSiSc/crypto-suite/util"
	"github.com/DSiSc/evm-NG/constant"
	"github.com/DSiSc/evm-NG/system/contract/buffer"
	"github.com/DSiSc/evm-NG/system/contract/util"
	"github.com/pkg/errors"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
)

var TencentCosAddr = cutil.HexToAddress("0000000000000000000000000000000000011110")

var (
	getObjectMethodHash = string(util.ExtractMethodHash(util.Hash([]byte("GetObject(string,string)"))))
	putObjectMethodHash = string(util.ExtractMethodHash(util.Hash([]byte("PutObject(string,string)"))))
)

// execute the cos contract
func CosExecute(cos *TencentCosContract, input []byte) ([]byte, error) {
	methodHash := util.ExtractMethodHash(input)
	switch string(methodHash) {
	case getObjectMethodHash:
		rawUrl := new(string)
		objName := new(string)
		err := util.ExtractParam(input[len(methodHash):], rawUrl, objName)
		if err != nil {
			return nil, err
		}
		bufferAddr, err := cos.GetObject(*rawUrl, *objName)
		if err != nil {
			return nil, err
		}
		return util.EncodeReturnValue(bufferAddr)
	case putObjectMethodHash:
		rawUrl := new(string)
		objName := new(string)
		err := util.ExtractParam(input[len(methodHash):], rawUrl, objName)
		if err != nil {
			return nil, err
		}
		objMeta, err := cos.PutObject(*rawUrl, *objName)
		if err != nil {
			return nil, err
		}
		ret, err := util.EncodeReturnValue(objMeta.ETag, objMeta.VersionId, objMeta.EncryptionAlg)
		if err != nil {
			return nil, err
		}
		return ret, nil
	default:
		return nil, errors.New("unknown method")
	}
}

//Cos response error
type RespError struct {
	Code      string `xml:"Code"`
	Message   string `xml:"Message"`
	Resource  string `xml:"Resource"`
	RequestId string `xml:"RequestId"`
	TraceId   string `xml:"TraceId"`
}

// ObjectMeta object meta info
type ObjectMeta struct {
	ETag          string
	VersionId     string
	EncryptionAlg string
}

// TencentCosContract `tencent cloud object storage` system contract
type TencentCosContract struct {
	sysBufferRW *buffer.SystemBufferReadWriterCloser
}

// create a new instance
func NewTencentCosContract(rw *buffer.SystemBufferReadWriterCloser) *TencentCosContract {
	return &TencentCosContract{
		sysBufferRW: rw,
	}
}

// GetObject download an object from the cloud server to `sysBufferRW`
func (this *TencentCosContract) GetObject(rawurl, name string) (types.Address, error) {
	client, err := buildCosClient(rawurl)
	if err != nil {
		return types.Address{}, err
	}

	resp, err := client.Object.Get(context.Background(), name, nil)
	if err != nil {
		return types.Address{}, err
	}
	defer resp.Body.Close()
	if err = checkResponse(resp); err != nil {
		return types.Address{}, err
	}

	bufferBytes := make([]byte, constant.BufferMaxReadWriteSize)
	for {
		nr, err := resp.Body.Read(bufferBytes)
		if nr > 0 {
			nw, err := this.sysBufferRW.Write(bufferBytes[:nr])
			if err != nil || nw < nr {
				return types.Address{}, err
			}
		}
		if err == io.EOF {
			return this.sysBufferRW.ContractAddress(), nil
		}
		if err != nil {
			return types.Address{}, err
		}
	}
}

// PutObject upload an object(stored in `sysBufferRW`) to cloud server
func (this *TencentCosContract) PutObject(rawurl, name string) (*ObjectMeta, error) {
	client, err := buildCosClient(rawurl)
	if err != nil {
		return nil, err
	}

	resp, err := client.Object.Put(context.Background(), name, this.sysBufferRW, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = checkResponse(resp); err != nil {
		return nil, err
	}

	objMeta := getObjectMeta(resp.Header)
	return objMeta, err
}

func (this *TencentCosContract) Address() types.Address {
	return TencentCosAddr
}

// build cos client with specified url
func buildCosClient(rawurl string) (*cos.Client, error) {
	u, e := url.Parse(rawurl)
	if e != nil {
		return nil, e
	}
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{})
	return c, nil
}

// extract object meta info from response header
func getObjectMeta(header http.Header) *ObjectMeta {
	return &ObjectMeta{
		ETag:          header.Get("ETag"),
		VersionId:     header.Get("x-cos-version-id"),
		EncryptionAlg: header.Get("x-cos-server-side-encryption"),
	}
}

// check response status
func checkResponse(resp *cos.Response) (err error) {
	if 200 != resp.StatusCode {
		var respError RespError
		if err1 := parseResp(resp.Body, xmlType, &respError); err1 != nil {
			return err1
		} else {
			return errors.Errorf("response error, Code: %s, Message: %s, Resource: %s, RequestId: %s, TraceId: %s", respError.Code, respError.Message, respError.Resource, respError.RequestId, respError.TraceId)
		}
	} else {
		return nil
	}
}

const (
	xmlType   = "xml"
	jsonType  = "json"
	bytesType = "bytes"
)

// parse response
func parseResp(resp io.Reader, parseType string, v interface{}) error {
	respBytes, err := ioutil.ReadAll(resp)
	if err != nil {
		return err
	}
	switch parseType {
	case xmlType:
		return xml.Unmarshal(respBytes, v)
	case jsonType:
		return json.Unmarshal(respBytes, v)
	case bytesType:
		srcV, dstV := reflect.ValueOf(v), reflect.ValueOf(&respBytes)
		srcV.Elem().Set(dstV.Elem())
		return nil
	default:
		return errors.New("unknown parse type")
	}
}
