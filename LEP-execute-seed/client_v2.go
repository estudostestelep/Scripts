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

// APIClientV2 é um cliente HTTP otimizado para a API LEP
type APIClientV2 struct {
	baseURL string
	token   string
	orgID   string
	projID  string
	logger  *Logger
	client  *http.Client
	config  *Config
}

// NewAPIClientV2 cria novo cliente de API
func NewAPIClientV2(baseURL string, logger *Logger, config *Config) *APIClientV2 {
	return &APIClientV2{
		baseURL: baseURL,
		logger:  logger,
		config:  config,
		client: &http.Client{
			Timeout: time.Duration(config.Server.Timeout) * time.Second,
		},
	}
}

// SetHeaders define headers de autenticação e multi-tenant
func (c *APIClientV2) SetHeaders(token, orgID, projID string) {
	c.token = token
	c.orgID = orgID
	c.projID = projID
}

// doRequest executa requisição com tratamento de erro
func (c *APIClientV2) doRequest(method, path string, body interface{}) (map[string]interface{}, int, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	var jsonBodyBytes []byte

	if body != nil {
		var err error
		jsonBodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("erro ao serializar body: %w", err)
		}

		// Log do payload se enabled
		if c.config.Logging.ShowPayloads {
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
			result = map[string]interface{}{"raw": string(respBody)}
		}

		// Log de erro se status não for sucesso
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			c.logger.Debug(fmt.Sprintf("[%s] Status: %d, Body: %s", path, resp.StatusCode, string(respBody)))
		}
	}

	return result, resp.StatusCode, nil
}

// CreateOrganization cria organização ou faz login se existir
func (c *APIClientV2) CreateOrganization(name, email, password string) (orgID, projID string, err error) {
	// 1. Tentar criar organização
	payload := map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/create-organization", payload)
	if err != nil {
		return "", "", err
	}

	// Se status 409, organização já existe - fazer login
	if status == 409 {
		c.logger.Info("Organização já existe, fazendo login...")
		return c.LoginAndGetIDs(email, password)
	}

	if status != 200 && status != 201 {
		if errMsg, ok := resp["message"].(string); ok {
			return "", "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		return "", "", fmt.Errorf("status %d", status)
	}

	// Extrair IDs da resposta
	return c.extractOrgAndProjID(resp, "")
}

// LoginAndGetIDs faz login e extrai IDs
func (c *APIClientV2) LoginAndGetIDs(email, password string) (orgID, projID string, err error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/login", payload)
	if err != nil {
		return "", "", err
	}

	if status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return "", "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		return "", "", fmt.Errorf("status %d", status)
	}

	// extractOrgAndProjID já extrai o token também
	return c.extractOrgAndProjID(resp, "")
}

// LoginAndGetIDsForOrg faz login e busca IDs de uma organização específica
func (c *APIClientV2) LoginAndGetIDsForOrg(email, password, orgName string) (orgID, projID string, err error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/login", payload)
	if err != nil {
		return "", "", err
	}

	if status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return "", "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		return "", "", fmt.Errorf("status %d", status)
	}

	// extractOrgAndProjID com nome da organização para buscar o projeto correto
	return c.extractOrgAndProjID(resp, orgName)
}

