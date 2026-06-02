package userclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/domain/port"
)

type httpClient struct {
	client  *http.Client
	baseURL string
}

func New(baseURL string) port.UserServiceClient {
	return &httpClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type createProfileRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func (c *httpClient) CreateProfile(ctx context.Context, userID, email, role string) error {
	body, _ := json.Marshal(createProfileRequest{
		UserID: userID,
		Email:  email,
		Role:   role,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/internal/users/profile", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", "internal-secret-change-in-production")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("user service returned %d", resp.StatusCode)
	}
	return nil
}
