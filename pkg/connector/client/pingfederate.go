package client

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type PingFederateClient struct {
	baseUrl     string
	client      *resty.Client
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

	restyClient := resty.New()
	restyClient.SetBaseURL(c.baseUrl)
	restyClient.SetHeader("X-XSRF-Header", "PingFederate")
	restyClient.SetHeader("Accept", "application/json")

	auth := c.Username + ":" + c.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	restyClient.SetHeader("Authorization", authHeader)

	c.client = restyClient
	c.initialized = true
	return nil
}

// GetUsers retrieves a list of PingFederate users from the API
func (c *PingFederateClient) GetUsers(ctx context.Context) ([]PingFederateUser, error) {
	err := c.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	var response getAdminUsersResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(c.baseUrl + API_PATH + "/administrativeAccounts")

	if err != nil {
		return nil, fmt.Errorf("failed to get admin users: %w", err)
	}
	logger := ctxzap.Extract(ctx)
	logger.Debug("response: ", zap.Any("response", resp))
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
	_, err = c.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(c.baseUrl + API_PATH + "/administrativeAccounts")

	if err != nil {
		return nil, fmt.Errorf("failed to get admin users: %w", err)
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

	logger.Info("roles: ", zap.Any("roles", roles))
	return roles, nil
}

func (c *PingFederateClient) GetRoleAssignments(ctx context.Context, roleID string) ([]PingFederateUser, error) {
	err := c.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	var response getAdminUsersResponse
	_, err = c.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(c.baseUrl + API_PATH + "/administrativeAccounts")

	if err != nil {
		return nil, fmt.Errorf("failed to get admin users: %w", err)
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
	_, err = c.client.R().
		SetContext(ctx).
		SetResult(&user).
		Get(c.baseUrl + API_PATH + "/administrativeAccounts/" + userId)

	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if roleId == "AUDITOR" {
		user.IsAuditor = true
	} else {
		user.Roles = append(user.Roles, roleId)
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(user).
		Put(c.baseUrl + API_PATH + "/administrativeAccounts/" + userId)

	if err != nil {
		return fmt.Errorf("failed to add role to user: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add role to user: status code %d, response: %s", resp.StatusCode(), resp.String())
	}
	logger.Debug("response: ", zap.Any("response", resp))
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
	_, err = c.client.R().
		SetContext(ctx).
		SetResult(&user).
		Get(c.baseUrl + API_PATH + "/administrativeAccounts/" + userId)

	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
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

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(user).
		Put(c.baseUrl + API_PATH + "/administrativeAccounts/" + userId)

	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to remove role from user: status code %d, response: %s", resp.StatusCode(), resp.String())
	}
	logger.Debug("response: ", zap.Any("response", resp))
	return nil
}
