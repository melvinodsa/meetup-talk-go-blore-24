package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	prov1 := NewMyKycVendorService(&http.Client{})
	ser1 := NewKycServiceImpl(prov1)
	handler := GetKycStatusController(ser1)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL + "?pan=ABCDE1234F")
	if err != nil {
		log.Fatal(err)
	}
	status, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(status))
}

// START SERVICE1 OMIT
type KycService interface {
	GetKycStatus(pan string) (string, error)
}

type KycServiceImpl struct {
	kyc KycProvider
}

func NewKycServiceImpl(kycProvider KycProvider) *KycServiceImpl {
	return &KycServiceImpl{kyc: kycProvider}
}

// END SERVICE1 OMIT

// START SERVICE2 OMIT
func (k KycServiceImpl) GetKycStatus(pan string) (string, error) {
	// check Pan
	err := validatePan(pan)
	if err != nil {
		return "", fmt.Errorf("invalid pan")
	}

	// get kyc status
	status, err := k.kyc.GetKycStatus(pan)
	if err != nil {
		return "", fmt.Errorf("failed to get kyc status from provider")
	}

	return status, nil
}

// END SERVICE2 OMIT

func GetKycStatusController(kServ KycService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pan := r.URL.Query().Get("pan")
		status, err := kServ.GetKycStatus(pan) // HL1
		if err != nil {                        // HL1
			http.Error(w, fmt.Errorf("failed to get kyc status %w", err).Error(), http.StatusInternalServerError) // HL1
			return                                                                                                // HL1
		} // HL1
		fmt.Fprintf(w, "Kyc Status: %s", status)
	}
}

func validatePan(pan string) error {
	// validate Pan
	if len(pan) != 10 {
		return fmt.Errorf("expected pan length is 10 but got %d", len(pan))
	}
	return nil
}

// START REPOSITORY OMIT
type KycProvider interface {
	GetKycStatus(pan string) (string, error)
}

type MyKycVendorService struct {
	httpClient *http.Client
}

func NewMyKycVendorService(httpClient *http.Client) *MyKycVendorService {
	if httpClient == nil { // HL1
		// we can have our internal http clients // HL1
		httpClient = &http.Client{} // HL1
	} // HL1
	return &MyKycVendorService{httpClient: httpClient}
}

func (k *MyKycVendorService) GetKycStatus(pan string) (string, error) {
	time.Sleep(1 * time.Second)
	return "KYC Valid", nil
}

// END REPOSITORY OMIT
