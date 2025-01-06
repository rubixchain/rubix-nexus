package did

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GenerateOneTestRBT(baseURL string, did string) (error) {
	requestURL, err := url.JoinPath(baseURL, "/api/generate-test-token")
	if err != nil {
		return fmt.Errorf("generate test token: unable to form request URL")
	}

	tokenRequest := rbtGenerateRequest{
		DID:            did,
		NumberOfTokens: 1,
	}

	// Marshal the struct to JSON
	jsonData, err := json.Marshal(tokenRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Create a POST request
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	err3 := json.Unmarshal(body, &response)
	if err3 != nil {
		fmt.Println("Error unmarshaling response:", err3)
	}

	result := response["result"].(map[string]interface{})
	id := result["id"].(string)

	signRespErr := signatureResponse(baseURL, id)
	if signRespErr != nil {
		return fmt.Errorf("failed to sign response: %v", signRespErr)
	}

	return signRespErr
}
