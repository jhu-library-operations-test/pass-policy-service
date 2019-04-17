package rule

type Condition map[string]interface{}

func (c Condition) apply(variables VariableResolver) (bool, error) {
	if c == nil {
		return true, nil
	}
	return false, nil
}
