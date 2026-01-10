package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// APIClient gerencia requisições HTTP para a API LEP
type APIClient struct {
	baseURL string
	token   string
	orgID   string
	projID  string
	logger  *Logger
	client  *http.Client
}

// NewAPIClient cria novo cliente de API
func NewAPIClient(baseURL string, logger *Logger) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		logger:  logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetHeaders define headers de autenticação e multi-tenant
func (c *APIClient) SetHeaders(token, orgID, projID string) {
	c.token = token
	c.orgID = orgID
	c.projID = projID
}

// doRequest executa uma requisição HTTP com headers apropriados
func (c *APIClient) doRequest(method, path string, body interface{}) (map[string]interface{}, int, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	var jsonBodyBytes []byte
	if body != nil {
		var err error
		jsonBodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("erro ao serializar body: %w", err)
		}
		// Log do payload se verbose
		if c.logger != nil {
			c.logger.Debug(fmt.Sprintf("[%s] Payload: %s", path, string(jsonBodyBytes)))
		}
		reqBody = bytes.NewBuffer(jsonBodyBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao criar request: %w", err)
	}

	// Headers
	req.Header.Set("Content-Type", "application/json")

	// Auth
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Multi-tenant
	if c.orgID != "" && c.projID != "" {
		req.Header.Set("X-Lpe-Organization-Id", c.orgID)
		req.Header.Set("X-Lpe-Project-Id", c.projID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao executar request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("erro ao ler response: %w", err)
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			// Se não conseguir parsear, armazenar como string raw
			result = map[string]interface{}{"raw": string(respBody)}
		}
		// Log da resposta se status não for 200/201
		if c.logger != nil && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
			c.logger.Debug(fmt.Sprintf("[%s] Response Status: %d, Body: %s", path, resp.StatusCode, string(respBody)))
		}
	}

	return result, resp.StatusCode, nil
}

// CreateOrganization cria uma nova organização
func (c *APIClient) CreateOrganization(name, password string) (orgID, projID, userEmail string, err error) {
	payload := map[string]string{
		"name":     name,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/create-organization", payload)
	if err != nil {
		return "", "", "", err
	}

	if status != 201 && status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return "", "", "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		return "", "", "", fmt.Errorf("status %d", status)
	}

	// Parsear response
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("formato de resposta inválido")
	}

	org, ok := data["organization"].(map[string]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("organização não encontrada na resposta")
	}

	proj, ok := data["project"].(map[string]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("projeto não encontrado na resposta")
	}

	user, ok := data["user"].(map[string]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("usuário não encontrado na resposta")
	}

	orgID = org["id"].(string)
	projID = proj["id"].(string)
	userEmail = user["email"].(string)

	return orgID, projID, userEmail, nil
}

// LoginWithIDs faz login e retorna token, orgID e projID
func (c *APIClient) LoginWithIDs(email, password string) (token, orgID, projID string, err error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/login", payload)
	if err != nil {
		return "", "", "", err
	}

	if status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return "", "", "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		return "", "", "", fmt.Errorf("status %d", status)
	}

	// Tentar parsear resposta com data
	if data, ok := resp["data"].(map[string]interface{}); ok {
		// Extrair token
		if tkn, ok := data["token"].(string); ok {
			token = tkn
		}

		// Extrair orgID
		if org, ok := data["organization"].(map[string]interface{}); ok {
			if id, ok := org["id"].(string); ok {
				orgID = id
			}
		} else if id, ok := data["organization_id"].(string); ok {
			orgID = id
		}

		// Extrair projID
		if proj, ok := data["project"].(map[string]interface{}); ok {
			if id, ok := proj["id"].(string); ok {
				projID = id
			}
		} else if id, ok := data["project_id"].(string); ok {
			projID = id
		}

		if token != "" {
			return token, orgID, projID, nil
		}
	}

	// Fallback: tentar token direto na raiz
	if tkn, ok := resp["token"].(string); ok {
		return tkn, "", "", nil
	}

	return "", "", "", fmt.Errorf("token não encontrado na resposta")
}

// Login faz login de um usuário
func (c *APIClient) Login(email, password string) (token string, err error) {
	token, _, _, err = c.LoginWithIDs(email, password)
	return token, err
}

