package did

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/rubixchain/rubix-nexus/config"
)

func CreateDID(homeDir string, isLocalnet bool) (string, error) {
	cfg, err := config.LoadConfig(homeDir)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	requestURL, err := url.JoinPath(cfg.Network.DeployerNodeURL, "/api/createdid")
	if err != nil {
		return "", fmt.Errorf("failed to join URL: %v", err)
	}

	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	didConfig := map[string]interface{}{
		"Type":          4,
		"priv_pwd":      "mypassword",
		"mnemonic_file": "",
		"childPath":     0,
	}
	didConfigBytes, err := json.Marshal(didConfig)
	if err != nil {
		return "", fmt.Errorf("failed to encode didConfig: %v", err)
	}

	err = writer.WriteField("did_config", string(didConfigBytes))
	if err != nil {
		return "", fmt.Errorf("failed to write didConfig field: %v", err)
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Create a POST request
	req, err := http.NewRequest("POST", requestURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if the response status is not 200
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(responseBody))
	}
	// Parse the JSON response
	var response createDidResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return "", fmt.Errorf("createDID: failed to parse response: %v", err)
	}

	// Check if the API call was successful
	if !response.Status {
		return "", fmt.Errorf("API error: %s", response.Message)
	}

	registerDidErr := registerDID(cfg.Network.DeployerNodeURL, response.Result.DID)
	if registerDidErr != nil {
		return "", fmt.Errorf("failed to register DID: %v", registerDidErr)
	}

	if isLocalnet {
		errGenerateTestRBT := GenerateOneTestRBT(cfg.Network.DeployerNodeURL, response.Result.DID)
		if errGenerateTestRBT != nil {
			return "", fmt.Errorf("failed to generate test RBT: %v", errGenerateTestRBT)
		}
	}

	// Return the DID from the response
	return response.Result.DID, nil
}

func registerDID(baseURL string, did string) error {
	requestURL, err := url.JoinPath(baseURL, "/api/register-did")
	if err != nil {
		return fmt.Errorf("failed to join URL: %v", err)
	}

	requestBody := map[string]interface{}{
		"did": did,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(responseBody))
	}

	var registerDidResp registerDidResponse
	if err = json.Unmarshal(responseBody, &registerDidResp); err != nil {
		return fmt.Errorf("registerDID: failed to unmarshal JSON: %v", err)
	}

	if !registerDidResp.Status {
		return fmt.Errorf("failed to Register DID, error: %s", registerDidResp.Message)
	}

	requestId := registerDidResp.Result.Id

	if err = signatureResponse(baseURL, requestId); err != nil {
		return fmt.Errorf("failed to send signature response: %v", err)
	}

	return nil
}

func signatureResponse(baseURL, requestId string) error {
	data := map[string]interface{}{
		"id":       requestId,
		"mode":     0,
		"password": "mypassword",
	}

	bodyJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON, err: %v", err)
	}

	requestURL, err := url.JoinPath(baseURL, "/api/signature-response")
	if err != nil {
		return fmt.Errorf("failed to join URL: %v", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return fmt.Errorf("error creating HTTP request, err: %v", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request, err: %v", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s\n", err)
	}

	return nil
}
