package fraud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/real-time-payments/pkg/logger"
)

type MLClient interface {
	Predict(request *MLPredictRequest) (*MLPredictResponse, error)
	Train() error
	HealthCheck() (bool, error)
}

type mlClient struct {
	baseURL    string
	httpClient *http.Client
}

type MLPredictRequest struct {
	Amount          float64 `json:"amount"`
	Timestamp       string  `json:"timestamp"`
	Velocity5Min    int     `json:"velocity_5min"`
	Velocity1Hour   int     `json:"velocity_1hour"`
	AvgAmount24h    float64 `json:"avg_amount_24h"`
	StdAmount24h    float64 `json:"std_amount_24h"`
	TimeSinceLastTx int     `json:"time_since_last_tx"`
	AmountScore     int     `json:"amount_score"`
	VelocityScore   int     `json:"velocity_score"`
	TimeScore       int     `json:"time_score"`
	PatternScore    int     `json:"pattern_score"`
}

type MLPredictResponse struct {
	Success      bool                   `json:"success"`
	MLScore      int                    `json:"ml_score"`
	IsAnomaly    bool                   `json:"is_anomaly"`
	AnomalyScore float64                `json:"anomaly_score"`
	FeaturesUsed map[string]interface{} `json:"features_used"`
	Error        string                 `json:"error,omitempty"`
}

func NewMLClient(baseURL string) MLClient {
	return &mlClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *mlClient) Predict(request *MLPredictRequest) (*MLPredictResponse, error) {
	url := fmt.Sprintf("%s/predict", c.baseURL)

	// Marshal request
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logger.Error().Err(err).Msg("ML service request failed")
		// Return default response on error (don't block transaction)
		return &MLPredictResponse{
			Success:   false,
			MLScore:   0,
			IsAnomaly: false,
			Error:     err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var mlResponse MLPredictResponse
	if err := json.Unmarshal(responseBody, &mlResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &mlResponse, nil
}

func (c *mlClient) Train() error {
	url := fmt.Sprintf("%s/train", c.baseURL)

	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to trigger training: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("training failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.Info().Msg("ML model training triggered successfully")
	return nil
}

func (c *mlClient) HealthCheck() (bool, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