// CreateMenu cria um novo menu
func (c *APIClient) CreateMenu(menu MenuData) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":   menu.Name,
		"order":  menu.Order,
		"active": true,
		// Optional: "styling"
	}

	resp, status, err := c.doRequest("POST", "/menu", payload)
	if err != nil {
		return uuid.Nil, err
	}

	// Status 409 = Already exists
	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 201 && status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return uuid.Nil, fmt.Errorf("status %d: %s", status, errMsg)
		}
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractID(resp)
}

// CreateCategory cria uma nova categoria
func (c *APIClient) CreateCategory(cat CategoryData, menuID *uuid.UUID) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":   cat.Name,
		"order":  cat.Order,
		"active": true,
		// Optional fields:
		"image_url": cat.Description, // Maps to 'photo' in DB
		"notes":     cat.Description,
	}

	// menu_id is REQUIRED
	if menuID != nil {
		payload["menu_id"] = menuID.String()
	}

	resp, status, err := c.doRequest("POST", "/category", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 201 && status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return uuid.Nil, fmt.Errorf("status %d: %s", status, errMsg)
		}
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractID(resp)
}

// CreateSubcategory cria uma nova subcategoria
func (c *APIClient) CreateSubcategory(sub SubcategoryData, catID uuid.UUID) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":        sub.Name,
		"description": sub.Description,
		"category_id": catID.String(),
		"active":      sub.Active,
		"order":       sub.Order,
	}

	resp, status, err := c.doRequest("POST", "/subcategory", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 201 && status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return uuid.Nil, fmt.Errorf("status %d: %s", status, errMsg)
		}
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractID(resp)
}

// CreateEnvironment cria um novo ambiente
func (c *APIClient) CreateEnvironment(env EnvironmentData) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":     env.Name,
		"capacity": env.Capacity,
		"active":   true,
		// Optional:
		"description": env.Description,
	}

	resp, status, err := c.doRequest("POST", "/environment", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 201 && status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return uuid.Nil, fmt.Errorf("status %d: %s", status, errMsg)
		}
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractID(resp)
}

// CreateTable cria uma nova mesa
func (c *APIClient) CreateTable(tbl TableData, envID *uuid.UUID) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"number":   tbl.Number,
		"capacity": tbl.Capacity,
		"location": tbl.Location,
		"status":   tbl.Status,
	}

	if envID != nil {
		payload["environment_id"] = envID.String()
	}

	resp, status, err := c.doRequest("POST", "/table", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 201 && status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return uuid.Nil, fmt.Errorf("status %d: %s", status, errMsg)
		}
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractID(resp)
}

// CreateProduct cria um novo produto
func (c *APIClient) CreateProduct(prod ProductData, menuID, catID, subCatID *uuid.UUID) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":              prod.Name,              // REQUIRED
		"type":              prod.Type,              // REQUIRED
		"price_normal":      prod.PriceNormal,       // REQUIRED
		"prep_time_minutes": prod.PrepTimeMinutes,   // REQUIRED
		"order":             prod.Order,             // Optional
		"active":            true,
		// Optional fields:
		"description": prod.Description,
	}

	// Adicionar price_promo apenas se for maior que 0
	if prod.PricePromo > 0 {
		payload["price_promo"] = prod.PricePromo
	}

	if menuID != nil {
		payload["menu_id"] = menuID.String()
	}

	if catID != nil {
		payload["category_id"] = catID.String()
	}

	if subCatID != nil {
		payload["subcategory_id"] = subCatID.String()
	}

	// Campos específicos de vinho
	if prod.Type == "vinho" {
		if prod.Vintage != "" {
			payload["vintage"] = prod.Vintage
		}
		if prod.Country != "" {
			payload["country"] = prod.Country
		}
		if prod.Region != "" {
			payload["region"] = prod.Region
		}
		if prod.Winery != "" {
			payload["winery"] = prod.Winery
		}
		if prod.WineType != "" {
			payload["wine_type"] = prod.WineType
		}
		if prod.Volume > 0 {
			payload["volume"] = prod.Volume
		}
		if prod.AlcoholContent > 0 {
			payload["alcohol_content"] = prod.AlcoholContent
		}
		if prod.PriceBottle > 0 {
			payload["price_bottle"] = prod.PriceBottle
		}
		if prod.PriceHalfBottle > 0 {
			payload["price_half_bottle"] = prod.PriceHalfBottle
		}
		if prod.PriceGlass > 0 {
			payload["price_glass"] = prod.PriceGlass
		}
	}

	resp, status, err := c.doRequest("POST", "/product", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 201 && status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return uuid.Nil, fmt.Errorf("status %d: %s", status, errMsg)
		}
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractID(resp)
}

