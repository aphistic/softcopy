package v1

type fileItem struct {
	ID           string    `json:"id"`
	Hash         *fileHash `json:"hash"`
	Filename     string    `json:"filename"`
	DocumentDate string    `json:"document_date"`
	Size         float64   `json:"size"`

	Tags []string `json:"tags"`
}

type fileHash struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
