package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	status, err := GetKycStatus("ABCDE1234F")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Kyc Status: ", status)
}

func GetKycStatus(pan string) (string, error) {
	// check Pan
	err := validatePan(pan)
	if err != nil {
		return "", fmt.Errorf("invalid pan")
	}

	// get kyc status
	status, err := GetKycStatusFromProvider(pan) // HL1
	if err != nil {                              // HL1
		return "", fmt.Errorf("failed to get kyc status") // HL1
	} // HL1

	return status, nil
}

func validatePan(pan string) error {
	// validate Pan
	if len(pan) != 10 {
		return fmt.Errorf("expected pan length is 10 but got %d", len(pan))
	}
	return nil
}

type KycServiceImpl struct {
	httpClient *http.Client
}

func NewKycServiceImpl(httpClient *http.Client) *KycServiceImpl {
	return &KycServiceImpl{httpClient: httpClient}
}

func GetKycStatusFromProvider(pan string) (string, error) {
	time.Sleep(1 * time.Second)
	return "KYC Valid", nil
}
