package cyberark

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// UserListResponse represents the response from CyberArk's list users endpoint
type UserListResponse struct {
	Users []User `json:"Users"`
	Total int    `json:"Total"`
}

// User represents a CyberArk user
type User struct {
	ID                    int              `json:"id"`
	Username              string           `json:"username"`
	Source                string           `json:"source"`
	UserType              string           `json:"userType"`
	ComponentUser         bool             `json:"componentUser"`
	GroupsMembership      []GroupMembership `json:"groupsMembership,omitempty"`
	VaultAuthorization    []string         `json:"vaultAuthorization,omitempty"`
	Location              string           `json:"location,omitempty"`
	PersonalDetails       *PersonalDetails `json:"personalDetails,omitempty"`
	EnableUser            bool             `json:"enableUser"`
	Suspended             bool             `json:"suspended"`
	LastSuccessfulLoginAt *int64           `json:"lastSuccessfulLoginDate,omitempty"`
	UnAuthorizedInterfaces []string        `json:"unAuthorizedInterfaces,omitempty"`
	AuthenticationMethod  []string         `json:"authenticationMethod,omitempty"`
	PasswordNeverExpires  bool             `json:"passwordNeverExpires"`
	DistinguishedName     string           `json:"distinguishedName,omitempty"`
	Description           string           `json:"description,omitempty"`
	Internet              *InternetAccess  `json:"internet,omitempty"`
	BusinessAddress       *BusinessAddress `json:"businessAddress,omitempty"`
	Phones                []Phone          `json:"phones,omitempty"`
}

// PersonalDetails represents user's personal information
type PersonalDetails struct {
	FirstName            string `json:"firstName,omitempty"`
	MiddleName           string `json:"middleName,omitempty"`
	LastName             string `json:"lastName,omitempty"`
	Organization         string `json:"organization,omitempty"`
	Department           string `json:"department,omitempty"`
	Profession           string `json:"profession,omitempty"`
	Street               string `json:"street,omitempty"`
	City                 string `json:"city,omitempty"`
	State                string `json:"state,omitempty"`
	ZipCode              string `json:"zipCode,omitempty"`
	Country              string `json:"country,omitempty"`
	Title                string `json:"title,omitempty"`
	HomeNumber           string `json:"homeNumber,omitempty"`
	HomeFax              string `json:"homeFax,omitempty"`
	CellularNumber       string `json:"cellularNumber,omitempty"`
	PagerNumber          string `json:"pagerNumber,omitempty"`
	Notes                string `json:"notes,omitempty"`
	ChangePassOnNextLogon bool   `json:"changePassOnNextLogon,omitempty"`
	ExpiryDate           *int64 `json:"expiryDate,omitempty"`
}

// GroupMembership represents a user's group membership
type GroupMembership struct {
	GroupID   int    `json:"groupId"`
	GroupName string `json:"groupName"`
	GroupType string `json:"groupType"`
}

// InternetAccess represents user's internet access details
type InternetAccess struct {
	Email         string `json:"email,omitempty"`
	HomeEmail     string `json:"homeEmail,omitempty"`
	HomePage      string `json:"homePage,omitempty"`
	OtherEmail    string `json:"otherEmail,omitempty"`
}

// BusinessAddress represents user's business address
type BusinessAddress struct {
	WorkStreet  string `json:"workStreet,omitempty"`
	WorkCity    string `json:"workCity,omitempty"`
	WorkState   string `json:"workState,omitempty"`
	WorkZip     string `json:"workZip,omitempty"`
	WorkCountry string `json:"workCountry,omitempty"`
}

// Phone represents a phone number entry
type Phone struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
	PhoneType   string `json:"phoneType,omitempty"`
}

// ListUsersOptions represents the options for listing users
type ListUsersOptions struct {
	PageOffset int    // Page number (1-based)
	PageSize   int    // Number of users per page
	Filter     string // Optional filter
	Sort       string // Optional sort
	ExtendedDetails bool // Whether to include extended user details
}

// ListUsers retrieves users from CyberArk with pagination
func (c *Client) ListUsers(ctx context.Context, opts ListUsersOptions) (*UserListResponse, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("client not authenticated")
	}

	// Build query parameters
	params := url.Values{}
	params.Set("pageSize", strconv.Itoa(opts.PageSize))
	params.Set("pageOffset", strconv.Itoa(opts.PageOffset))
	
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}
	
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	
	if opts.ExtendedDetails {
		params.Set("extendedDetails", "true")
	}

	// Build the URL
	// Ensure base URL has trailing slash
	baseURL := c.baseURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	reqURL := fmt.Sprintf("%sAPI/Users?%s", baseURL, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	httpClient := c.getHTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("authentication failed or token expired")
		case http.StatusForbidden:
			return nil, fmt.Errorf("insufficient permissions to list users")
		default:
			return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
		}
	}

	// Parse response
	var result UserListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Helper function to convert CyberArk timestamp to time.Time
func TimestampToTime(timestamp *int64) *time.Time {
	if timestamp == nil || *timestamp == 0 {
		return nil
	}
	// CyberArk timestamps are typically in seconds
	t := time.Unix(*timestamp, 0)
	return &t
}