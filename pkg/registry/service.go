package registry

import (
	"encoding/json"
)

type Service struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Endpoint string `json:"endpoint"`
}

func Marshal(s *Service) (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Unmarshal(data []byte) (s *Service, err error) {
	err = json.Unmarshal(data, &s)
	return
}
