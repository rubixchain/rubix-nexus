package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rubixchain/rubix-nexus/config"
	"github.com/rubixchain/rubix-nexus/utils"
)

// Deploy handles the contract deployment process
func Deploy(contractDir string, homeDir string, deployerDid string, onStage StageCallback) (*DeploymentResult, error) {
	// Load config to get API URL
	cfg, err := config.LoadConfig(homeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate contract directory
	if !isValidContractDir(contractDir) {
		return nil, fmt.Errorf("invalid contract directory: must contain lib.rs")
	}

	// Check build prerequisites
	if err := verifyBuildPrerequisites(); err != nil {
		return nil, err
	}

	onStage(StageBuild)
	// Build Rust project to WASM
	wasmPath, err := buildWasm(contractDir)
	if err != nil {
		return nil, fmt.Errorf("failed to build WASM: %w", err)
	}

	// Get file paths
	libPath := filepath.Join(contractDir, "src", "lib.rs")
	statePath := filepath.Join(filepath.Dir(contractDir), "artifacts", "state.json")

	// Create empty state.json if it doesn't exist
	if !utils.FileExists(statePath) {
		if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create artifacts directory: %w", err)
		}
		if err := os.WriteFile(statePath, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("failed to create state.json: %w", err)
		}
	}

	onStage(StageGenerate)
	contractHash, err := generateSmartContract(cfg.Network.DeployerNodeURL, deployerDid, wasmPath, libPath, statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate smart contract: %w", err)
	}

	onStage(StageDeploy)

	// Call deploy-smart-contract API
	requestID, err := deploySmartContract(cfg.Network.DeployerNodeURL, contractHash, deployerDid)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy smart contract: %w", err)
	}

	// Call signature-response API
	if err := signatureResponse(cfg.Network.DeployerNodeURL, requestID); err != nil {
		return nil, fmt.Errorf("failed to process signature response: %w", err)
	}

	return &DeploymentResult{
		ContractHash: contractHash,
		Success:      true,
		Message:      "Contract deployed successfully",
	}, nil
}

func generateSmartContract(baseURL, deployerDid, wasmPath, libPath, statePath string) (string, error) {
	// Create a buffer to store the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the deployerDid field
	if err := writer.WriteField("did", deployerDid); err != nil {
		return "", fmt.Errorf("failed to add did field: %w", err)
	}

	// Add the WASM file
	wasmFile, err := os.Open(wasmPath)
	if err != nil {
		return "", fmt.Errorf("failed to open WASM file: %w", err)
	}
	defer wasmFile.Close()
	wasmPart, err := writer.CreateFormFile("binaryCodePath", filepath.Base(wasmPath))
	if err != nil {
		return "", fmt.Errorf("failed to create WASM form file: %w", err)
	}
	if _, err := io.Copy(wasmPart, wasmFile); err != nil {
		return "", fmt.Errorf("failed to copy WASM file: %w", err)
	}

	// Add the lib.rs file
	libFile, err := os.Open(libPath)
	if err != nil {
		return "", fmt.Errorf("failed to open lib.rs file: %w", err)
	}
	defer libFile.Close()
	libPart, err := writer.CreateFormFile("rawCodePath", filepath.Base(libPath))
	if err != nil {
		return "", fmt.Errorf("failed to create lib.rs form file: %w", err)
	}
	if _, err := io.Copy(libPart, libFile); err != nil {
		return "", fmt.Errorf("failed to copy lib.rs file: %w", err)
	}

	// Add the state.json file
	stateFile, err := os.Open(statePath)
	if err != nil {
		return "", fmt.Errorf("failed to open state.json file: %w", err)
	}
	defer stateFile.Close()
	statePart, err := writer.CreateFormFile("schemaFilePath", filepath.Base(statePath))
	if err != nil {
		return "", fmt.Errorf("failed to create state.json form file: %w", err)
	}
	if _, err := io.Copy(statePart, stateFile); err != nil {
		return "", fmt.Errorf("failed to copy state.json file: %w", err)
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create the request
	url := fmt.Sprintf("%s/api/generate-smart-contract", baseURL)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "multipart/form-data")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var apiResp SmartContractAPIResponseV1
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check response status
	if !apiResp.Status {
		return "", fmt.Errorf(apiResp.Message)
	}

	return apiResp.Result, nil
}

