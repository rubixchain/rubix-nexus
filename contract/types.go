package contract

// DeploymentResult represents the result of a contract deployment
type DeploymentResult struct {
	ContractHash string
	Success      bool
	Message      string
}

// DeploymentStage represents a stage in the deployment process
type DeploymentStage int

const (
	StageBuild DeploymentStage = iota
	StageGenerate
	StageDeploy
)

// StageCallback is a function that gets called when a stage begins
type StageCallback func(stage DeploymentStage)

// ExecutionResult represents the result of a contract execution
type ExecutionResult struct {
	Success bool
	Message string
	ContractResult string
}

type SmartContractResult struct {
	Id          string `json:"id"`
	Mode        int    `json:"mode"`
	Hash        string `json:"string"`
	OnlyPrivKey bool   `json:"only_priv_key"`
}

// SmartContractAPIResponseV1 represents the standard API response structure
type SmartContractAPIResponseV1 struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// SmartContractAPIResponseV1 represents the standard API response structure
type SmartContractAPIResponseV2 struct {
	Status  bool                `json:"status"`
	Message string              `json:"message"`
	Result  SmartContractResult `json:"result"`
}

type SmartContractBlock struct {
	BlockNo           string `json:"BlockNo"`
	BlockId           string `json:"BlockId"`
	SmartContractData string `json:"SmartContractData"`
}
