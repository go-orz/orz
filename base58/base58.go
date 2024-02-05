package base58

import (
	"bytes"
	"math/big"
)

const encodeStd = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Encode encodes by base58.
// 通过 base58 编码
func Encode(src []byte) (dst []byte) {
	intBytes := big.NewInt(0).SetBytes(src)
	int0, int58 := big.NewInt(0), big.NewInt(58)
	for intBytes.Cmp(big.NewInt(0)) > 0 {
		intBytes.DivMod(intBytes, int58, int0)
		dst = append(dst, []byte(encodeStd)[int0.Int64()])
	}
	return reverseBytes(dst)
}

// Decode decodes by base58.
// 通过 base58 解码
func Decode(src []byte) []byte {
	bigInt := big.NewInt(0)
	for _, v := range src {
		index := bytes.IndexByte([]byte(encodeStd), v)
		bigInt.Mul(bigInt, big.NewInt(58))
		bigInt.Add(bigInt, big.NewInt(int64(index)))
	}
	return bigInt.Bytes()
}

// reverses byte slice.
func reverseBytes(b []byte) []byte {
	i := 0
	j := len(b) - 1
	for i < j {
		b[i], b[j] = b[j], b[i]
		i++
		j--
	}
	if b == nil {
		return []byte("")
	}
	return b
}
