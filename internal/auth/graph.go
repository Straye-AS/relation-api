package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/straye-as/relation-api/internal/config"
)

// GraphClient fetches user profile data from Microsoft Graph API
type GraphClient struct {
	httpClient *http.Client
	config     *config.AzureAdConfig
}

// GraphUserProfile contains user profile data from Microsoft Graph
type GraphUserProfile struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"displayName"`
	GivenName         string   `json:"givenName"`
	Surname           string   `json:"surname"`
	Mail              string   `json:"mail"`
	UserPrincipalName string   `json:"userPrincipalName"`
	JobTitle          string   `json:"jobTitle"`
	Department        string   `json:"department"`
	OfficeLocation    string   `json:"officeLocation"`
	MobilePhone       string   `json:"mobilePhone"`
	BusinessPhones    []string `json:"businessPhones"`
}

// tokenResponse represents the OAuth2 token response
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// NewGraphClient creates a new Microsoft Graph API client
func NewGraphClient(cfg *config.AzureAdConfig) *GraphClient {
	return &GraphClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		config: cfg,
	}
}

// GetUserProfile fetches the user's profile from Microsoft Graph
// Uses the On-Behalf-Of (OBO) flow to exchange the user's API token for a Graph token
func (c *GraphClient) GetUserProfile(ctx context.Context, userAccessToken string) (*GraphUserProfile, error) {
	// Check if we have the required config for OBO flow
	if c.config.ClientSecret == "" {
		return nil, fmt.Errorf("client secret not configured - cannot use OBO flow")
	}

	// Exchange the user's token for a Graph token using OBO flow
	graphToken, err := c.exchangeTokenForGraph(ctx, userAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Call Graph API with the new token
	return c.callGraphAPI(ctx, graphToken)
}

// exchangeTokenForGraph uses the On-Behalf-Of flow to get a Graph API token
func (c *GraphClient) exchangeTokenForGraph(ctx context.Context, userToken string) (string, error) {
	tokenURL := fmt.Sprintf("%s%s/oauth2/v2.0/token", c.config.InstanceUrl, c.config.TenantId)

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("client_id", c.config.ClientId)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("assertion", userToken)
	data.Set("scope", "https://graph.microsoft.com/User.Read")
	data.Set("requested_token_use", "on_behalf_of")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.ErrorDescription != "" {
			return "", fmt.Errorf("token exchange failed (%d): %s - %s", resp.StatusCode, errorResp.Error, errorResp.ErrorDescription)
		}
		return "", fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// callGraphAPI calls the Microsoft Graph API to get user profile
func (c *GraphClient) callGraphAPI(ctx context.Context, accessToken string) (*GraphUserProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me?$select=id,displayName,givenName,surname,mail,userPrincipalName,jobTitle,department,officeLocation,mobilePhone,businessPhones", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error.Message != "" {
			return nil, fmt.Errorf("Graph API error (%d): %s - %s", resp.StatusCode, errorResp.Error.Code, errorResp.Error.Message)
		}
		return nil, fmt.Errorf("Graph API returned status %d", resp.StatusCode)
	}

	var profile GraphUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode Graph API response: %w", err)
	}

	return &profile, nil
}
