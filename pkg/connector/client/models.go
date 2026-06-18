// Package client provides the PingFederate API client.
package client

// PingFederateUser represents a PingFederate administrative account.
type PingFederateUser struct {
	Email             string   `json:"emailAddress,omitempty"`
	EncryptedPassword string   `json:"encryptedPassword"`
	Username          string   `json:"username"`
	PhoneNumber       string   `json:"phoneNumber,omitempty"`
	Department        string   `json:"department,omitempty"`
	Description       string   `json:"description"`
	IsAuditor         bool     `json:"auditor"`
	IsActive          bool     `json:"active"`
	Roles             []string `json:"roles"`
}

type getAdminUsersResponse struct {
	Items []PingFederateUser `json:"items"`
}

// PingFederateRole represents a PingFederate administrative role.
type PingFederateRole struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
