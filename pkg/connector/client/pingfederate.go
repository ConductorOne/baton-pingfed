package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type PingFederateClient struct {
	baseURL     string
	client      *uhttp.BaseHttpClient
	Username    string
	Password    string
	initialized bool
}

const (
	APIPath     = "/pf-admin-api/v1"
	AuditorRole = "AUDITOR"
)

func New(
	baseURL string,
	username string,
	password string,
) *PingFederateClient {
	return &PingFederateClient{
		baseURL:  baseURL,
		Password: password,
		Username: username,
	}
}

func (c *PingFederateClient) Initialize(ctx context.Context) error {
	logger := ctxzap.Extract(ctx)
	if c.initialized {
		logger.Debug("PingFederate client already initialized")
		return nil
	}
	logger.Debug("Initializing PingFederate client")

	if c.baseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return err
	}

	c.client = uhttp.NewBaseHttpClient(httpClient)
	c.initialized = true
	return nil
}

// doRequest performs an HTTP request and handles common response processing.
func (c *PingFederateClient) doRequest(ctx context.Context, method, path string, body interface{}, response interface{}) error {
	logger := ctxzap.Extract(ctx)
	url := c.baseURL + APIPath + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("X-XSRF-Header", "PingFederate")
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	var resp *http.Response
	if response != nil {
		resp, err = c.client.Do(req, uhttp.WithJSONResponse(response))
	} else {
		resp, err = c.client.Do(req)
	}

	if err != nil {
		return fmt.Errorf("request failed: %w. status code: %d", err, resp.StatusCode)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	logger.Debug("response",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", resp.StatusCode),
		zap.String("body", string(bodyBytes)),
	)

	return nil
}

// GetUsers retrieves a list of PingFederate users from the API.
func (c *PingFederateClient) GetUsers(ctx context.Context) ([]PingFederateUser, error) {
	err := c.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	var response getAdminUsersResponse
	err = c.doRequest(ctx, http.MethodGet, "/administrativeAccounts", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return response.Items, nil
}

// GetRoles retrieves a list of PingFederate roles from the API.
func (c *PingFederateClient) GetRoles(ctx context.Context) ([]PingFederateRole, error) {
	err := c.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	var response getAdminUsersResponse
	err = c.doRequest(ctx, http.MethodGet, "/administrativeAccounts", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	roles := make([]PingFederateRole, 0)
	roleMap := make(map[string]bool)

	for _, user := range response.Items {
		for _, role := range user.Roles {
			if !roleMap[role] {
				roleMap[role] = true
				roles = append(roles, PingFederateRole{
					Name: role,
					ID:   role,
				})
			}
		}
	}

	// Adding the Auditor role.
	roles = append(roles, PingFederateRole{
		Name: AuditorRole,
		ID:   AuditorRole,
	})

	return roles, nil
}

func (c *PingFederateClient) GetRoleAssignments(ctx context.Context, roleID string) ([]PingFederateUser, error) {
	err := c.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	var response getAdminUsersResponse
	err = c.doRequest(ctx, http.MethodGet, "/administrativeAccounts", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get role assignments: %w", err)
	}

	var usersWithRole []PingFederateUser
	if c.isAuditor(roleID) {
		for _, user := range response.Items {
			if user.IsAuditor {
				usersWithRole = append(usersWithRole, user)
			}
		}
	} else {
		for _, user := range response.Items {
			for _, role := range user.Roles {
				if role == roleID {
					usersWithRole = append(usersWithRole, user)
					break
				}
			}
		}
	}

	return usersWithRole, nil
}

func (c *PingFederateClient) AddUserToRole(ctx context.Context, userID string, roleID string) error {
	err := c.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	var user PingFederateUser
	err = c.doRequest(ctx, http.MethodGet, "/administrativeAccounts/"+userID, nil, &user)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if c.isAuditor(roleID) {
		user.IsAuditor = true
	} else {
		user.Roles = append(user.Roles, roleID)
	}

	err = c.doRequest(ctx, http.MethodPut, "/administrativeAccounts/"+userID, user, nil)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (c *PingFederateClient) RemoveUserFromRole(ctx context.Context, userID string, roleID string) error {
	err := c.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	var user PingFederateUser
	err = c.doRequest(ctx, http.MethodGet, "/administrativeAccounts/"+userID, nil, &user)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if c.isAuditor(roleID) {
		user.IsAuditor = false
	} else {
		// Remove the role from the user's roles
		newRoles := make([]string, 0)
		for _, role := range user.Roles {
			if role != roleID {
				newRoles = append(newRoles, role)
			}
		}
		user.Roles = newRoles
	}

	err = c.doRequest(ctx, http.MethodPut, "/administrativeAccounts/"+userID, user, nil)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (c *PingFederateClient) isAuditor(roleID string) bool {
	return roleID == AuditorRole
}
