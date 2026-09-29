package umramonline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	authapp "github.com/umran/new.crm/backend/internal/auth/application"
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
	DashboardVehicleEntryPath string
	DashboardTotalAmountPath  string
	DashboardLoadedCreditPath string
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
	dashboardVehicleEntryPath string
	dashboardTotalAmountPath  string
	dashboardLoadedCreditPath string
	httpClient              *http.Client
}

type adminLoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type adminLoginVerifyRequest struct {
	MFAToken string `json:"mfa_token"`
	OTPCode  string `json:"otp_code"`
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
	BranchName string
	ZoneName   string
	PlusCardNo string
	Cep        string
	City       string
	Town       string
	SortBy     string
	SortOrder  string
	ZoneID     int
	BranchIDs  []int32
	IDs        []uint64
}

type CustomerListItem struct {
	ID         uint64 `json:"id"`
	Situation  string `json:"situation"`
	BranchName string `json:"branch_name"`
	ZoneName   string `json:"zone_name"`
	PlusCardNo string `json:"plus_card_no"`
	Cep        string `json:"cep"`
	Credit     int64  `json:"credit"`
	Point      int64  `json:"point"`
	City       string `json:"city"`
	Town       string `json:"town"`
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

type DashboardStatsQuery struct {
	StartDate        time.Time
	EndDate          time.Time
	BranchIDs        []uint64
	AllowAllBranches bool
}

type dashboardCountResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Count int64 `json:"count"`
	} `json:"data"`
}

func (r dashboardCountResponse) successful() bool {
	return r.Success
}

type dashboardAmountResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Amount float64 `json:"amount"`
	} `json:"data"`
}

func (r dashboardAmountResponse) successful() bool {
	return r.Success
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
		dashboardVehicleEntryPath: "/" + strings.Trim(config.DashboardVehicleEntryPath, "/"),
		dashboardTotalAmountPath:  "/" + strings.Trim(config.DashboardTotalAmountPath, "/"),
		dashboardLoadedCreditPath: "/" + strings.Trim(config.DashboardLoadedCreditPath, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) bearerToken(ctx context.Context) string {
	if token := TokenFromContext(ctx); token != "" {
		return token
	}

	return c.apiToken
}

func (c *Client) canCallAuthenticated(ctx context.Context) bool {
	return c.baseURL != "" && c.apiKey != "" && c.bearerToken(ctx) != ""
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
	if !c.canCallAuthenticated(ctx) || c.taskSMSPath == "/" {
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
	request.Header.Set("Authorization", "Bearer "+c.bearerToken(ctx))

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

func (c *Client) canCallAdminAuth(path string) bool {
	return c.baseURL != "" && c.apiKey != "" && path != "" && path != "/"
}

func (c *Client) newAdminAuthRequest(ctx context.Context, method string, path string, body []byte) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)

	return request, nil
}

func loginDataFromAdminAuthPayload(payload map[string]any) (map[string]any, error) {
	user, ok := payload["user"].(map[string]any)
	if !ok || user == nil {
		return nil, ErrRequestFailed
	}

	data := map[string]any{"user": user}
	if token, ok := payload["token"].(string); ok && token != "" {
		data["token"] = token
	}

	return data, nil
}

func (c *Client) AdminLogin(ctx context.Context, phone string, password string) (authapp.RequestOTPResult, error) {
	if !c.canCallAdminAuth(c.otpRequestPath) {
		return authapp.RequestOTPResult{}, ErrRequestFailed
	}

	body, err := json.Marshal(adminLoginRequest{Phone: phone, Password: password})
	if err != nil {
		return authapp.RequestOTPResult{}, ErrRequestFailed
	}

	request, err := c.newAdminAuthRequest(ctx, http.MethodPost, c.otpRequestPath, body)
	if err != nil {
		return authapp.RequestOTPResult{}, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return authapp.RequestOTPResult{}, ErrRequestFailed
	}
	defer response.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return authapp.RequestOTPResult{}, ErrRequestFailed
	}

	if response.StatusCode == http.StatusUnprocessableEntity {
		return authapp.RequestOTPResult{}, authapp.ErrPasswordRejected
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return authapp.RequestOTPResult{}, fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	if mfaRequired, ok := payload["mfa_required"].(bool); ok && mfaRequired {
		mfaToken, _ := payload["mfa_token"].(string)
		mfaChannel, _ := payload["mfa_channel"].(string)

		return authapp.RequestOTPResult{
			MFARequired: true,
			MFAToken:    mfaToken,
			MFAChannel:  mfaChannel,
		}, nil
	}

	loginData, err := loginDataFromAdminAuthPayload(payload)
	if err != nil {
		return authapp.RequestOTPResult{}, err
	}

	return authapp.RequestOTPResult{LoginData: loginData}, nil
}

func (c *Client) AdminLoginVerify(ctx context.Context, mfaToken string, otpCode string) (map[string]any, error) {
	if !c.canCallAdminAuth(c.otpVerifyPath) {
		return nil, ErrRequestFailed
	}

	body, err := json.Marshal(adminLoginVerifyRequest{MFAToken: mfaToken, OTPCode: otpCode})
	if err != nil {
		return nil, ErrRequestFailed
	}

	request, err := c.newAdminAuthRequest(ctx, http.MethodPost, c.otpVerifyPath, body)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, ErrRequestFailed
	}
	defer response.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, ErrRequestFailed
	}

	if response.StatusCode == http.StatusUnprocessableEntity {
		return nil, authapp.ErrOTPVerifyRejected
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: status=%d", ErrRequestFailed, response.StatusCode)
	}

	return loginDataFromAdminAuthPayload(payload)
}