func deploySmartContract(baseURL, contractHash, deployerDid string) (string, error) {
	// Create request body
	requestBody := struct {
		Comment            string  `json:"comment"`
		DeployerAddr       string  `json:"deployerAddr"`
		QuorumType         int     `json:"quorumType"`
		RbtAmount          float64 `json:"rbtAmount"`
		SmartContractToken string  `json:"smartContractToken"`
	}{
		Comment:            "Contract deployment",
		DeployerAddr:       deployerDid,
		QuorumType:         2,
		RbtAmount:          0.001,
		SmartContractToken: contractHash,
	}

	// Marshal request body
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	requestURL, err := url.JoinPath(baseURL, "/api/deploy-smart-contract")
	if err != nil {
		return "", fmt.Errorf("deploy: unable to form request URL")
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var apiResp SmartContractAPIResponseV2
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check response status
	if !apiResp.Status {
		return "", fmt.Errorf(apiResp.Message)
	}

	return apiResp.Result.Id, nil
}

// verifyBuildPrerequisites verifies that all required build tools are available
func verifyBuildPrerequisites() error {
	// Check if cargo is available
	if _, err := exec.LookPath("cargo"); err != nil {
		return fmt.Errorf("Rust toolchain not found. Please install Rust from https://rustup.rs/")
	}

	// Check for wasm32-unknown-unknown target
	cmd := exec.Command("rustup", "target", "list", "--installed")
	output, err := cmd.Output()
	if err != nil || !strings.Contains(string(output), "wasm32-unknown-unknown") {
		// Try to add the target
		addCmd := exec.Command("rustup", "target", "add", "wasm32-unknown-unknown")
		if err := addCmd.Run(); err != nil {
			return fmt.Errorf("failed to add wasm32-unknown-unknown target: %w", err)
		}
	}

	// Windows-specific checks
	// 	if runtime.GOOS == "windows" {
	// 		// Check for MSVC build tools
	// 		if _, err := exec.LookPath("link.exe"); err != nil {
	// 			return fmt.Errorf(`Build tools not found. On Windows, you need:

	// 1. Visual Studio Build Tools with C++ support
	//    Download from: https://visualstudio.microsoft.com/visual-cpp-build-tools/

	// 2. During installation, select "Desktop development with C++"

	// Alternative: Consider using Windows Subsystem for Linux (WSL)
	// 1. Install WSL: wsl --install
	// 2. Install Rust in WSL
	// 3. Run this tool in WSL`)
	// 		}
	// 	}

	return nil
}

// buildWasm builds the Rust project targeting wasm32-unknown-unknown
func buildWasm(projectDir string) (string, error) {
	// Create target directory if it doesn't exist
	targetDir := filepath.Join(projectDir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	// Build the project
	buildCmd := exec.Command("cargo", "build", "--target", "wasm32-unknown-unknown")
	buildCmd.Dir = projectDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		// Provide more context for build failures
		errMsg := string(output)
		if runtime.GOOS == "windows" && strings.Contains(errMsg, "linker `link.exe` not found") {
			return "", fmt.Errorf("MSVC build tools not found. Please install Visual Studio Build Tools with C++ support")
		}
		return "", fmt.Errorf("build failed: %s: %w", errMsg, err)
	}

	// Get the WASM file name from the directory name, replacing hyphens with underscores
	projectName := strings.ReplaceAll(filepath.Base(projectDir), "-", "_")
	wasmFile := filepath.Join(projectDir, "target", "wasm32-unknown-unknown", "debug", projectName+".wasm")

	// Verify the WASM file was created
	if !utils.FileExists(wasmFile) {
		return "", fmt.Errorf("WASM file not found after build at %s", wasmFile)
	}

	// Create artifacts directory and copy WASM file
	artifactsDir := filepath.Join(filepath.Dir(projectDir), "artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	targetFile := filepath.Join(artifactsDir, projectName+".wasm")
	input, err := os.ReadFile(wasmFile)
	if err != nil {
		return "", fmt.Errorf("failed to read WASM file: %w", err)
	}

	if err := os.WriteFile(targetFile, input, 0644); err != nil {
		return "", fmt.Errorf("failed to copy WASM file to artifacts: %w", err)
	}

	return targetFile, nil
}

// isValidContractDir checks if the directory contains required contract files
func isValidContractDir(dir string) bool {
	// Only check for lib.rs, as artifacts will be created during build
	libPath := filepath.Join(dir, "src", "lib.rs")
	return utils.FileExists(libPath)
}

// Dummy API functions (to be implemented with real API calls)
// func registerCallbackURL(baseURL string, dappServer string, contractEndpoint string, contractHash string) error {
// 	if contractEndpoint == "" {
// 		return fmt.Errorf("smart contract callback endpoint for %v cannot be empty", contractHash)
// 	}

// 	if dappServer == "" {
// 		return fmt.Errorf("smart contract dapp server url for %v cannot be empty", contractHash)
// 	}

// 	callbackURL, err := url.JoinPath(dappServer, contractEndpoint)
// 	if err != nil {
// 		return fmt.Errorf("unable to form callback url, err: %v", err)
// 	}

// 	requestBody := struct {
// 		CallbackURL        string `json:"CallBackURL"`
// 		SmartContractToken string `json:"SmartContractToken"`
// 	}{
// 		CallbackURL:        callbackURL,
// 		SmartContractToken: contractHash,
// 	}

// 	fmt.Printf("Request Body: %v", requestBody)

// 	// Marshal request body
// 	bodyBytes, err := json.Marshal(requestBody)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal request body: %w", err)
// 	}

// 	// Create request
// 	requestURL, err := url.JoinPath(baseURL, "/api/register-callback-url")
// 	if err != nil {
// 		return fmt.Errorf("deploy: unable to form request URL")
// 	}

// 	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyBytes))
// 	if err != nil {
// 		return fmt.Errorf("failed to create request: %w", err)
// 	}

// 	// Set headers
// 	req.Header.Set("Content-Type", "application/json")

// 	// Send request
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("failed to send request: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return fmt.Errorf("failed to read response: %w", err)
// 	}

// 	// Parse response
// 	var apiResp SmartContractAPIResponseV1
// 	if err := json.Unmarshal(body, &apiResp); err != nil {
// 		return fmt.Errorf("failed to parse response: %w", err)
// 	}

// 	// Check response status
// 	if !apiResp.Status {
// 		return fmt.Errorf(apiResp.Message)
// 	}

// 	fmt.Println("Callback URL %v for contract %v registered successfully", callbackURL)
// 	return nil
// }

func signatureResponse(baseURL, requestID string) error {
	// Create request body
	requestBody := struct {
		Id       string `json:"id"`
		Mode     int    `json:"mode"`
		Password string `json:"password"`
	}{
		Id:       requestID,
		Mode:     0,
		Password: "mypassword",
	}

	// Marshal request body
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	requestURL, err := url.JoinPath(baseURL, "/api/signature-response")
	if err != nil {
		return fmt.Errorf("signature response: unable to form request URL")
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("signature request: failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("signature request: failed to read response: %w", err)
	}

	// Parse response
	var apiResp SmartContractAPIResponseV1
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("signature request: failed to parse response: %w", err)
	}

	// Check response status
	if !apiResp.Status {
		return fmt.Errorf(apiResp.Message)
	}

	return nil
}