// extractID extrai o ID de uma resposta de criação
func extractID(resp map[string]interface{}) (uuid.UUID, error) {
	// Tentar em data.id
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			return uuid.Parse(id)
		}
	}

	// Tentar em id direto
	if id, ok := resp["id"].(string); ok {
		return uuid.Parse(id)
	}

	return uuid.Nil, fmt.Errorf("ID não encontrado na resposta")
}

// GetMenus lista todos os menus existentes
func (c *APIClient) GetMenus() ([]map[string]interface{}, error) {
	respRaw, status, err := c.doRequest("GET", "/menu", nil)
	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, fmt.Errorf("status %d ao buscar menus", status)
	}

	// A resposta pode ser um array ou estar em um campo data
	if arr, ok := respRaw["data"].([]interface{}); ok {
		result := make([]map[string]interface{}, len(arr))
		for i, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result[i] = m
			}
		}
		return result, nil
	}

	// Se não tiver "data", tentar como array de items direto
	result := make([]map[string]interface{}, 0)
	for k, v := range respRaw {
		if k != "data" && k != "message" && k != "status" {
			// Pode ser um item individual, retornar como array
			if m, ok := v.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return nil, fmt.Errorf("formato inesperado na resposta dos menus")
}

// GetCategories lista todas as categorias existentes
func (c *APIClient) GetCategories() ([]map[string]interface{}, error) {
	respRaw, status, err := c.doRequest("GET", "/category", nil)
	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, fmt.Errorf("status %d ao buscar categorias", status)
	}

	// A resposta pode ser um array ou estar em um campo data
	if arr, ok := respRaw["data"].([]interface{}); ok {
		result := make([]map[string]interface{}, len(arr))
		for i, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result[i] = m
			}
		}
		return result, nil
	}

	// Se não tiver "data", tentar como array de items direto
	result := make([]map[string]interface{}, 0)
	for k, v := range respRaw {
		if k != "data" && k != "message" && k != "status" {
			if m, ok := v.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return nil, fmt.Errorf("formato inesperado na resposta das categorias")
}

// GetSubcategories lista todas as subcategorias existentes
func (c *APIClient) GetSubcategories() ([]map[string]interface{}, error) {
	respRaw, status, err := c.doRequest("GET", "/subcategory", nil)
	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, fmt.Errorf("status %d ao buscar subcategorias", status)
	}

	// A resposta pode ser um array ou estar em um campo data
	if arr, ok := respRaw["data"].([]interface{}); ok {
		result := make([]map[string]interface{}, len(arr))
		for i, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result[i] = m
			}
		}
		return result, nil
	}

	// Se não tiver "data", tentar como array de items direto
	result := make([]map[string]interface{}, 0)
	for k, v := range respRaw {
		if k != "data" && k != "message" && k != "status" {
			if m, ok := v.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return nil, fmt.Errorf("formato inesperado na resposta das subcategorias")
}

// GetEnvironments lista todos os ambientes existentes
func (c *APIClient) GetEnvironments() ([]map[string]interface{}, error) {
	respRaw, status, err := c.doRequest("GET", "/environment", nil)
	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, fmt.Errorf("status %d ao buscar ambientes", status)
	}

	// A resposta pode ser um array ou estar em um campo data
	if arr, ok := respRaw["data"].([]interface{}); ok {
		result := make([]map[string]interface{}, len(arr))
		for i, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result[i] = m
			}
		}
		return result, nil
	}

	// Se não tiver "data", tentar como array de items direto
	result := make([]map[string]interface{}, 0)
	for k, v := range respRaw {
		if k != "data" && k != "message" && k != "status" {
			if m, ok := v.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return nil, fmt.Errorf("formato inesperado na resposta dos ambientes")
}
