package rule

type Policy struct {
	ID           string       `json:"policy-id"`
	Description  string       `json:"description"`
	Repositories []Repository `json:"repositories"`
	Conditions   []Condition  `json:"conditions"`
}
