package umramonline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrRequestFailed = errors.New("umramonline request failed")

type Config struct {
	BaseURL           string
	APIKey            string
	OTPRequestPath    string
	OTPVerifyPath     string
	PasswordLoginPath string
	UserRolesPath     string
	CustomersPath     string
	Timeout           time.Duration
}

type Client struct {
	baseURL           string
	apiKey            string
	otpRequestPath    string
	otpVerifyPath     string
	passwordLoginPath string
	userRolesPath     string
	customersPath     string
	httpClient        *http.Client
}

type otpRequest struct {
	Phone string `json:"phone"`
}

type otpVerifyRequest struct {
	Phone   string `json:"phone"`
	OTPCode string `json:"otp_code"`
}

type passwordLoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type apiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type listResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Items   []T    `json:"items"`
}

type Role struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	RoleID   uint64 `json:"role_id"`
	RoleName string `json:"role_name"`
}

type CustomerListQuery struct {
	Page       int
	PerPage    int
	Situation  string
	Unvan      string
	Cep        string
	Ad         string
	Soyad      string
	BranchName string
	PlusCardNo string
	Source     string
	City       string
	Town       string
	CreatedAt  string
	Type       string
	SortBy     string
	SortOrder  string
}

type CustomerListItem struct {
	Situation    string     `json:"situation"`
	Unvan        string     `json:"unvan"`
	Cep          string     `json:"cep"`
	Ad           string     `json:"ad"`
	Soyad        string     `json:"soyad"`
	BranchName   string     `json:"branch_name"`
	PlusCardNo   string     `json:"plus_card_no"`
	Credit       float64    `json:"credit"`
	Source       string     `json:"source"`
	City         string     `json:"city"`
	Town         string     `json:"town"`
	CreatedAt    *time.Time `json:"created_at"`
	Type         string     `json:"type"`
	DaysSpending *int       `json:"daysSpending"`
	DaysLoading  *int       `json:"daysLoading"`
}

type Pagination struct {
	CurrentPage int  `json:"current_page"`
	LastPage    int  `json:"last_page"`
	PerPage     int  `json:"per_page"`
	Total       int  `json:"total"`
	From        *int `json:"from"`
	To          *int `json:"to"`
}

type CustomerListResult struct {
	Items      []CustomerListItem
	Pagination Pagination
}

