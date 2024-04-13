package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"
)

const key = "rckuSMeBTCDj8a2k8RUJR8IIaccxG3AE"

func main() {
	testSuite := []testing.InternalTest{
		{
			Name: "TestPii",
			F:    TestPii,
		},
	}
	testing.Main(matchString, testSuite, nil, nil)
}

func matchString(a, b string) (bool, error) {
	return a == b, nil
}

var reusePii pii

func init() {
	enc, err := NewDummyEncrypter(key)
	if err != nil {
		panic(err)
	}
	pii, err := NewPii("ABCDE1234F", enc)
	if err != nil {
		panic(err)
	}
	reusePii = *pii
}

// START TESTCASEBASE OMIT
type piiTestcase struct {
	name           string
	pan            *pii
	expectedString string
	expectedValue  string
}

var piiTestcases = []piiTestcase{
	{name: "Normal PII", pan: &reusePii, expectedString: "🔒", expectedValue: "ABCDE1234F"},
	{name: "Reuse PII", pan: &reusePii, expectedString: "🔒", expectedValue: "🚫"},
}

// END TESTCASEBASE OMIT

// START RUNTEST OMIT
func TestPii(t *testing.T) {
	enc, err := NewDummyEncrypter(key)
	if err != nil {
		t.Errorf("failed to create encrypter: %v", err)
	}
	for _, tc := range piiTestcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pan.String(); got != tc.expectedString {
				t.Errorf("🤕 String() = %v, want %v", got, tc.expectedString)
			}
			if got, _ := tc.pan.GetValue(enc); got != tc.expectedValue {
				t.Errorf("🤕 GetValue() = %v, want %v", got, tc.expectedValue)
			}
			fmt.Printf("👍 Test case %s passed\n", tc.name)
		})
	}
}

// END RUNTEST OMIT

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

// START MOCK OMIT
type encrypter struct {
	cipher cipher.Block
}

func NewDummyEncrypter(key string) (*encrypter, error) {
	keyB := []byte(key)
	block, err := aes.NewCipher(keyB)
	if err != nil {
		return nil, err
	}
	return &encrypter{cipher: block}, nil
}

// END MOCK OMIT

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
