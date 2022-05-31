package vncpasswd

import "crypto/des"

func AuthVNCEncode(plaintext []byte) []byte {
	// 定义vnc加密的key
	key := []byte{23, 82, 107, 6, 35, 78, 88, 7}

	// 密码的每个字节都需要反转。这是一个
	// VNC客户端和服务器的非rfc记录行为
	for i := range key {
		key[i] = (key[i]&0x55)<<1 | (key[i]&0xAA)>>1 // Swap adjacent bits
		key[i] = (key[i]&0x33)<<2 | (key[i]&0xCC)>>2 // Swap adjacent pairs
		key[i] = (key[i]&0x0F)<<4 | (key[i]&0xF0)>>4 // Swap the 2 halves
	}

	// Encrypt challenge with key.
	cipher, err := des.NewCipher(key)
	if err != nil {
		return nil
	}
	for i := 0; i < len(plaintext); i += cipher.BlockSize() {
		cipher.Encrypt(plaintext[i:i+cipher.BlockSize()], plaintext[i:i+cipher.BlockSize()])
	}

	return plaintext
}

func AuthVNCDecrypt(cipherB []byte) []byte {
	// 定义vnc加密的key
	key := []byte{23, 82, 107, 6, 35, 78, 88, 7}

	// 密码的每个字节都需要反转。这是一个
	// VNC客户端和服务器的非rfc记录行为
	for i := range key {
		key[i] = (key[i]&0x55)<<1 | (key[i]&0xAA)>>1 // Swap adjacent bits
		key[i] = (key[i]&0x33)<<2 | (key[i]&0xCC)>>2 // Swap adjacent pairs
		key[i] = (key[i]&0x0F)<<4 | (key[i]&0xF0)>>4 // Swap the 2 halves
	}

	// Encrypt challenge with key.
	cipher, err := des.NewCipher(key)
	if err != nil {
		return nil
	}
	var plaintext []byte
	for i := 0; i < len(cipherB); i += cipher.BlockSize() {
		tmp := make([]byte, i+cipher.BlockSize())
		cipher.Decrypt(tmp, cipherB[i:i+cipher.BlockSize()])
		plaintext = append(plaintext, tmp...)
	}
	return plaintext
}
