package umramonline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrRequestFailed = errors.New("umramonline request failed")

type Config struct {
	BaseURL                 string
	APIKey                  string
	APIToken                string
	OTPRequestPath          string
	OTPVerifyPath           string
	PasswordLoginPath       string
	UserRolesPath           string
	CustomersPath           string
	CustomerSearchPath      string
	CustomerPhoneExistsPath string
	ZonesPath               string
	CitiesPath              string
	TownsPath               string
	BranchesPath            string
	TaskSMSPath             string
	Timeout                 time.Duration
}

type Client struct {
	baseURL                 string
	apiKey                  string
	apiToken                string
	otpRequestPath          string
	otpVerifyPath           string
	passwordLoginPath       string
	userRolesPath           string
	customersPath           string
	customerSearchPath      string
	customerPhoneExistsPath string
	zonesPath               string
	citiesPath              string
	townsPath               string
	branchesPath            string
	taskSMSPath             string
	httpClient              *http.Client
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

type taskCreatedSMSRequest struct {
	Phone                string `json:"phone"`
	TaskUUID             string `json:"task_uuid"`
	Title                string `json:"title"`
	AssignedUserFullName string `json:"assigned_user_full_name"`
	BranchName           string `json:"branch_name"`
	VisitDate            string `json:"visit_date"`
	DueDate              string `json:"due_date"`
	Priority             string `json:"priority"`
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

func (r listResponse[T]) successful() bool {
	return r.Success
}

type Role struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type Zone struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type City struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
}

type Town struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	CityID    uint64 `json:"city_id"`
	CityTitle string `json:"city_title"`
}

type Branch struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
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
	ZoneID     int
	BranchIDs  []int32
}

type CustomerListItem struct {
	ID           uint64  `json:"id"`
	Situation    string  `json:"situation"`
	Unvan        string  `json:"unvan"`
	Cep          string  `json:"cep"`
	Ad           string  `json:"ad"`
	Soyad        string  `json:"soyad"`
	BranchName   string  `json:"branch_name"`
	ZoneName     string  `json:"zone_name"`
	PlusCardNo   string  `json:"plus_card_no"`
	Credit       string  `json:"credit"`
	Source       string  `json:"source"`
	City         string  `json:"city"`
	Town         string  `json:"town"`
	CreatedAt    *string `json:"created_at"`
	Type         string  `json:"type"`
	DaysSpending *int    `json:"daysSpending"`
	DaysLoading  *int    `json:"daysLoading"`
}

