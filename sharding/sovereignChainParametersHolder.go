package sharding

type sovereignChainParametersHolder struct {
	*chainParametersHolder
}

func NewSovereignChainParametersHolder() (*sovereignChainParametersHolder, error) {
	return &sovereignChainParametersHolder{}, nil
}
