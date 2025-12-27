package buffer

import (
	"encoding/binary"
	"errors"
	"github.com/DSiSc/craft/types"
	cutil "github.com/DSiSc/crypto-suite/util"
	"github.com/DSiSc/evm-NG/common/math"
	"github.com/DSiSc/evm-NG/system/contract/util"
	"github.com/DSiSc/repository"
	"math/big"
)

var SystemBufferAddr = cutil.HexToAddress("0000000000000000000000000000000000011111")

const (
	systemBufferCacheKey = "SystemBufferCacheKey"
	truncSize            = 256
)

var (
	systemBufferCacheStart = util.Hash([]byte(systemBufferCacheKey))
	readMethodHash         = string(util.ExtractMethodHash(util.Hash([]byte("Read(uint64,uint64)"))))
	writeMethodHash        = string(util.ExtractMethodHash(util.Hash([]byte("Write(bytes)"))))
	lengthMethodHash       = string(util.ExtractMethodHash(util.Hash([]byte("Length()"))))
	closeMethodHash        = string(util.ExtractMethodHash(util.Hash([]byte("Close()"))))
)

// execute the system buffer contract
func BufferExecute(sysBuffer *SystemBufferContract, input []byte) ([]byte, error) {
	methodHash := util.ExtractMethodHash(input)
	switch string(methodHash) {
	case readMethodHash:
		var offset, size uint64
		err := util.ExtractParam(input[len(methodHash):], &offset, &size)
		if err != nil {
			return nil, err
		}
		data, err := sysBuffer.Read(offset, size)
		if err != nil {
			return nil, err
		}
		retData, err := util.EncodeReturnValue(data)
		if err != nil {
			return nil, err
		}
		return retData, nil
	case writeMethodHash:
		data := make([]byte, 0)
		err := util.ExtractParam(input[len(methodHash):], &data)
		if err != nil {
			return nil, err
		}
		size, err := sysBuffer.Write(data)
		if err != nil {
			return nil, err
		}
		retData, err := util.EncodeReturnValue(size)
		if err != nil {
			return nil, err
		}
		return retData, nil
	case lengthMethodHash:
		len := sysBuffer.Length()
		return util.EncodeReturnValue(len)
	case closeMethodHash:
		err := sysBuffer.Close()
		return nil, err
	default:
		return nil, errors.New("unknown method")
	}
}

// SystemBufferContract used to cache the system contract data
type SystemBufferContract struct {
	db *repository.Repository
}

// NewSystemBufferContract create a SystemBufferContract instance.
func NewSystemBufferContract(db *repository.Repository) *SystemBufferContract {
	return &SystemBufferContract{
		db: db,
	}
}

// Read read the data recorded in buffer
// offset: length of the bytes to be skipped
// size: max length to read
func (this *SystemBufferContract) Read(offset, size uint64) ([]byte, error) {
	data := make([]byte, 0)
	val, err := this.db.Get([]byte(systemBufferCacheKey))
	if err != nil {
		return nil, err
	}
	if val == nil || offset+size > binary.BigEndian.Uint64(val) {
		return nil, errors.New("invalid read position")
	}
	preLen := int(offset % truncSize)
	for start := offset / truncSize; uint64(len(data)) < size+uint64(preLen); start++ {
		key := getKey(int64(start))
		val, err := this.db.Get(key)
		if err != nil || len(val) <= 0 {
			return make([]byte, 0), err
		}
		data = append(data, val...)
	}
	return data[preLen : preLen+int(size)], nil
}

// Write write data to buffer
// data: data to be written
// return an error if write failed, otherwise return the data length have been written to buffer.
func (this *SystemBufferContract) Write(data []byte) (uint64, error) {
	saveLen := len(data)
	currentLen := this.Length()
	newStart := currentLen / truncSize

	// TODO：preReserve大小不是一个整的trunc ？？? truncSize - (currentLen % truncSize)
	// fill full the pre-reserve location
	if currentLen%truncSize != 0 {
		preReserve := currentLen % truncSize
		key := getKey(int64(newStart))
		preData, err := this.db.Get(key)
		if err != nil {
			return 0, err
		}

		if uint64(len(data)) > preReserve {
			err = this.db.Put(key, append(preData, data[:preReserve]...))
			data = data[preReserve:]
		} else {
			this.db.Put(key, append(preData, data...))
			data = data[len(data):]
		}
		if err != nil {
			return 0, err
		}
		newStart++
	}

	// save data
	for i := 0; i*truncSize < len(data); i++ {
		key := getKey(int64(newStart))
		start := i * truncSize
		end := len(data)
		if len(data) > (i+1)*truncSize {
			end = (i + 1) * truncSize
		}
		err := this.db.Put(key, data[start:end])
		if err != nil {
			return 0, err
		}
		newStart++
	}
	newLen := currentLen + uint64(saveLen)
	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, newLen)
	err := this.db.Put([]byte(systemBufferCacheKey), val)
	if err != nil {
		return 0, err
	}
	return uint64(saveLen), nil
}

// Length return the length of the data in buffer
func (this *SystemBufferContract) Length() uint64 {
	val, err := this.db.Get([]byte(systemBufferCacheKey))
	if err != nil || val == nil {
		return 0
	}
	return binary.BigEndian.Uint64(val)
}

// Length return the length of the data in buffer
func (this *SystemBufferContract) Close() error {
	cacheLen := this.Length()
	if cacheLen <= 0 {
		return nil
	}

	err := this.db.Delete([]byte(systemBufferCacheKey))
	if err != nil {
		return err
	}

	//TODO：可以整除的情况下是否会多释放？（8/4 == 2，释放0，1，2）
	for i := 0; uint64(i) <= (cacheLen / truncSize); i++ {
		key := getKey(int64(i))
		this.db.Delete(key)
	}
	return nil
}

func (this *SystemBufferContract) Address() types.Address {
	return SystemBufferAddr
}

// get db key based on the offset
func getKey(offset int64) []byte {
	posStart := big.NewInt(0).SetBytes(systemBufferCacheStart)
	pos := posStart.Add(posStart, big.NewInt(offset))
	return math.PaddedBigBytes(pos, util.HashLenght)
}