// extractOrgAndProjID extrai IDs da resposta do login
// Estrutura esperada:
// {
//   "user": {...},
//   "token": "jwt",
//   "organizations": [...],
//   "projects": [
//     {
//       "id": "project-uuid",
//       "project_id": "project-uuid",
//       "organization_id": "org-uuid",
//       "organization_name": "LEP Fattoria",
//       ...
//     }
//   ]
// }
// Se orgName for fornecido, procura pelo projeto da organização com esse nome
func (c *APIClientV2) extractOrgAndProjID(resp map[string]interface{}, orgName string) (orgID, projID string, err error) {
	// Extrair token se estiver na resposta
	if tkn, ok := resp["token"].(string); ok && tkn != "" {
		c.token = tkn
	}

	// Extrair do array "projects" (estrutura real do backend)
	if projects, ok := resp["projects"].([]interface{}); ok && len(projects) > 0 {
		// Se orgName foi fornecido, procurar pelo projeto da organização correta
		if orgName != "" {
			// Log debug: listar todas as organizações disponíveis
			c.logger.Debug(fmt.Sprintf("Organizações disponíveis para este usuário (total: %d):", len(projects)))
			for i, p := range projects {
				if project, ok := p.(map[string]interface{}); ok {
					orgNameInProject := project["organization_name"]
					orgID := project["organization_id"]
					projectID := project["project_id"]
					c.logger.Debug(fmt.Sprintf("  [%d] org_name=%v, org_id=%v, proj_id=%v", i, orgNameInProject, orgID, projectID))
				}
			}

			// ESTRATÉGIA 1: Tentar buscar pelo organization_name (se existir)
			for _, p := range projects {
				if project, ok := p.(map[string]interface{}); ok {
					// Verificar se o nome da organização corresponde
					if orgNameInProject, ok := project["organization_name"].(string); ok && orgNameInProject == orgName {
						// Extrair project_id
						if id, ok := project["project_id"].(string); ok {
							projID = id
						}
						// Extrair organization_id
						if id, ok := project["organization_id"].(string); ok {
							orgID = id
						}

						// Se tiver sucesso, retornar
						if orgID != "" && projID != "" {
							c.orgID = orgID
							c.projID = projID
							c.logger.Info(fmt.Sprintf("Usando organização: %s (ID: %s)", orgName, orgID))
							return orgID, projID, nil
						}
					}
				}
			}

			// ESTRATÉGIA 2: Backend não retorna organization_name, usar o SEGUNDO projeto
			// (pois o primeiro é geralmente DEFAULT e o segundo é LEP Fattoria)
			c.logger.Info("Backend não retorna 'organization_name', usando segundo projeto como 'LEP Fattoria'")
			if len(projects) >= 2 {
				if secondProject, ok := projects[1].(map[string]interface{}); ok {
					// Extrair project_id
					if id, ok := secondProject["project_id"].(string); ok {
						projID = id
					}
					// Extrair organization_id
					if id, ok := secondProject["organization_id"].(string); ok {
						orgID = id
					}

					// Se tiver sucesso, retornar
					if orgID != "" && projID != "" {
						c.orgID = orgID
						c.projID = projID
						c.logger.Info(fmt.Sprintf("Usando segunda organização: %s (ID: %s)", orgName, orgID))
						return orgID, projID, nil
					}
				}
			}

			// Se não encontrou a organização específica, retornar erro
			return "", "", fmt.Errorf("organização '%s' não encontrada nos projetos do usuário", orgName)
		}

		// Se orgName não foi fornecido, usar o primeiro projeto (comportamento original)
		if firstProject, ok := projects[0].(map[string]interface{}); ok {
			// Extrair project_id
			if id, ok := firstProject["project_id"].(string); ok {
				projID = id
			}
			// Extrair organization_id
			if id, ok := firstProject["organization_id"].(string); ok {
				orgID = id
			}

			// Se tiver sucesso, retornar
			if orgID != "" && projID != "" {
				c.orgID = orgID
				c.projID = projID
				return orgID, projID, nil
			}
		}
	}

	// Fallback: tentar estrutura anteprovada com "data"
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if org, ok := data["organization"].(map[string]interface{}); ok {
			if id, ok := org["id"].(string); ok {
				orgID = id
			}
		}
		if orgID == "" {
			if id, ok := data["organization_id"].(string); ok {
				orgID = id
			}
		}

		if proj, ok := data["project"].(map[string]interface{}); ok {
			if id, ok := proj["id"].(string); ok {
				projID = id
			}
		}
		if projID == "" {
			if id, ok := data["project_id"].(string); ok {
				projID = id
			}
		}

		if orgID != "" && projID != "" {
			c.orgID = orgID
			c.projID = projID
			return orgID, projID, nil
		}
	}

	return "", "", fmt.Errorf("IDs não encontrados na resposta (esperado: projects[0].project_id e projects[0].organization_id)")
}

