package rule

type Document struct {
	Schema   string   `json:"$schema"`
	Policies []Policy `json:"policy-rules"`
}
