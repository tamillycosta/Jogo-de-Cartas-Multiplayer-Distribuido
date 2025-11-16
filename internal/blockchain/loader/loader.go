package loader

import (
	"encoding/json"
	"fmt"
	"os"
)

type TruffleArtifact struct {
	Networks map[string]struct {
		Address string `json:"address"`
	} `json:"networks"`
}

func LoadContractAddress(artifact string, network string) (string, error) {
	b, err := os.ReadFile(artifact)
	if err != nil {
		return "", err
	}

	var a TruffleArtifact
	if err := json.Unmarshal(b, &a); err != nil {
		return "", err
	}

	n, ok := a.Networks[network]
	if !ok {
		return "", fmt.Errorf("network %s não encontrada", network)
	}

	return n.Address, nil
}