// CreateMenu cria menu
func (c *APIClientV2) CreateMenu(name string, order int) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":   name,
		"order":  order,
		"active": true,
	}

	resp, status, err := c.doRequest("POST", "/menu", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// CreateCategory cria categoria
func (c *APIClientV2) CreateCategory(menuID string, name string, order int) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"menu_id": menuID,
		"name":    name,
		"order":   order,
		"active":  true,
	}

	resp, status, err := c.doRequest("POST", "/category", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// CreateSubcategory cria subcategoria
func (c *APIClientV2) CreateSubcategory(catID string, name string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"category_id": catID,
		"name":        name,
		"active":      true,
	}

	resp, status, err := c.doRequest("POST", "/subcategory", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// CreateEnvironment cria ambiente
func (c *APIClientV2) CreateEnvironment(name string, capacity int) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":     name,
		"capacity": capacity,
		"active":   true,
	}

	resp, status, err := c.doRequest("POST", "/environment", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// CreateTable cria mesa
func (c *APIClientV2) CreateTable(number int, capacity int, envID *string, status string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"number":   number,
		"capacity": capacity,
		"status":   status,
		"active":   true,
	}

	if envID != nil && *envID != "" {
		payload["environment_id"] = *envID
	}

	resp, respStatus, err := c.doRequest("POST", "/table", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if respStatus == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if respStatus != 200 && respStatus != 201 {
		return uuid.Nil, fmt.Errorf("status %d", respStatus)
	}

	return extractIDFromResponse(resp)
}

// CreateProduct cria produto
func (c *APIClientV2) CreateProduct(name string, productType string, priceNormal float64, prepTime int, menuID, categoryID, subcategoryID *string, wineData *WineData) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":              name,
		"type":              productType,
		"price_normal":      priceNormal,
		"prep_time_minutes": prepTime,
		"active":            true,
		"order":             0,
	}

	if menuID != nil && *menuID != "" {
		payload["menu_id"] = *menuID
	}

	if categoryID != nil && *categoryID != "" {
		payload["category_id"] = *categoryID
	}

	if subcategoryID != nil && *subcategoryID != "" {
		payload["subcategory_id"] = *subcategoryID
	}

	// Adicionar campos específicos de vinho se fornecidos
	if wineData != nil {
		if wineData.Vintage != "" {
			payload["vintage"] = wineData.Vintage
		}
		if wineData.Country != "" {
			payload["country"] = wineData.Country
		}
		if wineData.Region != "" {
			payload["region"] = wineData.Region
		}
		if wineData.Winery != "" {
			payload["winery"] = wineData.Winery
		}
		if wineData.WineType != "" {
			payload["wine_type"] = wineData.WineType
		}
		if wineData.Volume > 0 {
			payload["volume"] = wineData.Volume
		}
		if wineData.AlcoholContent > 0 {
			payload["alcohol_content"] = wineData.AlcoholContent
		}
		if wineData.PriceBottle > 0 {
			payload["price_bottle"] = wineData.PriceBottle
		}
		if wineData.PriceGlass > 0 {
			payload["price_glass"] = wineData.PriceGlass
		}
	}

	resp, status, err := c.doRequest("POST", "/product", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// WineData contém dados específicos de vinhos
type WineData struct {
	Vintage        string
	Country        string
	Region         string
	Winery         string
	WineType       string
	Volume         int
	AlcoholContent float64
	PriceBottle    float64
	PriceGlass     float64
}

// GetMenuByName busca um menu pelo nome (para evitar duplicatas)
func (c *APIClientV2) GetMenuByName(name string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/menu", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	// Procurar menu com esse nome
	if menus, ok := resp["data"].([]interface{}); ok {
		for _, m := range menus {
			if menu, ok := m.(map[string]interface{}); ok {
				if menuName, ok := menu["name"].(string); ok && menuName == name {
					if id, ok := menu["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetCategoryByName busca uma categoria pelo nome
func (c *APIClientV2) GetCategoryByName(name string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/category", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if categories, ok := resp["data"].([]interface{}); ok {
		for _, cat := range categories {
			if category, ok := cat.(map[string]interface{}); ok {
				if catName, ok := category["name"].(string); ok && catName == name {
					if id, ok := category["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetSubcategoryByName busca uma subcategoria pelo nome
func (c *APIClientV2) GetSubcategoryByName(name string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/subcategory", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if subcategories, ok := resp["data"].([]interface{}); ok {
		for _, subcat := range subcategories {
			if subcategory, ok := subcat.(map[string]interface{}); ok {
				if subcatName, ok := subcategory["name"].(string); ok && subcatName == name {
					if id, ok := subcategory["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetProductByName busca um produto pelo nome
func (c *APIClientV2) GetProductByName(name string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/product", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if products, ok := resp["data"].([]interface{}); ok {
		for _, prod := range products {
			if product, ok := prod.(map[string]interface{}); ok {
				if prodName, ok := product["name"].(string); ok && prodName == name {
					if id, ok := product["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetEnvironmentByName busca um ambiente pelo nome
func (c *APIClientV2) GetEnvironmentByName(name string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/environment", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if environments, ok := resp["data"].([]interface{}); ok {
		for _, env := range environments {
			if environment, ok := env.(map[string]interface{}); ok {
				if envName, ok := environment["name"].(string); ok && envName == name {
					if id, ok := environment["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetTableByNumber busca uma mesa pelo número
func (c *APIClientV2) GetTableByNumber(number int) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/table", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if tables, ok := resp["data"].([]interface{}); ok {
		for _, tbl := range tables {
			if table, ok := tbl.(map[string]interface{}); ok {
				if tblNum, ok := table["number"].(float64); ok && int(tblNum) == number {
					if id, ok := table["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// CreateUser cria um novo usuário
func (c *APIClientV2) CreateUser(name, email, password, role string, permissions []string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":     name,
		"email":    email,
		"password": password,
		"role":     role,
		"active":   true,
	}

	if len(permissions) > 0 {
		payload["permissions"] = permissions
	}

	resp, status, err := c.doRequest("POST", "/user", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// GetUserByEmail busca um usuário pelo email
func (c *APIClientV2) GetUserByEmail(email string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/user", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if users, ok := resp["data"].([]interface{}); ok {
		for _, u := range users {
			if user, ok := u.(map[string]interface{}); ok {
				if userEmail, ok := user["email"].(string); ok && userEmail == email {
					if id, ok := user["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// CreateCustomer cria um novo cliente
func (c *APIClientV2) CreateCustomer(name, email, phone, birthDate, notes string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":   name,
		"email":  email,
		"phone":  phone,
		"active": true,
	}

	if birthDate != "" {
		payload["birth_date"] = birthDate
	}

	if notes != "" {
		payload["notes"] = notes
	}

	resp, status, err := c.doRequest("POST", "/customer", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// GetCustomerByEmail busca um cliente pelo email
func (c *APIClientV2) GetCustomerByEmail(email string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/customer", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if customers, ok := resp["data"].([]interface{}); ok {
		for _, cust := range customers {
			if customer, ok := cust.(map[string]interface{}); ok {
				if custEmail, ok := customer["email"].(string); ok && custEmail == email {
					if id, ok := customer["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// CreateReservation cria uma nova reserva
func (c *APIClientV2) CreateReservation(customerID, tableID string, dateTime string, partySize int, notes, status, confirmationKey string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"customer_id": customerID,
		"table_id":    tableID,
		"datetime":    dateTime,
		"party_size":  partySize,
		"status":      status,
	}

	if notes != "" {
		payload["notes"] = notes
	}

	if confirmationKey != "" {
		payload["confirmation_key"] = confirmationKey
	}

	resp, respStatus, err := c.doRequest("POST", "/reservation", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if respStatus == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if respStatus != 200 && respStatus != 201 {
		return uuid.Nil, fmt.Errorf("status %d", respStatus)
	}

	return extractIDFromResponse(resp)
}

// GetReservationByConfirmationKey busca uma reserva pela chave de confirmação
func (c *APIClientV2) GetReservationByConfirmationKey(confirmationKey string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/reservation", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if reservations, ok := resp["data"].([]interface{}); ok {
		for _, res := range reservations {
			if reservation, ok := res.(map[string]interface{}); ok {
				if key, ok := reservation["confirmation_key"].(string); ok && key == confirmationKey {
					if id, ok := reservation["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// CreateTag cria uma nova tag
func (c *APIClientV2) CreateTag(name, color, description, entityType string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":   name,
		"active": true,
	}

	if color != "" {
		payload["color"] = color
	}

	if description != "" {
		payload["description"] = description
	}

	if entityType != "" {
		payload["entity_type"] = entityType
	}

	resp, status, err := c.doRequest("POST", "/tag", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// GetTagByName busca uma tag pelo nome
func (c *APIClientV2) GetTagByName(name string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/tag", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if tags, ok := resp["data"].([]interface{}); ok {
		for _, t := range tags {
			if tag, ok := t.(map[string]interface{}); ok {
				if tagName, ok := tag["name"].(string); ok && tagName == name {
					if id, ok := tag["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// AddCategoryToSubcategory vincula subcategoria a uma categoria (relacionamento N:M)
func (c *APIClientV2) AddCategoryToSubcategory(subcatID, catID string) error {
	path := fmt.Sprintf("/subcategory/%s/category/%s", subcatID, catID)

	// Backend espera JSON body com category_id (mesmo com path param)
	payload := map[string]interface{}{
		"category_id": catID,
	}

	_, status, err := c.doRequest("POST", path, payload)
	if err != nil {
		return err
	}

	if status == 409 {
		return nil // Relacionamento já existe, ignorar
	}

	if status != 200 && status != 201 {
		return fmt.Errorf("status %d", status)
	}

	return nil
}

// AddTagToProduct vincula tag a um produto (relacionamento N:M)
func (c *APIClientV2) AddTagToProduct(productID, tagID string) error {
	path := fmt.Sprintf("/product/%s/tag/%s", productID, tagID)

	// Backend espera JSON body com tag_id (mesmo com path param)
	payload := map[string]interface{}{
		"tag_id": tagID,
	}

	_, status, err := c.doRequest("POST", path, payload)
	if err != nil {
		return err
	}

	if status == 409 {
		return nil // Relacionamento já existe, ignorar
	}

	if status != 200 && status != 201 {
		return fmt.Errorf("status %d", status)
	}

	return nil
}

// CreateSettings cria configurações do projeto
func (c *APIClientV2) CreateSettings(settings *SettingsData) error {
	payload := map[string]interface{}{}

	if settings.ReservationMinAdvanceHours > 0 {
		payload["min_advance_hours"] = settings.ReservationMinAdvanceHours
	}
	if settings.ReservationMaxAdvanceDays > 0 {
		payload["max_advance_days"] = settings.ReservationMaxAdvanceDays
	}

	payload["notify_reservation_create"] = settings.NotifyReservationCreate
	payload["notify_reservation_update"] = settings.NotifyReservationUpdate
	payload["notify_reservation_cancel"] = settings.NotifyReservationCancel
	payload["notify_table_available"] = settings.NotifyTableAvailable
	payload["notify_confirmation_24h"] = settings.NotifyConfirmation24h

	if settings.DefaultNotificationChannel != "" {
		payload["default_notification_channel"] = settings.DefaultNotificationChannel
	}

	payload["enable_sms"] = settings.EnableSMS
	payload["enable_email"] = settings.EnableEmail
	payload["enable_whatsapp"] = settings.EnableWhatsApp

	if settings.Timezone != "" {
		payload["timezone"] = settings.Timezone
	}

	_, status, err := c.doRequest("POST", "/settings", payload)
	if err != nil {
		return err
	}

	if status == 409 {
		// Settings já existe, tentar atualizar
		_, status, err = c.doRequest("PUT", "/settings", payload)
		if err != nil {
			return err
		}
	}

	if status != 200 && status != 201 {
		return fmt.Errorf("status %d", status)
	}

	return nil
}

// CreateNotificationTemplate cria template de notificação
func (c *APIClientV2) CreateNotificationTemplate(template *NotificationTemplateData) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":    template.Name,
		"channel": template.Channel,
		"body":    template.Body,
		"active":  template.Active,
	}

	if template.Subject != "" {
		payload["subject"] = template.Subject
	}

	resp, status, err := c.doRequest("POST", "/notification-template", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if status == 409 {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// CreateThemeCustomization cria customização de tema
func (c *APIClientV2) CreateThemeCustomization(theme *ThemeCustomizationData) error {
	// Converter cores single para Light + Dark
	payload := map[string]interface{}{
		// Light mode
		"primary_color_light":          theme.PrimaryColor,
		"secondary_color_light":        theme.SecondaryColor,
		"background_color_light":       theme.BackgroundColor,
		"card_background_color_light":  theme.CardBackgroundColor,
		"text_color_light":             theme.TextColor,
		"text_secondary_color_light":   theme.TextSecondaryColor,
		"accent_color_light":           theme.AccentColor,
		"success_color_light":          theme.SuccessColor,
		"error_color_light":            theme.ErrorColor,
		"warning_color_light":          theme.WarningColor,
		"info_color_light":             theme.InfoColor,

		// Dark mode (usando mesmas cores como fallback)
		"primary_color_dark":           theme.PrimaryColor,
		"secondary_color_dark":         theme.SecondaryColor,
		"background_color_dark":        "#1a1a1a",  // Dark background
		"card_background_color_dark":   "#2d2d2d",  // Dark card
		"text_color_dark":              "#f0f0f0",  // Light text
		"text_secondary_color_dark":    "#b0b0b0",  // Gray text
		"accent_color_dark":            theme.AccentColor,
		"success_color_dark":           theme.SuccessColor,
		"error_color_dark":             theme.ErrorColor,
		"warning_color_dark":           theme.WarningColor,
		"info_color_dark":              theme.InfoColor,

		"disabled_opacity":             theme.DisabledOpacity,
		"shadow_intensity":             theme.ShadowIntensity,
		"is_active":                    theme.IsActive,
	}

	_, status, err := c.doRequest("POST", "/theme-customization", payload)
	if err != nil {
		return err
	}

	if status == 409 {
		// Theme já existe, tentar atualizar
		_, status, err = c.doRequest("PUT", "/theme-customization", payload)
		if err != nil {
			return err
		}
	}

	if status != 200 && status != 201 {
		return fmt.Errorf("status %d", status)
	}

	return nil
}

// GetNotificationTemplateByName busca template por nome
func (c *APIClientV2) GetNotificationTemplateByName(name string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/notification-template", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	if templates, ok := resp["data"].([]interface{}); ok {
		for _, t := range templates {
			if template, ok := t.(map[string]interface{}); ok {
				if tName, ok := template["name"].(string); ok && tName == name {
					if id, ok := template["id"].(string); ok {
						return uuid.Parse(id)
					}
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// extractIDFromResponse extrai ID da resposta JSON
func extractIDFromResponse(resp map[string]interface{}) (uuid.UUID, error) {
	// Tentar extrair do campo "data"
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			return uuid.Parse(id)
		}
	}

	// Fallback: tentar direto na raiz
	if id, ok := resp["id"].(string); ok {
		return uuid.Parse(id)
	}

	return uuid.Nil, fmt.Errorf("ID não encontrado na resposta")
}
