package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type MidtransQRISResponse struct {
	StatusCode        string            `json:"status_code"`
	StatusMessage     string            `json:"status_message"`
	TransactionID     string            `json:"transaction_id"`
	OrderID           string            `json:"order_id"`
	GrossAmount       string            `json:"gross_amount"`
	PaymentType       string            `json:"payment_type"`
	TransactionTime   string            `json:"transaction_time"`
	TransactionStatus string            `json:"transaction_status"`
	Actions           []MidtransAction  `json:"actions"`
	QRString          string            `json:"qr_string"`
}

type MidtransAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

type MidtransChargeRequest struct {
	PaymentType        string                 `json:"payment_type"`
	TransactionDetails MidtransTxDetail       `json:"transaction_details"`
	ItemDetails        []MidtransItemDetail   `json:"item_details,omitempty"`
	CustomerDetails    *MidtransCustomerDetail `json:"customer_details,omitempty"`
}

type MidtransTxDetail struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type MidtransItemDetail struct {
	ID       string `json:"id"`
	Price    int64  `json:"price"`
	Quantity int32  `json:"quantity"`
	Name     string `json:"name"`
}

type MidtransCustomerDetail struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

// ChargeMidtransQRIS sends request to Midtrans Core API Sandbox/Production to get genuine QRIS
func ChargeMidtransQRIS(serverKey string, isProduction bool, req MidtransChargeRequest) (*MidtransQRISResponse, error) {
	req.PaymentType = "qris"

	baseURL := "https://api.sandbox.midtrans.com/v2/charge"
	if isProduction {
		baseURL = "https://api.midtrans.com/v2/charge"
	}

	reqBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, err
	}

	authStr := base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(serverKey) + ":"))
	httpReq.Header.Set("Authorization", "Basic "+authStr)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed calling midtrans charge API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var midtransResp MidtransQRISResponse
	if err := json.Unmarshal(bodyBytes, &midtransResp); err != nil {
		return nil, fmt.Errorf("error parsing midtrans response: %w (raw: %s)", err, string(bodyBytes))
	}

	// Midtrans returns 201 for successfully created charge
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.New(midtransResp.StatusMessage)
	}

	return &midtransResp, nil
}
