package utils

import (
	"encoding/hex"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

func Test_OrPanic(t *testing.T) {
	func() {
		defer func() {
			assert.Nil(t, recover(), "The function did not panic")
		}()
		var err error
		OrPanic(10, err)
	}()

	func() {
		defer func() {
			assert.NotNil(t, recover(), "The function did panic")
		}()
		OrPanic(10, fmt.Errorf("bad"))
	}()
}

func Test_GenerateRandomBytes(t *testing.T) {
	arr1, err := GenerateRandomBytes(65535)
	assert.ErrorIs(t, err, nil, "no error when generating the array")
	arr2, err := GenerateRandomBytes(65535)
	assert.ErrorIs(t, err, nil, "no error when generating the array")

	hash := func(data []byte) string {
		h := sha3.New512()
		h.Write(data)
		return hex.EncodeToString(h.Sum(nil))
	}

	assert.NotEqual(t, hash(arr1), hash(arr2), "The bytes are random")
}

func Test_IsPowerOfTwo(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		n := math.Log2(float64(i))
		pot := n == math.Floor(n)
		if i == 0 {
			pot = false
		}
		assert.Equal(t, pot, IsPowerOfTwo(i), fmt.Sprintf("%d power of 2", i))
	}
}

func Test_GenerateRandomString(t *testing.T) {
	arr1, err := GenerateRandomString(65535)
	assert.ErrorIs(t, err, nil, "no error when generating the array")
	arr2, err := GenerateRandomString(65535)
	assert.ErrorIs(t, err, nil, "no error when generating the array")

	assert.NotEqual(t, arr1, arr2, "Strings are random")
}

func Test_B2S(t *testing.T) {
	arr := []byte("testing")
	str := B2S(arr)

	assert.Equal(t, string(arr), str, "The string was converted successfully")

	arr[0] = 'T'

	assert.Equal(t, string(arr), str, "The string was converted without memory allocation")
}

func Test_S2B(t *testing.T) {
	iArr := []byte("testing")
	str := B2S(iArr)
	arr := S2B(str)

	assert.Equal(t, string(arr), str, "The string was converted successfully")

	arr[0] = 'T'

	assert.Equal(t, string(arr), str, "The string was converted without memory allocation")
}

func Test_StringPointer(t *testing.T) {
	str := "pogu"

	strP := StringPointer(str)

	assert.Equal(t, str, *strP, "The string is a new pointer")
}

func Test_Int64Pointer(t *testing.T) {
	i64 := int64(100)

	i64P := Int64Pointer(i64)

	assert.Equal(t, i64, *i64P, "The int64 is a new pointer")
}
