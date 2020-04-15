package main

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/aead/ecdh"
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
func (env *Env) Encrypt(data []byte) ([]byte, error) {
	// Allocate a nonce with the NonceSize length and capacity of the whole encrypted data
	nonce := make([]byte, env.aead.NonceSize(), env.aead.NonceSize()+len(data)+env.aead.Overhead())

	// We can use secretbox which uses XChaCha20-Poly1305 AEAD to encrypt and authenticate messages
	// Generate random nonce
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("error while encrypting data: %v", err)
	}

	// Seal using AEAD.
	// We are adding our encrypted data to nonce
	return env.aead.Seal(nonce, nonce, data, nil), nil
}

// Decrypt the given data with the given key
// We will use XSalsa20 and Poly1305 to decrypt and authenticate messages
func (env *Env) Decrypt(encData []byte) ([]byte, error) {
	if len(encData) < env.aead.NonceSize() {
		return nil, errors.New("invalid ciphertext")
	}

	return env.aead.Open(nil, encData[:env.aead.NonceSize()], encData[env.aead.NonceSize():], nil)
}
