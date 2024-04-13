package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func main() {
	testSuite := []testing.InternalTest{
		{
			Name: "TestGetKycStatus",
			F:    TestGetKycStatus,
		},
	}
	testing.Main(matchString, testSuite, nil, nil)
}

func matchString(a, b string) (bool, error) {
	return a == b, nil
}

// START TESTCASEBASE OMIT
var kycResMap = map[string]kycResp{
	"ABCDE1234F": {status: "KYC Valid", err: nil},
	"ABCDE1234D": {status: "KYC Invalid", err: nil},
	"ABCDE1234E": {status: "", err: fmt.Errorf("pan not found")},
}

type kycTestcase struct {
	name     string
	pan      string
	expected string
}

var kycTestcases = []kycTestcase{
	{name: "Positive Flow", pan: "ABCDE1234F", expected: "Kyc Status: KYC Valid"},
	{name: "Invalid Pan", pan: "ABCDE1234", expected: "failed to get kyc status invalid pan"},
	{name: "Pan Not Found", pan: "ABCDE1234E", expected: "failed to get kyc status failed to get kyc status from provider"},
	{name: "Negative Flow", pan: "ABCDE1234D", expected: "Kyc Status: KYC Invalid"},
}

// END TESTCASEBASE OMIT

func (k kycTestcase) runTest(ts *httptest.Server) string {
	res, err := http.Get(ts.URL + fmt.Sprintf("?pan=%s", k.pan))
	if err != nil {
		return fmt.Sprintf("failed to get response: %v", err)
	}
	status, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return fmt.Sprintf("failed to read response body: %v", err)
	}
	return strings.TrimSpace(string(status))
}

// START MOCK OMIT
type kycResp struct {
	status string
	err    error
}
type KycProviderMock struct {
	respMap map[string]kycResp
}

func NewKycProviderMock(r map[string]kycResp) *KycProviderMock {
	return &KycProviderMock{respMap: r}
}

func (k *KycProviderMock) GetKycStatus(pan string) (string, error) {
	resp, ok := k.respMap[pan] // HL1
	if !ok {                   // HL1
		return "", fmt.Errorf("pan not found") // HL1
	} // HL1
	return resp.status, resp.err
}

// END MOCK OMIT

// START RUNTEST OMIT
func TestGetKycStatus(t *testing.T) {
	for _, tc := range kycTestcases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(inializeTest())
			defer ts.Close()

			status := tc.runTest(ts)

			if status != tc.expected {
				t.Errorf("🤕 expected \"%s\" but got \"%s\"", tc.expected, status)
				return
			}
			fmt.Printf("👍 Test case %s passed\n", tc.name)
		})
	}
}

// END RUNTEST OMIT

// START TESTSETUP OMIT
func inializeTest() http.HandlerFunc {
	prov1 := NewKycProviderMock(kycResMap)
	ser1 := NewKycServiceImpl(prov1)
	handler := GetKycStatusController(ser1)
	return handler
}

// END TESTSETUP OMIT

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
