package did

type createDidResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Result  struct {
		DID    string `json:"did"`
		PeerID string `json:"peer_id"`
	} `json:"result"`
}

type registerDidResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Id string `json:"id"`
		Mode int `json:"mode"`
		OnlyPrivKey bool `json:"only_priv_key"`
	}
}

type singatureResReponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

type rbtGenerateRequest struct {
	NumberOfTokens int    `json:"number_of_tokens"`
	DID            string `json:"did"`
}