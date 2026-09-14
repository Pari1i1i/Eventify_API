package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MidtransSnapRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CustomerDetails    *CustomerDetails   `json:"customer_details,omitempty"`
	Expiry             *ExpiryDetails     `json:"expiry,omitempty"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type ExpiryDetails struct {
	StartTime string `json:"start_time,omitempty"`
	Duration  int    `json:"duration"`
	Unit      string `json:"unit"`
}

type MidtransSnapResponse struct {
	Token         string   `json:"token"`
	RedirectURL   string   `json:"redirect_url"`
	ErrorMessages []string `json:"error_messages,omitempty"`
}

func CreateMidtransSnapTransaction(serverKey string, orderCode string, amount float64, customerName, customerEmail string) (*MidtransSnapResponse, error) {
	if serverKey == "" {
		return nil, errors.New("midtrans server key is not configured")
	}

	snapURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"

	reqPayload := MidtransSnapRequest{
		TransactionDetails: TransactionDetails{
			OrderID:     orderCode,
			GrossAmount: int64(amount),
		},
		CustomerDetails: &CustomerDetails{
			FirstName: customerName,
			Email:     customerEmail,
		},
		Expiry: &ExpiryDetails{
			Duration: 2,
			Unit:     "hours",
		},
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snap request: %w", err)
	}

	req, err := http.NewRequest("POST", snapURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	authStr := serverKey + ":"
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Midtrans Snap API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Midtrans response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("midtrans API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var snapResp MidtransSnapResponse
	if err := json.Unmarshal(bodyBytes, &snapResp); err != nil {
		return nil, fmt.Errorf("failed to parse Midtrans response: %w", err)
	}

	return &snapResp, nil
}
