package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HTTPClient represents an HTTP client for the wallet API
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateAccountRequest represents the request for creating an account
type CreateAccountRequest struct {
	// Empty for now, could add account options later
}

// CreateAccountResponse represents the response from creating an account
type CreateAccountResponse struct {
	FujiAvaxAddress string `json:"fuji_avax_address"`
	AvaxAddress     string `json:"avax_address"`
	EthAddress      string `json:"eth_address"`
}

// GetAccountRequest represents the request for getting an account
type GetAccountRequest struct {
	Address string `json:"address"`
}

// GetAccountResponse represents the response from getting an account
type GetAccountResponse struct {
	Address   string   `json:"address"`
	Policies  []string `json:"policies"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	KeyType   string   `json:"key_type"`
}

// CreateAccount creates a new account via HTTP
func (c *HTTPClient) CreateAccount(ctx context.Context) (*CreateAccountResponse, error) {
	reqBody := CreateAccountRequest{}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/wallet/accounts", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response CreateAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// GetAccount retrieves an account by address via HTTP
func (c *HTTPClient) GetAccount(ctx context.Context, address string) (*GetAccountResponse, error) {
	url := fmt.Sprintf("%s/v1/wallet/accounts/%s", c.baseURL, address)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GetAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func main() {
	// Create HTTP client
	client := NewHTTPClient("http://localhost:8081")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call CreateAccount
	fmt.Println("Calling CreateAccount via HTTP...")
	createResp, err := client.CreateAccount(ctx)
	if err != nil {
		log.Fatalf("CreateAccount failed: %v", err)
	}

	// Print the response
	fmt.Printf("CreateAccount Response:\n")
	fmt.Printf("  FujiAvaxAddress: %s\n", createResp.FujiAvaxAddress)
	fmt.Printf("  AvaxAddress: %s\n", createResp.AvaxAddress)
	fmt.Printf("  EthAddress: %s\n", createResp.EthAddress)

	// Test GetAccount using the Fuji address
	fmt.Println("\nCalling GetAccount via HTTP...")
	getResp, err := client.GetAccount(ctx, createResp.FujiAvaxAddress)
	if err != nil {
		log.Fatalf("GetAccount failed: %v", err)
	}

	// Print the GetAccount response
	fmt.Printf("GetAccount Response:\n")
	fmt.Printf("  Address: %s\n", getResp.Address)
	fmt.Printf("  Policies: %v\n", getResp.Policies)
	fmt.Printf("  Created At: %s\n", getResp.CreatedAt)
	fmt.Printf("  Updated At: %s\n", getResp.UpdatedAt)
	fmt.Printf("  Key Type: %s\n", getResp.KeyType)

	fmt.Println("\nHTTP API test completed successfully!")
}