func NewClient(config Config) *Client {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		baseURL:           strings.TrimRight(config.BaseURL, "/"),
		apiKey:            config.APIKey,
		otpRequestPath:    "/" + strings.Trim(config.OTPRequestPath, "/"),
		otpVerifyPath:     "/" + strings.Trim(config.OTPVerifyPath, "/"),
		passwordLoginPath: "/" + strings.Trim(config.PasswordLoginPath, "/"),
		userRolesPath:     "/" + strings.Trim(config.UserRolesPath, "/"),
		customersPath:     "/" + strings.Trim(config.CustomersPath, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) RequestOTP(ctx context.Context, phone string) error {
	if c.baseURL == "" || c.apiKey == "" || c.otpRequestPath == "/" {
		return ErrRequestFailed
	}

	body, err := json.Marshal(otpRequest{Phone: phone})
	if err != nil {
		return ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.otpRequestPath, bytes.NewReader(body))
	if err != nil {
		return ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return ErrRequestFailed
	}
	defer response.Body.Close()

	var apiResponse apiResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return ErrRequestFailed
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !apiResponse.Success {
		return fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	return nil
}

func (c *Client) VerifyOTP(ctx context.Context, phone string, otpCode string) (bool, error) {
	if c.baseURL == "" || c.apiKey == "" || c.otpVerifyPath == "/" {
		return false, ErrRequestFailed
	}

	body, err := json.Marshal(otpVerifyRequest{Phone: phone, OTPCode: otpCode})
	if err != nil {
		return false, ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.otpVerifyPath, bytes.NewReader(body))
	if err != nil {
		return false, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return false, ErrRequestFailed
	}
	defer response.Body.Close()

	var apiResponse apiResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return false, ErrRequestFailed
	}

	if response.StatusCode == http.StatusUnprocessableEntity {
		return false, nil
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return false, fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	return apiResponse.Success, nil
}

func (c *Client) LoginWithPassword(ctx context.Context, phone string, password string) (map[string]any, error) {
	if c.baseURL == "" || c.apiKey == "" || c.passwordLoginPath == "/" {
		return nil, ErrRequestFailed
	}

	body, err := json.Marshal(passwordLoginRequest{Phone: phone, Password: password})
	if err != nil {
		return nil, ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.passwordLoginPath, bytes.NewReader(body))
	if err != nil {
		return nil, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, ErrRequestFailed
	}
	defer response.Body.Close()

	var apiResponse apiResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, ErrRequestFailed
	}

	if response.StatusCode == http.StatusUnprocessableEntity {
		return nil, nil
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	if !apiResponse.Success {
		return nil, nil
	}

	if len(apiResponse.Data) == 0 {
		return map[string]any{}, nil
	}

	var data map[string]any
	if err := json.Unmarshal(apiResponse.Data, &data); err != nil {
		return nil, ErrRequestFailed
	}

	return data, nil
}

func (c *Client) ListRoles(ctx context.Context) ([]Role, error) {
	if c.baseURL == "" || c.apiKey == "" || c.userRolesPath == "/" {
		return nil, ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+c.userRolesPath, nil)
	if err != nil {
		return nil, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, ErrRequestFailed
	}
	defer response.Body.Close()

	var apiResponse listResponse[Role]
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, ErrRequestFailed
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !apiResponse.Success {
		return nil, fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	return apiResponse.Items, nil
}

func (c *Client) ListCustomers(ctx context.Context, query CustomerListQuery) (CustomerListResult, error) {
	if c.baseURL == "" || c.apiKey == "" || c.customersPath == "/" {
		return CustomerListResult{}, ErrRequestFailed
	}

	requestURL := c.baseURL + c.customersPath
	if encoded := customerListQueryValues(query).Encode(); encoded != "" {
		requestURL += "?" + encoded
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return CustomerListResult{}, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return CustomerListResult{}, ErrRequestFailed
	}
	defer response.Body.Close()

	fmt.Println("response", response)
	fmt.Println("response.Body", response.Body)
	bodyx, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("read error:", err)
	}

	fmt.Println("response:", string(bodyx))

	var apiResponse customerListResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return CustomerListResult{}, ErrRequestFailed
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !apiResponse.Success {
		return CustomerListResult{}, fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	return CustomerListResult{
		Items:      apiResponse.Items,
		Pagination: apiResponse.Pagination,
	}, nil
}

type customerListResponse struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message"`
	Items      []CustomerListItem `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

func customerListQueryValues(query CustomerListQuery) url.Values {
	values := url.Values{}

	setQueryInt(values, "page", query.Page)
	setQueryInt(values, "per_page", query.PerPage)
	setQueryString(values, "situation", query.Situation)
	setQueryString(values, "unvan", query.Unvan)
	setQueryString(values, "cep", query.Cep)
	setQueryString(values, "ad", query.Ad)
	setQueryString(values, "soyad", query.Soyad)
	setQueryString(values, "branch_name", query.BranchName)
	setQueryString(values, "plus_card_no", query.PlusCardNo)
	setQueryString(values, "source", query.Source)
	setQueryString(values, "city", query.City)
	setQueryString(values, "town", query.Town)
	setQueryString(values, "created_at", query.CreatedAt)
	setQueryString(values, "type", query.Type)
	setQueryString(values, "sort_by", query.SortBy)
	setQueryString(values, "sort_order", query.SortOrder)

	return values
}

func setQueryInt(values url.Values, key string, value int) {
	if value > 0 {
		values.Set(key, strconv.Itoa(value))
	}
}

func setQueryString(values url.Values, key string, value string) {
	if strings.TrimSpace(value) != "" {
		values.Set(key, strings.TrimSpace(value))
	}
}
