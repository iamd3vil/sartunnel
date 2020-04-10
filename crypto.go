package main

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/aead/ecdh"
	"golang.org/x/crypto/nacl/secretbox"
)

// GenerateKeys generates EC-Curve22591 private and public keys for the given curve
func GenerateKeys(curve ecdh.KeyExchange) (private crypto.PrivateKey, public crypto.PublicKey, err error) {
	return curve.GenerateKey(rand.Reader)
}

// EncodePrivateKey takes a private key and returns a encoded string
func EncodePrivateKey(privKey crypto.PrivateKey) string {
	b := privKey.([32]byte)
	return base64.StdEncoding.EncodeToString(b[:])
}

// EncodePublicKey takes a public key and returns a encoded string
func EncodePublicKey(pubKey crypto.PublicKey) string {
	b := pubKey.([32]byte)
	return base64.StdEncoding.EncodeToString(b[:])
}

// DecodePrivateKey decodes an encoded string into private key
func DecodePrivateKey(privKey string) (crypto.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, err
	}

	return crypto.PrivateKey(b), nil
}

// DecodePublicKey decodes an encoded string into public key
func DecodePublicKey(pubKey string) (crypto.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}

	return crypto.PublicKey(b), nil
}

// Encrypt encrypts the given data with the given key and returns it.
// We will use XSalsa20 and Poly1305 to encrypt and authenticate messages
func Encrypt(key []byte, data []byte) ([]byte, error) {
	// Encrypt the marshalled data using the key.
	var (
		k [32]byte
	)
	copy(k[:], key)

	// We can use secretbox which uses XSalsa20 and Poly1305 to encrypt and authenticate messages
	// Generate random nonce
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return []byte{}, fmt.Errorf("error while encrypting data: %v", err)
	}

	// Seal using secretbox
	return secretbox.Seal(nonce[:], data, &nonce, &k), nil
}

// Decrypt the given data with the given key
// We will use XSalsa20 and Poly1305 to decrypt and authenticate messages
func Decrypt(key []byte, encData []byte) ([]byte, error) {
	var (
		decryptNonce [24]byte
		k            [32]byte
	)
	copy(decryptNonce[:], encData[:24])
	copy(k[:], key[:])

	decrypted, ok := secretbox.Open(nil, encData[24:], &decryptNonce, &k)
	if !ok {
		return nil, errors.New("invalid packet")
	}

	return decrypted, nil
}
