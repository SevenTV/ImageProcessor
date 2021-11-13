package utils

import (
	"crypto/rand"
	"encoding/base64"
	"reflect"
	"unsafe"
)

func OrPanic(args ...interface{}) []interface{} {
	for _, v := range args {
		switch v.(type) {
		case error:
			if v != nil {
				panic(v)
			}
		}
	}
	return args
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

//
// Util - Is Power Of Two
//
func IsPowerOfTwo(n int64) bool {
	return (n != 0) && ((n & (n - 1)) == 0)
}

// GenerateRandomString returns a URL-safe
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.StdEncoding.EncodeToString(b)[s:], err
}

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func B2S(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// S2B converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
// Warning: Writing data into the array is not a good idea, as this will result in a segfault
// This is really bad do not writing or modity the data in the array which is returned.
func S2B(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

func StringPointer(s string) *string {
	return &s
}

func Int64Pointer(i int64) *int64 {
	return &i
}