type CustomerSearchItem struct {
	ID         uint64  `json:"id"`
	UOId       uint64  `json:"uo_id"`
	BranchID   *int32  `json:"branch_id"`
	Unvan      string  `json:"unvan"`
	Ad         string  `json:"ad"`
	Soyad      string  `json:"soyad"`
	YetkiliAdi string  `json:"yetkili_adi"`
	Cep        string  `json:"cep"`
	Telefon    string  `json:"telefon"`
	Mahalle    string  `json:"mahalle"`
	IlKodu     string  `json:"il_kodu"`
	IlceKodu   string  `json:"ilce_kodu"`
	VergiNo    string  `json:"vergi_no"`
	TCNo       string  `json:"tc_no"`
	Type       string  `json:"type"`
	CreatedAt  *string `json:"created_at"`
	PlusCardNo string  `json:"plus_card_no"`
	Credit     uint64  `json:"credit"`
	Point      uint64  `json:"point"`
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
		baseURL:                 strings.TrimRight(config.BaseURL, "/"),
		apiKey:                  config.APIKey,
		apiToken:                config.APIToken,
		otpRequestPath:          "/" + strings.Trim(config.OTPRequestPath, "/"),
		otpVerifyPath:           "/" + strings.Trim(config.OTPVerifyPath, "/"),
		passwordLoginPath:       "/" + strings.Trim(config.PasswordLoginPath, "/"),
		userRolesPath:           "/" + strings.Trim(config.UserRolesPath, "/"),
		customersPath:           "/" + strings.Trim(config.CustomersPath, "/"),
		customerSearchPath:      "/" + strings.Trim(config.CustomerSearchPath, "/"),
		customerPhoneExistsPath: "/" + strings.Trim(config.CustomerPhoneExistsPath, "/"),
		zonesPath:               "/" + strings.Trim(config.ZonesPath, "/"),
		citiesPath:              "/" + strings.Trim(config.CitiesPath, "/"),
		townsPath:               "/" + strings.Trim(config.TownsPath, "/"),
		branchesPath:            "/" + strings.Trim(config.BranchesPath, "/"),
		taskSMSPath:             "/" + strings.Trim(config.TaskSMSPath, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) SendTaskCreatedSMS(
	ctx context.Context,
	phone string,
	taskUUID string,
	title string,
	assignedUserFullName string,
	branchName string,
	visitDate string,
	dueDate string,
	priority string,
) error {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.taskSMSPath == "/" {
		return ErrRequestFailed
	}

	body, err := json.Marshal(taskCreatedSMSRequest{
		Phone:                phone,
		TaskUUID:             taskUUID,
		Title:                title,
		AssignedUserFullName: assignedUserFullName,
		BranchName:           branchName,
		VisitDate:            visitDate,
		DueDate:              dueDate,
		Priority:             priority,
	})
	if err != nil {
		return ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.taskSMSPath, bytes.NewReader(body))
	if err != nil {
		return ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

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

func (c *Client) RequestOTP(ctx context.Context, phone string) error {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.otpRequestPath == "/" {
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
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

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
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.otpVerifyPath == "/" {
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
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

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
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.passwordLoginPath == "/" {
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
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

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
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.userRolesPath == "/" {
		return nil, ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+c.userRolesPath, nil)
	if err != nil {
		return nil, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

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

func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.zonesPath == "/" {
		return nil, ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+c.zonesPath, nil)
	if err != nil {
		return nil, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, ErrRequestFailed
	}
	defer response.Body.Close()

	var apiResponse listResponse[Zone]
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, ErrRequestFailed
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !apiResponse.Success {
		return nil, fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	return apiResponse.Items, nil
}

func (c *Client) SearchCustomer(ctx context.Context, query string) (CustomerSearchItem, bool, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.customerSearchPath == "/" {
		return CustomerSearchItem{}, false, ErrRequestFailed
	}

	values := url.Values{}
	setQueryString(values, "q", query)

	var apiResponse customerSearchResponse
	if err := c.getJSON(ctx, c.customerSearchPath, values, &apiResponse); err != nil {
		return CustomerSearchItem{}, false, err
	}

	if apiResponse.Data == nil {
		return CustomerSearchItem{}, false, nil
	}

	return *apiResponse.Data, true, nil
}

func (c *Client) CustomerPhoneExists(ctx context.Context, phone string) (bool, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.customerPhoneExistsPath == "/" {
		return false, ErrRequestFailed
	}

	values := url.Values{}
	setQueryString(values, "phone", phone)

	var apiResponse customerPhoneExistsResponse
	if err := c.getJSON(ctx, c.customerPhoneExistsPath, values, &apiResponse); err != nil {
		return false, err
	}

	return apiResponse.Data.Exists, nil
}

func (c *Client) ListCities(ctx context.Context) ([]City, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.citiesPath == "/" {
		return nil, ErrRequestFailed
	}

	var apiResponse listResponse[City]
	if err := c.getJSON(ctx, c.citiesPath, nil, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) ListTowns(ctx context.Context, cityID uint64) ([]Town, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.townsPath == "/" {
		return nil, ErrRequestFailed
	}

	values := url.Values{}
	if cityID > 0 {
		values.Set("city_id", strconv.FormatUint(cityID, 10))
	}

	var apiResponse listResponse[Town]
	if err := c.getJSON(ctx, c.townsPath, values, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) ListBranches(ctx context.Context) ([]Branch, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.branchesPath == "/" {
		return nil, ErrRequestFailed
	}

	var apiResponse listResponse[Branch]
	if err := c.getJSON(ctx, c.branchesPath, nil, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) GetBranch(ctx context.Context, branchID uint64) (Branch, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.branchesPath == "/" || branchID == 0 {
		return Branch{}, ErrRequestFailed
	}

	var apiResponse branchResponse
	if err := c.getJSON(ctx, c.branchesPath+"/"+strconv.FormatUint(branchID, 10), nil, &apiResponse); err != nil {
		return Branch{}, err
	}

	if apiResponse.Data == nil {
		return Branch{}, ErrRequestFailed
	}

	return *apiResponse.Data, nil
}

func (c *Client) ListBranchUsers(ctx context.Context, branchID uint64) ([]User, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.branchesPath == "/" || branchID == 0 {
		return nil, ErrRequestFailed
	}

	var apiResponse listResponse[User]
	if err := c.getJSON(ctx, c.branchesPath+"/"+strconv.FormatUint(branchID, 10)+"/users", nil, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) GetBranchUser(ctx context.Context, branchID uint64, userID uint64) (User, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.branchesPath == "/" || branchID == 0 || userID == 0 {
		return User{}, ErrRequestFailed
	}

	var apiResponse userResponse
	path := c.branchesPath + "/" + strconv.FormatUint(branchID, 10) + "/users/" + strconv.FormatUint(userID, 10)
	if err := c.getJSON(ctx, path, nil, &apiResponse); err != nil {
		return User{}, err
	}

	if apiResponse.Data == nil {
		return User{}, ErrRequestFailed
	}

	return *apiResponse.Data, nil
}

func (c *Client) ListCustomers(ctx context.Context, query CustomerListQuery) (CustomerListResult, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.customersPath == "/" {
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
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return CustomerListResult{}, ErrRequestFailed
	}
	defer response.Body.Close()

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

func (c *Client) GetCustomer(ctx context.Context, id uint64) (CustomerSearchItem, error) {
	if c.baseURL == "" || c.apiKey == "" || c.apiToken == "" || c.customersPath == "/" || id == 0 {
		return CustomerSearchItem{}, ErrRequestFailed
	}

	var apiResponse customerSearchResponse
	if err := c.getJSON(ctx, c.customersPath+"/"+strconv.FormatUint(id, 10), nil, &apiResponse); err != nil {
		return CustomerSearchItem{}, err
	}

	if apiResponse.Data == nil {
		return CustomerSearchItem{}, ErrRequestFailed
	}

	return *apiResponse.Data, nil
}

type customerListResponse struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message"`
	Items      []CustomerListItem `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

type customerSearchResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    *CustomerSearchItem `json:"data"`
}

func (r customerSearchResponse) successful() bool {
	return r.Success
}

type branchResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Data    *Branch `json:"data"`
}

func (r branchResponse) successful() bool {
	return r.Success
}

type userResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *User  `json:"data"`
}

func (r userResponse) successful() bool {
	return r.Success
}

type customerPhoneExistsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Exists bool `json:"exists"`
	} `json:"data"`
}

func (r customerPhoneExistsResponse) successful() bool {
	return r.Success
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
	setQueryInt(values, "zone_id", query.ZoneID)
	for _, branchID := range query.BranchIDs {
		if branchID > 0 {
			values.Add("branch_ids[]", strconv.FormatInt(int64(branchID), 10))
		}
	}

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

func (c *Client) getJSON(ctx context.Context, path string, values url.Values, target any) error {
	requestURL := c.baseURL + path
	if encoded := values.Encode(); encoded != "" {
		requestURL += "?" + encoded
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return ErrRequestFailed
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return ErrRequestFailed
	}

	successfulResponse, ok := target.(interface{ successful() bool })
	if ok && !successfulResponse.successful() {
		return fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	return nil
}
