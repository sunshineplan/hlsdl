package hlsdl

import (
	"crypto/aes"
	"crypto/cipher"
)

func decryptAES128(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, iv[:block.BlockSize()])
	b := make([]byte, len(data))
	blockMode.CryptBlocks(b, data)

	return pkcs5UnPadding(b), nil
}

func pkcs5UnPadding(b []byte) []byte {
	length := len(b)
	unPadding := int(b[length-1])
	return b[:(length - unPadding)]
}
