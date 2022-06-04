package openvpn

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const someCommonName = "real_name"
const someOtherCommonName = "other_real_name"
const someStatusFilePath = "/does/not/exist"

func generateSomeStatus(commonName string) Status {
	return Status{
		ClientList: []Client{{
			CommonName:     commonName,
			RealAddress:    "",
			BytesReceived:  0,
			BytesSent:      0,
			ConnectedSince: time.Time{},
		}},
		GlobalStats: GlobalStats{},
		ServerInfo:  ServerInfo{},
		UpdatedAt:   time.Time{},
	}
}

func getPseudonymizer(pseudonymLength int) PseudonymizingDecorator {
	return NewOpenVPNPseudonymizingDecorator(pseudonymLength)
}

func getParseFileFunc(status *Status, err error) func(statusfile string) (*Status, error) {
	return func(statusfile string) (*Status, error) {
		return status, err
	}
}

func getCommonName(status *Status) string {
	return status.ClientList[0].CommonName
}

func pseudonymizeStatus(pseudonymizer PseudonymizingDecorator, status Status) (*Status, error) {
	parseFileFuncWithSuccess := getParseFileFunc(&status, nil)
	pseudonymizingParseFileFunc := pseudonymizer.DecorateParseFile(parseFileFuncWithSuccess)
	pseudonymizedStatus, err := pseudonymizingParseFileFunc(someStatusFilePath)
	return pseudonymizedStatus, err
}

func TestCommonNamePseudonymized(t *testing.T) {
	pseudonymizer := getPseudonymizer(5)
	pseudonymizedStatus, err := pseudonymizeStatus(pseudonymizer, generateSomeStatus(someCommonName))
	assert.Nil(t, err)
	assert.NotEqual(t, someCommonName, getCommonName(pseudonymizedStatus))
}

func TestCommonNamePseudonymizedWithCorrectLength(t *testing.T) {
	pseudonymLength := 10
	pseudonymizer := getPseudonymizer(pseudonymLength)
	pseudonymizedStatus, _ := pseudonymizeStatus(pseudonymizer, generateSomeStatus(someCommonName))
	assert.Len(t, getCommonName(pseudonymizedStatus), pseudonymLength)
}

func TestSamePseudonymUsedForSameCommonName(t *testing.T) {
	pseudonymizer := getPseudonymizer(5)
	pseudonymizedStatus1, _ := pseudonymizeStatus(pseudonymizer, generateSomeStatus(someCommonName))
	pseudonymizedStatus2, _ := pseudonymizeStatus(pseudonymizer, generateSomeStatus(someCommonName))
	assert.Equal(t, getCommonName(pseudonymizedStatus1), getCommonName(pseudonymizedStatus2))
}

func TestDifferentPseudonymUsedForDifferentCommonName(t *testing.T) {
	pseudonymizer := getPseudonymizer(5)
	pseudonymizedStatus1, _ := pseudonymizeStatus(pseudonymizer, generateSomeStatus(someCommonName))
	pseudonymizedStatus2, _ := pseudonymizeStatus(pseudonymizer, generateSomeStatus(someOtherCommonName))
	assert.NotEqual(t, getCommonName(pseudonymizedStatus1), getCommonName(pseudonymizedStatus2))
}

func TestPseudonymsNotPersistentAcrossInstances(t *testing.T) {
	pseudonymizer1 := getPseudonymizer(5)
	pseudonymizer2 := getPseudonymizer(5)
	pseudonymizedStatus1, _ := pseudonymizeStatus(pseudonymizer1, generateSomeStatus(someCommonName))
	pseudonymizedStatus2, _ := pseudonymizeStatus(pseudonymizer2, generateSomeStatus(someCommonName))
	assert.NotEqual(t, getCommonName(pseudonymizedStatus1), getCommonName(pseudonymizedStatus2))
}

func TestErrorIsPassedDown(t *testing.T) {
	testError := errors.New("test error")
	pseudonymizer := getPseudonymizer(5)
	parseFileFuncWithFailure := getParseFileFunc(nil, testError)
	pseudonymizingParseFileFunc := pseudonymizer.DecorateParseFile(parseFileFuncWithFailure)
	_, err := pseudonymizingParseFileFunc(someStatusFilePath)
	assert.Equal(t, err, testError)
}
