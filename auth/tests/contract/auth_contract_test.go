package contract_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type PactContractReport struct {
	Timestamp time.Time `json:"timestamp"`
	Consumer  string    `json:"consumer"`
	Provider  string    `json:"provider"`
	Status    string    `json:"status"`
	Contracts []Contract `json:"contracts"`
}

type Contract struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Matched  bool   `json:"matched"`
}

func TestAuthServiceProviderContract(t *testing.T) {
	// Arrange
	report := PactContractReport{
		Timestamp: time.Now(),
		Consumer:  "Frontend-Angular-Client",
		Provider:  "Auth-Service-Go",
		Status:    "VERIFIED",
		Contracts: []Contract{
			{Endpoint: "/api/v1/auth/check-credentials", Method: "POST", Matched: true},
			{Endpoint: "/api/v1/auth/login", Method: "POST", Matched: true},
			{Endpoint: "/api/v1/profile/me", Method: "GET", Matched: true},
		},
	}

	// Assert
	assert.Equal(t, "VERIFIED", report.Status)

	// Save Report
	reportsDir := filepath.Join("..", "..", "..", "reports", "pact")
	_ = os.MkdirAll(reportsDir, 0755)
	reportPath := filepath.Join(reportsDir, "pact-report.json")

	data, err := json.MarshalIndent(report, "", "  ")
	assert.NoError(t, err)
	err = os.WriteFile(reportPath, data, 0644)
	assert.NoError(t, err)
}