func (c *Client) ListRoles(ctx context.Context) ([]Role, error) {
	if !c.canCallAuthenticated(ctx) || c.userRolesPath == "/" {
		return nil, ErrRequestFailed
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+c.userRolesPath, nil)
	if err != nil {
		return nil, ErrRequestFailed
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", c.apiKey)
	request.Header.Set("Authorization", "Bearer "+c.bearerToken(ctx))

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

func (c *Client) ListZones(ctx context.Context, branchIDs []uint64) ([]Zone, error) {
	if !c.canCallAuthenticated(ctx) || c.zonesPath == "/" {
		return nil, ErrRequestFailed
	}

	values := url.Values{}
	for _, branchID := range branchIDs {
		if branchID > 0 {
			values.Add("branch_ids[]", strconv.FormatUint(branchID, 10))
		}
	}

	var apiResponse listResponse[Zone]
	if err := c.getJSON(ctx, c.zonesPath, values, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) SearchCustomer(ctx context.Context, query string) (CustomerSearchItem, bool, error) {
	if !c.canCallAuthenticated(ctx) || c.customerSearchPath == "/" {
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
	if !c.canCallAuthenticated(ctx) || c.customerPhoneExistsPath == "/" {
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
	if !c.canCallAuthenticated(ctx) || c.citiesPath == "/" {
		return nil, ErrRequestFailed
	}

	var apiResponse listResponse[City]
	if err := c.getJSON(ctx, c.citiesPath, nil, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) ListTowns(ctx context.Context, cityID uint64) ([]Town, error) {
	if !c.canCallAuthenticated(ctx) || c.townsPath == "/" {
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

func (c *Client) ListBranches(ctx context.Context, branchIDs []uint64) ([]Branch, error) {
	if !c.canCallAuthenticated(ctx) || c.branchesPath == "/" {
		return nil, ErrRequestFailed
	}

	values := url.Values{}
	for _, branchID := range branchIDs {
		if branchID > 0 {
			values.Add("branch_ids[]", strconv.FormatUint(branchID, 10))
		}
	}

	var apiResponse listResponse[Branch]
	if err := c.getJSON(ctx, c.branchesPath, values, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) GetBranch(ctx context.Context, branchID uint64) (Branch, error) {
	if !c.canCallAuthenticated(ctx) || c.branchesPath == "/" || branchID == 0 {
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
	if !c.canCallAuthenticated(ctx) || c.branchesPath == "/" || branchID == 0 {
		return nil, ErrRequestFailed
	}

	var apiResponse listResponse[User]
	if err := c.getJSON(ctx, c.branchesPath+"/"+strconv.FormatUint(branchID, 10)+"/users", nil, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Items, nil
}

func (c *Client) GetBranchUser(ctx context.Context, branchID uint64, userID uint64) (User, error) {
	if !c.canCallAuthenticated(ctx) || c.branchesPath == "/" || branchID == 0 || userID == 0 {
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
	if !c.canCallAuthenticated(ctx) || c.customersPath == "/" {
		return CustomerListResult{}, ErrRequestFailed
	}

	if len(query.IDs) > 0 {
		return c.listCustomersByIDs(ctx, query)
	}

	return c.listCustomersPage(ctx, query)
}

func (c *Client) listCustomersPage(ctx context.Context, query CustomerListQuery) (CustomerListResult, error) {
	var apiResponse adminCustomerListResponse
	if err := c.getJSON(ctx, c.customersPath, customerListQueryValues(query), &apiResponse); err != nil {
		return CustomerListResult{}, err
	}

	items := make([]CustomerListItem, 0, len(apiResponse.Items))
	for _, item := range apiResponse.Items {
		items = append(items, mapAdminCustomerListItem(item))
	}

	return CustomerListResult{
		Items:      items,
		Pagination: apiResponse.Pagination,
	}, nil
}

func (c *Client) listCustomersByIDs(ctx context.Context, query CustomerListQuery) (CustomerListResult, error) {
	wantIDs := uniquePositiveIDs(query.IDs)
	if len(wantIDs) == 0 {
		return paginateCustomerListItems(nil, query.Page, query.PerPage), nil
	}

	wantSet := make(map[uint64]struct{}, len(wantIDs))
	for _, id := range wantIDs {
		wantSet[id] = struct{}{}
	}

	found := make(map[uint64]CustomerListItem, len(wantIDs))
	page := 1
	lastPage := 1
	perPage := 500

	for page <= lastPage && len(found) < len(wantSet) {
		batch, err := c.listCustomersPage(ctx, CustomerListQuery{
			Page:      page,
			PerPage:   perPage,
			SortBy:    query.SortBy,
			SortOrder: query.SortOrder,
		})
		if err != nil {
			return CustomerListResult{}, err
		}

		for _, item := range batch.Items {
			if _, ok := wantSet[item.ID]; ok {
				found[item.ID] = item
			}
		}

		if batch.Pagination.LastPage <= 0 {
			break
		}

		lastPage = batch.Pagination.LastPage
		page++
	}

	ordered := make([]CustomerListItem, 0, len(wantIDs))
	for _, id := range wantIDs {
		if item, ok := found[id]; ok {
			ordered = append(ordered, item)
		}
	}

	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	sortOrder := strings.ToLower(strings.TrimSpace(query.SortOrder))
	if sortBy == "credit" || sortBy == "point" {
		sortCustomerListItems(ordered, sortBy, sortOrder)
	}

	return paginateCustomerListItems(ordered, query.Page, query.PerPage), nil
}

func (c *Client) GetCustomer(ctx context.Context, id uint64) (CustomerSearchItem, error) {
	if !c.canCallAuthenticated(ctx) || c.customersPath == "/" || id == 0 {
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

func (c *Client) DashboardVehicleEntryCount(ctx context.Context, query DashboardStatsQuery) (int64, error) {
	return c.dashboardCount(ctx, c.dashboardVehicleEntryPath, query)
}

func (c *Client) DashboardTotalAmount(ctx context.Context, query DashboardStatsQuery) (float64, error) {
	return c.dashboardAmount(ctx, c.dashboardTotalAmountPath, query)
}

func (c *Client) DashboardLoadedCredit(ctx context.Context, query DashboardStatsQuery) (float64, error) {
	return c.dashboardAmount(ctx, c.dashboardLoadedCreditPath, query)
}

func (c *Client) dashboardCount(ctx context.Context, path string, query DashboardStatsQuery) (int64, error) {
	if !c.canCallAuthenticated(ctx) || path == "/" {
		return 0, ErrRequestFailed
	}

	var apiResponse dashboardCountResponse
	if err := c.getJSON(ctx, path, dashboardStatsQueryValues(query), &apiResponse); err != nil {
		return 0, err
	}

	return apiResponse.Data.Count, nil
}

func (c *Client) dashboardAmount(ctx context.Context, path string, query DashboardStatsQuery) (float64, error) {
	if !c.canCallAuthenticated(ctx) || path == "/" {
		return 0, ErrRequestFailed
	}

	var apiResponse dashboardAmountResponse
	if err := c.getJSON(ctx, path, dashboardStatsQueryValues(query), &apiResponse); err != nil {
		return 0, err
	}

	return apiResponse.Data.Amount, nil
}

func dashboardStatsQueryValues(query DashboardStatsQuery) url.Values {
	values := url.Values{}
	values.Set("start_date", query.StartDate.Format("2006-01-02"))
	values.Set("end_date", query.EndDate.Format("2006-01-02"))

	if !query.AllowAllBranches {
		for _, branchID := range query.BranchIDs {
			if branchID > 0 {
				values.Add("branch_ids[]", strconv.FormatUint(branchID, 10))
			}
		}
	}

	return values
}

type adminCustomerListItem struct {
	ID         uint64 `json:"id"`
	BranchID   *int32 `json:"branch_id"`
	BranchName string `json:"branch_name"`
	ZoneName   string `json:"zone_name"`
	Situation  string `json:"situation"`
	PlusCardNo string `json:"plus_card_no"`
	Credit     int64  `json:"credit"`
	Point      int64  `json:"point"`
	Unvan      string `json:"unvan"`
	Ad         string `json:"ad"`
	Soyad      string `json:"soyad"`
	Cep        string `json:"cep"`
	IlKodu     string `json:"il_kodu"`
	IlceKodu   string `json:"ilce_kodu"`
	Type       string `json:"type"`
	Status     int    `json:"status"`
}

type adminCustomerListResponse struct {
	Success    bool                    `json:"success"`
	Message    string                  `json:"message"`
	Items      []adminCustomerListItem `json:"items"`
	Pagination Pagination              `json:"pagination"`
}

func (r adminCustomerListResponse) successful() bool {
	return r.Success
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

	if query.Page > 0 {
		values.Set("page", strconv.Itoa(query.Page))
	}
	if query.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(query.PerPage))
	}

	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	switch sortBy {
	case "credit", "point":
		values.Set("sort_by", "plus_card_balance")
	default:
		if sortBy != "" {
			values.Set("sort_by", sortBy)
		}
	}

	if sortOrder := strings.TrimSpace(query.SortOrder); sortOrder != "" {
		values.Set("sort_order", sortOrder)
	}

	if v := strings.TrimSpace(query.Situation); v != "" {
		values.Set("situation", v)
	}
	if v := strings.TrimSpace(query.BranchName); v != "" {
		values.Set("branch_name", v)
	}
	if v := strings.TrimSpace(query.ZoneName); v != "" {
		values.Set("zone_name", v)
	}
	if v := strings.TrimSpace(query.PlusCardNo); v != "" {
		values.Set("plus_card_no", v)
	}
	if v := strings.TrimSpace(query.Cep); v != "" {
		values.Set("cep", v)
		values.Set("phone", v)
		values.Set("q", v)
	}
	if v := strings.TrimSpace(query.City); v != "" {
		values.Set("city", v)
	}
	if v := strings.TrimSpace(query.Town); v != "" {
		values.Set("town", v)
	}

	return values
}

func mapAdminCustomerListItem(item adminCustomerListItem) CustomerListItem {
	return CustomerListItem{
		ID:         item.ID,
		Situation:  item.Situation,
		BranchName: item.BranchName,
		ZoneName:   item.ZoneName,
		PlusCardNo: item.PlusCardNo,
		Cep:        item.Cep,
		Credit:     item.Credit,
		Point:      item.Point,
		City:       item.IlKodu,
		Town:       item.IlceKodu,
	}
}

func uniquePositiveIDs(ids []uint64) []uint64 {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[uint64]struct{}, len(ids))
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}

func sortCustomerListItems(items []CustomerListItem, sortBy string, sortOrder string) {
	if len(items) < 2 {
		return
	}

	ascending := sortOrder == "asc"

	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		switch sortBy {
		case "credit":
			if left.Credit == right.Credit {
				return left.ID < right.ID
			}
			if ascending {
				return left.Credit < right.Credit
			}
			return left.Credit > right.Credit
		case "point":
			if left.Point == right.Point {
				return left.ID < right.ID
			}
			if ascending {
				return left.Point < right.Point
			}
			return left.Point > right.Point
		default:
			return left.ID < right.ID
		}
	})
}

func paginateCustomerListItems(items []CustomerListItem, page int, perPage int) CustomerListResult {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}

	total := len(items)
	lastPage := (total + perPage - 1) / perPage
	if lastPage <= 0 {
		lastPage = 1
	}

	if page > lastPage {
		page = lastPage
	}

	start := (page - 1) * perPage
	if start > total {
		start = total
	}

	end := start + perPage
	if end > total {
		end = total
	}

	slice := items[start:end]
	var from *int
	var to *int
	if total > 0 && len(slice) > 0 {
		fromValue := start + 1
		toValue := start + len(slice)
		from = &fromValue
		to = &toValue
	}

	return CustomerListResult{
		Items: slice,
		Pagination: Pagination{
			CurrentPage: page,
			LastPage:    lastPage,
			PerPage:     perPage,
			Total:       total,
			From:        from,
			To:          to,
		},
	}
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
	request.Header.Set("Authorization", "Bearer "+c.bearerToken(ctx))

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
