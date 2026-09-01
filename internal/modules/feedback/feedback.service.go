package feedback

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const defaultPriority = "low"

var (
	ErrInvalidType    = errors.New("invalid feedback type")
	ErrRelayHubFailed = errors.New("failed to send feedback")
)

var allowedTypes = map[string]bool{
	"report":          true,
	"feature_request": true,
	"feedback":        true,
}

type RelayHubConfig struct {
	BaseURL string
	APIKey  string
}

type Service struct {
	cfg    RelayHubConfig
	client *http.Client
}

func NewService(cfg RelayHubConfig) *Service {
	return &Service{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *Service) Create(in FeedbackInput, reporterName, reporterEmail, userID string) (*FeedbackResponse, error) {
	if !allowedTypes[in.Type] {
		return nil, ErrInvalidType
	}

	payload, err := json.Marshal(toRelayHubRequest(in, reporterName, reporterEmail, userID))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.cfg.BaseURL+"/api/v1/items", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, ErrRelayHubFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, ErrRelayHubFailed
	}

	var out relayHubItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, ErrRelayHubFailed
	}

	res := toFeedbackResponse(out)
	return &res, nil
}
