package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const secret = "rckuSMeBTCDj8a2k8RUJR8IIaccxG3AE"
const url = "https://mykycvendor.com"
const apiKey = "mysecret"

// START MAIN OMIT
func main() {
	enc, err := NewEncrypter(url, apiKey, secret)
	if err != nil {
		panic(err)
	}
	pan, err := NewPii("ABCDE1234F", enc)
	if err != nil {
		panic(err)
	}
	s := secretData{pan: pan}
	fmt.Println("Secret", s)
	fmt.Printf("Secret %+v\n", s)
	fmt.Println("Secret", s.GetPan())
	fmt.Println(s.GetPan().GetValue(enc))
	fmt.Println(s.GetPan().GetValue(enc))
}

// END MAIN OMIT

type secretData struct {
	pan *pii
}

func (s secretData) GetPan() *pii {
	return s.pan
}

type pii struct {
	data interface{}
}

// START ENCRYPTPROVIDER OMIT

func NewPii(data interface{}, enc EncryptProvider) (*pii, error) {
	d := fmt.Sprintf("%+v", data)
	encryptedData, err := enc.Encrypt(d)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt pii")
	}
	return &pii{data: encryptedData}, nil
}

type EncryptProvider interface {
	Encrypt(data string) (string, error)
	Decrypt(data string) (string, error)
}

// END ENCRYPTPROVIDER OMIT

// START PIIDEF OMIT

func (p pii) String() string {
	return "🔒"
}

func (p *pii) GetValue(enc EncryptProvider) (interface{}, error) {
	d, err := enc.Decrypt(p.data.(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt pii")
	}
	var res interface{}
	fmt.Sscanf(d, "%+v", &res)
	p.data, err = enc.Encrypt("🚫") // HL1
	if err != nil {                // HL1
		return nil, fmt.Errorf("failed to encrypt pii") // HL1
	} // HL1
	return d, nil
}

// END PIIDEF OMIT

type encrypter struct {
	cipher cipher.Block
}

func NewEncrypter(apiUrl, apiKey, secret string) (*encrypter, error) {
	keyB := []byte(secret)
	block, err := aes.NewCipher(keyB)
	if err != nil {
		return nil, err
	}
	return &encrypter{cipher: block}, nil
}

func (e encrypter) Encrypt(data string) (string, error) {
	plaintext := []byte(data)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(e.cipher, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e encrypter) Decrypt(data string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(e.cipher, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}
