package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type PingFederateClient struct {
	baseUrl     string
	client      *uhttp.BaseHttpClient
	Username    string
	Password    string
	initialized bool
}

const (
	API_PATH = "/pf-admin-api/v1"
)

func New(
	baseUrl string,
	username string,
	password string,
) *PingFederateClient {
	return &PingFederateClient{
		baseUrl:  baseUrl,
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

	if c.baseUrl == "" {
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

// GetUsers retrieves a list of PingFederate users from the API
func (c *PingFederateClient) GetUsers(ctx context.Context) ([]PingFederateUser, error) {
	err := c.Initialize(ctx)
	logger := ctxzap.Extract(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	if c.client == nil {
		return nil, fmt.Errorf("client is not properly initialized")
	}

	url := c.baseUrl + API_PATH + "/administrativeAccounts"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("X-XSRF-Header", "PingFederate")
	req.Header.Set("Accept", "application/json")

	var response getAdminUsersResponse
	resp, err := c.client.Do(
		req,
		uhttp.WithJSONResponse(&response),
	)
	if resp == nil {
		return nil, fmt.Errorf("API response was nil")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admon users: %w. status code: %d", err, resp.StatusCode)
	}

	logger.Debug("response: ", zap.Any("response", resp.Body))
	return response.Items, nil
}

// GetRoles retrieves a list of PingFederate roles from the API
// The default Initial administrator user has all the roles
func (c *PingFederateClient) GetRoles(ctx context.Context) ([]PingFederateRole, error) {
	err := c.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	var response getAdminUsersResponse
	url := c.baseUrl + API_PATH + "/administrativeAccounts"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("X-XSRF-Header", "PingFederate")
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(
		req,
		uhttp.WithJSONResponse(&response),
	)
	if resp == nil {
		return nil, fmt.Errorf("API response was nil")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admon users: %w. status code: %d", err, resp.StatusCode)
	}

	logger := ctxzap.Extract(ctx)
	var stringRoles []string
	for _, user := range response.Items {
		if user.Username == "Administrator" {
			stringRoles = user.Roles
			break
		}
	}

	var roles []PingFederateRole
	for _, role := range stringRoles {
		roles = append(roles, PingFederateRole{
			Name: role,
			ID:   role,
		})
	}

	//adding the Auditor role:
	roles = append(roles, PingFederateRole{
		Name: "AUDITOR",
		ID:   "AUDITOR",
	})

	logger.Debug("response: ", zap.Any("response", resp.Body))
	return roles, nil
}

func (c *PingFederateClient) GetRoleAssignments(ctx context.Context, roleID string) ([]PingFederateUser, error) {
	err := c.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	var response getAdminUsersResponse
	url := c.baseUrl + API_PATH + "/administrativeAccounts"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("X-XSRF-Header", "PingFederate")
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(
		req,
		uhttp.WithJSONResponse(&response),
	)
	if resp == nil {
		return nil, fmt.Errorf("API response was nil")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admon users: %w. status code: %d", err, resp.StatusCode)
	}

	var usersWithRole []PingFederateUser
	//Auditor is indicated by a bool
	if roleID == "AUDITOR" {
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
					break // Found the role, no need to check other roles for this user
				}
			}
		}
	}

	logger := ctxzap.Extract(ctx)
	logger.Debug("users with role",
		zap.String("roleID", roleID),
		zap.Int("userCount", len(usersWithRole)),
	)

	return usersWithRole, nil
}

func (c *PingFederateClient) AddUserToRole(
	ctx context.Context,
	userId string,
	roleId string,
) error {
	err := c.Initialize(ctx)
	logger := ctxzap.Extract(ctx)

	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	var user PingFederateUser
	url := c.baseUrl + API_PATH + "/administrativeAccounts/" + userId

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("X-XSRF-Header", "PingFederate")
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(
		req,
		uhttp.WithJSONResponse(&user),
	)
	if resp == nil {
		return fmt.Errorf("API response was nil")
	}
	if err != nil {
		return fmt.Errorf("failed to get admon users: %w. status code: %d", err, resp.StatusCode)
	}

	if roleId == "AUDITOR" {
		user.IsAuditor = true
	} else {
		user.Roles = append(user.Roles, roleId)
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Error marshalling user")
	}

	req2, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(userJSON))
	if err != nil {
		return err
	}
	req2.SetBasicAuth(c.Username, c.Password)
	req2.Header.Set("X-XSRF-Header", "PingFederate")
	req2.Header.Set("Accept", "application/json")
	resp2, err := c.client.Do(
		req2,
	)
	if resp2 == nil {
		return fmt.Errorf("API response was nil")
	}
	if err != nil {
		return fmt.Errorf("failed to get admon users: %w. status code: %d", err, resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to add role to user: status code %d, response: %s", resp.StatusCode, resp.Body)
	}
	logger.Debug("response: ", zap.Any("response", resp.Body))
	return nil
}

func (c *PingFederateClient) RemoveUserFromRole(
	ctx context.Context,
	userId string,
	roleId string,
) error {
	err := c.Initialize(ctx)
	logger := ctxzap.Extract(ctx)

	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	var user PingFederateUser
	url := c.baseUrl + API_PATH + "/administrativeAccounts/" + userId

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("X-XSRF-Header", "PingFederate")
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(
		req,
		uhttp.WithJSONResponse(&user),
	)
	if resp == nil {
		return fmt.Errorf("API response was nil")
	}
	if err != nil {
		return fmt.Errorf("failed to get admon users: %w. status code: %d", err, resp.StatusCode)
	}

	if roleId == "AUDITOR" {
		user.IsAuditor = false
	}
	// Remove the role from the user's roles
	var newRoles []string
	for _, role := range user.Roles {
		if role != roleId {
			newRoles = append(newRoles, role)
		}
	}
	user.Roles = newRoles

	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Error marshalling user")
	}

	req2, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(userJSON))
	if err != nil {
		return err
	}
	req2.SetBasicAuth(c.Username, c.Password)
	req2.Header.Set("X-XSRF-Header", "PingFederate")
	req2.Header.Set("Accept", "application/json")
	resp2, err := c.client.Do(
		req2,
	)
	if resp2 == nil {
		return fmt.Errorf("API response was nil")
	}
	if err != nil {
		return fmt.Errorf("failed to get admon users: %w. status code: %d", err, resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to add role to user: status code %d, response: %s", resp.StatusCode, resp.Body)
	}
	logger.Debug("response: ", zap.Any("response", resp.Body))
	return nil
}
