package proofscache

import "github.com/multiversx/mx-chain-core-go/data"

type proofNonceBucket struct {
	maxNonce                   uint64
	proofsByNonce              map[uint64]string
	processedHashProofsByNonce map[uint64]string
}

func newProofBucket() *proofNonceBucket {
	return &proofNonceBucket{
		proofsByNonce:              make(map[uint64]string),
		processedHashProofsByNonce: make(map[uint64]string),
	}
}

func (p *proofNonceBucket) size() int {
	return len(p.proofsByNonce)
}

func (p *proofNonceBucket) insert(proof data.HeaderProofHandler) {
	p.proofsByNonce[proof.GetHeaderNonce()] = string(proof.GetHeaderHash())
	p.processedHashProofsByNonce[proof.GetHeaderNonce()] = string(proof.GetProcessedHeaderHash())

	if proof.GetHeaderNonce() > p.maxNonce {
		p.maxNonce = proof.GetHeaderNonce()
	}
}
