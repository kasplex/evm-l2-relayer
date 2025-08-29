package impl

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/sha3"
)

func keccak256(data []byte) []byte {
	d := sha3.NewLegacyKeccak256()
	d.Write(data)
	return d.Sum(nil)
}

func decodeToHex(str string) ([]byte, error) {
	str = strings.TrimPrefix(str, "0x")
	if len(str)%2 != 0 {
		str = "0" + str
	}
	return hex.DecodeString(str)
}

func zlibCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	zw.Close()
	return buf.Bytes(), nil
}
