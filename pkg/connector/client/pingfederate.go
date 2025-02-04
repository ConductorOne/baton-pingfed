package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

type PingFederateClient struct {
	baseURL  string
	client   *uhttp.BaseHttpClient
	Username string
	Password string
}

const (
	APIPath     = "/pf-admin-api/v1"
	AuditorRole = "AUDITOR"
)

func New(
	ctx context.Context,
	baseURL string,
	username string,
	password string,
) (*PingFederateClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, err
	}

	client, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return &PingFederateClient{
		baseURL:  baseURL,
		Password: password,
		Username: username,
		client:   client,
	}, nil
}

// doRequest performs an HTTP request and handles common response processing.
func (c *PingFederateClient) doRequest(ctx context.Context, method, path string, body interface{}, response interface{}) error {
	// logger := ctxzap.Extract(ctx)
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	u = u.JoinPath(APIPath, path)

	reqOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithHeader("X-XSRF-Header", "PingFederate"),
	}
	if body != nil {
		reqOpts = append(reqOpts, uhttp.WithJSONBody(body), uhttp.WithContentTypeJSONHeader())
	}

	req, err := c.client.NewRequest(ctx, method, u, reqOpts...)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)

	doOpts := []uhttp.DoOption{}
	if response != nil {
		doOpts = append(doOpts, uhttp.WithJSONResponse(response))
	}

	resp, err := c.client.Do(req, doOpts...)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}

// GetUsers retrieves a list of PingFederate users from the API.
func (c *PingFederateClient) GetUsers(ctx context.Context) ([]PingFederateUser, error) {
	var response getAdminUsersResponse
	err := c.doRequest(ctx, http.MethodGet, "/administrativeAccounts", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return response.Items, nil
}

// GetRoles retrieves a list of PingFederate roles from the API.
func (c *PingFederateClient) GetRoles(ctx context.Context) ([]PingFederateRole, error) {
	var response getAdminUsersResponse
	err := c.doRequest(ctx, http.MethodGet, "/administrativeAccounts", nil, &response)
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
	var response getAdminUsersResponse
	err := c.doRequest(ctx, http.MethodGet, "/administrativeAccounts", nil, &response)
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
	var user PingFederateUser
	err := c.doRequest(ctx, http.MethodGet, "/administrativeAccounts/"+userID, nil, &user)
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
	var user PingFederateUser
	err := c.doRequest(ctx, http.MethodGet, "/administrativeAccounts/"+userID, nil, &user)
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
