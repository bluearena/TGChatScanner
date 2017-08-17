package clarifaiApi

type (
	Status struct {
		Code        uint   `json:"code"`
		Description string `json:"description"`
	}

	Concept struct {
		Id    string  `json:"id"`
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		AppId string  `json:"app_id"`
	}

	Model struct {
		Name      string `json:"name"`
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		AppID     string `json:"app_id"`
	}

	ModelVersion struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		Status    Status `json:"status"`
	}

	OutputInfo struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	}

	Output struct {
		Id          string      `json:"id"`
		Status      Status      `json:"status"`
		CreatedAt   string      `json:"created_at"`
		Model       Model       `json:"model"`
		Input       Input       `json:"input"`
		ConceptData ConceptData `json:"data"`
	}

	ConceptData struct {
		Concepts []Concept `json:"concepts"`
	}

	Data struct {
		Image Image `json:"image"`
	}

	Image struct {
		Url string `json:"url"`
	}

	Input struct {
		Data Data `json:"data"`
	}

	Request struct {
		Inputs []Input `json:"inputs"`
	}

	Response struct {
		Status  Status   `json:"status"`
		Outputs []Output `json:"outputs"`
	}
)
