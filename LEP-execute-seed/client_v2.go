package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
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
	roleMap map[string]string
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
		// Tentar primeiro como objeto
		if err := json.Unmarshal(respBody, &result); err != nil {
			// Se falhar, tentar como array
			var arrResult []interface{}
			if arrErr := json.Unmarshal(respBody, &arrResult); arrErr == nil {
				// Sucesso: colocar array em um campo "items"
				result = map[string]interface{}{"items": arrResult}
			} else {
				// Falhou em ambos: guardar como raw
				result = map[string]interface{}{"raw": string(respBody)}
			}
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

	// Se organização já existe (409, 400 com "já existe", ou 500 com duplicate key)
	if status == 409 || isOrgAlreadyExistsError(resp, status) {
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

// LoginAndGetIDs faz login e extrai IDs (LEGACY - usa /login)
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

// AdminLogin faz login como admin usando /admin/login
// Retorna token e permissões do admin
func (c *APIClientV2) AdminLogin(email, password string) (adminID string, err error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/admin/login", payload)
	if err != nil {
		return "", err
	}

	if status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		if errMsg, ok := resp["error"].(string); ok {
			return "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		return "", fmt.Errorf("status %d", status)
	}

	// Extrair token
	if tkn, ok := resp["token"].(string); ok && tkn != "" {
		c.token = tkn
	}

	// Extrair admin ID
	if admin, ok := resp["admin"].(map[string]interface{}); ok {
		if id, ok := admin["id"].(string); ok {
			adminID = id
		}
	}

	c.logger.Info(fmt.Sprintf("Admin login bem-sucedido (ID: %s)", adminID))
	return adminID, nil
}

// ClientLogin faz login como client usando /client/login
// Requer org_slug para identificar a organização
func (c *APIClientV2) ClientLogin(email, password, orgSlug string) (clientID, orgID, projID string, err error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
		"org_slug": orgSlug,
	}

	resp, status, err := c.doRequest("POST", "/client/login", payload)
	if err != nil {
		return "", "", "", err
	}

	if status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return "", "", "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		if errMsg, ok := resp["error"].(string); ok {
			return "", "", "", fmt.Errorf("status %d: %s", status, errMsg)
		}
		return "", "", "", fmt.Errorf("status %d", status)
	}

	// Extrair token
	if tkn, ok := resp["token"].(string); ok && tkn != "" {
		c.token = tkn
	}

	// Extrair client ID
	if client, ok := resp["client"].(map[string]interface{}); ok {
		if id, ok := client["id"].(string); ok {
			clientID = id
		}
	}

	// Extrair organization ID
	if org, ok := resp["organization"].(map[string]interface{}); ok {
		if id, ok := org["id"].(string); ok {
			orgID = id
			c.orgID = orgID
		}
	}

	// Extrair primeiro projeto
	if projects, ok := resp["projects"].([]interface{}); ok && len(projects) > 0 {
		if proj, ok := projects[0].(map[string]interface{}); ok {
			if id, ok := proj["project_id"].(string); ok {
				projID = id
				c.projID = projID
			}
		}
	}

	c.logger.Info(fmt.Sprintf("Client login bem-sucedido (ID: %s, Org: %s)", clientID, orgID))
	return clientID, orgID, projID, nil
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

	// Fallback: se projects/organizations estão vazios (novo sistema user_roles),
	// buscar organizações via API e encontrar a organização pelo nome
	if orgName != "" && c.token != "" {
		c.logger.Info("Projects vazio no login, buscando organizações via API...")
		foundOrgID, foundProjID, apiErr := c.findOrgAndProjectByName(orgName)
		if apiErr == nil && foundOrgID != "" && foundProjID != "" {
			c.orgID = foundOrgID
			c.projID = foundProjID
			return foundOrgID, foundProjID, nil
		}
		if apiErr != nil {
			c.logger.Debug(fmt.Sprintf("Fallback API falhou: %v", apiErr))
		}
	}

	return "", "", fmt.Errorf("IDs não encontrados na resposta (esperado: projects[0].project_id e projects[0].organization_id)")
}

// findOrgAndProjectByName busca organização pelo nome via API e retorna orgID e projID
func (c *APIClientV2) findOrgAndProjectByName(orgName string) (orgID, projID string, err error) {
	// Buscar todas as organizações
	resp, status, err := c.doRequest("GET", "/organization", nil)
	if err != nil {
		return "", "", err
	}
	if status != 200 {
		return "", "", fmt.Errorf("status %d ao buscar organizações", status)
	}

	orgs := extractArrayFromResponse(resp)
	for _, o := range orgs {
		if org, ok := o.(map[string]interface{}); ok {
			name, _ := org["name"].(string)
			if name == orgName {
				if id, ok := org["id"].(string); ok {
					orgID = id
				}
				break
			}
		}
	}

	if orgID == "" {
		return "", "", fmt.Errorf("organização '%s' não encontrada via API", orgName)
	}

	// Buscar projetos dessa organização
	projResp, projStatus, projErr := c.doRequest("GET", fmt.Sprintf("/project/organization/%s", orgID), nil)
	if projErr != nil || projStatus != 200 {
		// Fallback: tentar /project com header da organização
		c.orgID = orgID
		projResp, projStatus, projErr = c.doRequest("GET", "/project", nil)
		if projErr != nil || projStatus != 200 {
			return orgID, "", fmt.Errorf("não foi possível buscar projetos da organização %s", orgID)
		}
	}

	projects := extractArrayFromResponse(projResp)
	if len(projects) > 0 {
		if proj, ok := projects[0].(map[string]interface{}); ok {
			if id, ok := proj["id"].(string); ok {
				projID = id
			}
		}
	}

	if projID == "" {
		return orgID, "", fmt.Errorf("nenhum projeto encontrado para organização %s", orgID)
	}

	c.logger.Info(fmt.Sprintf("Encontrado via API: org=%s, proj=%s", orgID, projID))
	return orgID, projID, nil
}

// CreateMenu cria menu
func (c *APIClientV2) CreateMenu(name string, order int) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":   name,
		"order":  order,
		"active": true,
	}

	resp, status, err := c.doRequest("POST", "/admin/menu", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if isAlreadyExistsError(resp, status) {
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

	resp, status, err := c.doRequest("POST", "/admin/category", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if isAlreadyExistsError(resp, status) {
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

	resp, status, err := c.doRequest("POST", "/admin/subcategory", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if isAlreadyExistsError(resp, status) {
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

	if isAlreadyExistsError(resp, status) {
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

	if isAlreadyExistsError(resp, respStatus) {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if respStatus != 200 && respStatus != 201 {
		return uuid.Nil, fmt.Errorf("status %d", respStatus)
	}

	return extractIDFromResponse(resp)
}

// CreateProduct cria produto
func (c *APIClientV2) CreateProduct(name string, productType string, priceNormal float64, prepTime int, menuID, categoryID, subcategoryID *string, wineData *WineData, imageURL string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":              name,
		"type":              productType,
		"price_normal":      priceNormal,
		"prep_time_minutes": prepTime,
		"active":            true,
		"order":             0,
	}

	// Adicionar URL da imagem se fornecida
	if imageURL != "" {
		payload["image_url"] = imageURL
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

	if isAlreadyExistsError(resp, status) {
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
	// Backend usa /menu para GET (client routes), não /admin/menu
	resp, status, err := c.doRequest("GET", "/menu", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	// Resposta pode vir como array direto ou dentro de "data"
	var menus []interface{}
	if data, ok := resp["data"].([]interface{}); ok {
		menus = data
	} else {
		// Tentar como array na raiz (resposta direta do backend)
		for _, v := range resp {
			if arr, ok := v.([]interface{}); ok {
				menus = arr
				break
			}
		}
	}

	// Procurar menu com esse nome
	for _, m := range menus {
		if menu, ok := m.(map[string]interface{}); ok {
			if menuName, ok := menu["name"].(string); ok && menuName == name {
				if id, ok := menu["id"].(string); ok {
					return uuid.Parse(id)
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetCategoryByName busca uma categoria pelo nome
func (c *APIClientV2) GetCategoryByName(name string) (uuid.UUID, error) {
	// Backend usa /category para GET (client routes), não /admin/category
	resp, status, err := c.doRequest("GET", "/category", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	categories := extractArrayFromResponse(resp)
	for _, cat := range categories {
		if category, ok := cat.(map[string]interface{}); ok {
			if catName, ok := category["name"].(string); ok && catName == name {
				if id, ok := category["id"].(string); ok {
					return uuid.Parse(id)
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetSubcategoryByName busca uma subcategoria pelo nome
func (c *APIClientV2) GetSubcategoryByName(name string) (uuid.UUID, error) {
	// Backend usa /subcategory para GET (client routes), não /admin/subcategory
	resp, status, err := c.doRequest("GET", "/subcategory", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	subcategories := extractArrayFromResponse(resp)
	for _, subcat := range subcategories {
		if subcategory, ok := subcat.(map[string]interface{}); ok {
			if subcatName, ok := subcategory["name"].(string); ok && subcatName == name {
				if id, ok := subcategory["id"].(string); ok {
					return uuid.Parse(id)
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

	products := extractArrayFromResponse(resp)
	for _, prod := range products {
		if product, ok := prod.(map[string]interface{}); ok {
			if prodName, ok := product["name"].(string); ok && prodName == name {
				if id, ok := product["id"].(string); ok {
					return uuid.Parse(id)
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

	environments := extractArrayFromResponse(resp)
	for _, env := range environments {
		if environment, ok := env.(map[string]interface{}); ok {
			if envName, ok := environment["name"].(string); ok && envName == name {
				if id, ok := environment["id"].(string); ok {
					return uuid.Parse(id)
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

	tables := extractArrayFromResponse(resp)
	for _, tbl := range tables {
		if table, ok := tbl.(map[string]interface{}); ok {
			if tblNum, ok := table["number"].(float64); ok && int(tblNum) == number {
				if id, ok := table["id"].(string); ok {
					return uuid.Parse(id)
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetSystemRoles busca os cargos do sistema e retorna mapa nome->id
func (c *APIClientV2) GetSystemRoles() (map[string]string, error) {
	resp, status, err := c.doRequest("GET", "/role/system", nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("status %d ao buscar roles", status)
	}
	roles := extractArrayFromResponse(resp)
	roleMap := make(map[string]string)
	for _, r := range roles {
		if role, ok := r.(map[string]interface{}); ok {
			name, _ := role["name"].(string)
			id, _ := role["id"].(string)
			if name != "" && id != "" {
				roleMap[name] = id
			}
		}
	}
	return roleMap, nil
}

// CreateUser cria um novo usuário (client-user) na organização
func (c *APIClientV2) CreateUser(name, email, password, role string, permissions []string) (uuid.UUID, error) {
	active := true
	payload := map[string]interface{}{
		"name":     name,
		"email":    email,
		"password": password,
		"org_id":   c.orgID,
		"active":   active,
	}

	if c.projID != "" {
		payload["proj_ids"] = []string{c.projID}
	}

	resp, status, err := c.doRequest("POST", "/client-user", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if isAlreadyExistsError(resp, status) {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		details := ""
		for _, key := range []string{"details", "error", "message"} {
			if v, ok := resp[key].(string); ok && v != "" {
				details += " " + v
			}
		}
		return uuid.Nil, fmt.Errorf("status %d:%s", status, details)
	}

	return extractIDFromResponse(resp)
}

// CreateAdmin cria um novo admin (tabela admins)
func (c *APIClientV2) CreateAdmin(name, email, password string, permissions []string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":     name,
		"email":    email,
		"password": password,
		"active":   true,
	}

	if len(permissions) > 0 {
		payload["permissions"] = permissions
	}

	resp, status, err := c.doRequest("POST", "/admin/admin-user", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if isAlreadyExistsError(resp, status) {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// GetAdminByEmail busca um admin pelo email
func (c *APIClientV2) GetAdminByEmail(email string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/admin/admin-user", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	admins := extractArrayFromResponse(resp)
	for _, a := range admins {
		if admin, ok := a.(map[string]interface{}); ok {
			if adminEmail, ok := admin["email"].(string); ok && adminEmail == email {
				if id, ok := admin["id"].(string); ok {
					return uuid.Parse(id)
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// CreateClientUser cria um novo client (tabela clients)
// orgID: organização a qual o client pertence
// projIDs: lista de projetos que o client pode acessar
func (c *APIClientV2) CreateClientUser(name, email, password, orgID string, projIDs, permissions []string) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"name":     name,
		"email":    email,
		"password": password,
		"org_id":   orgID,
		"active":   true,
	}

	if len(projIDs) > 0 {
		payload["proj_ids"] = projIDs
	}

	if len(permissions) > 0 {
		payload["permissions"] = permissions
	}

	resp, status, err := c.doRequest("POST", "/admin/client-user", payload)
	if err != nil {
		return uuid.Nil, err
	}

	if isAlreadyExistsError(resp, status) {
		return uuid.Nil, fmt.Errorf("already_exists")
	}

	if status != 200 && status != 201 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	return extractIDFromResponse(resp)
}

// GetClientUserByEmail busca um client pelo email
func (c *APIClientV2) GetClientUserByEmail(email string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/admin/client-user", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	clients := extractArrayFromResponse(resp)
	for _, cl := range clients {
		if client, ok := cl.(map[string]interface{}); ok {
			if clientEmail, ok := client["email"].(string); ok && clientEmail == email {
				if id, ok := client["id"].(string); ok {
					return uuid.Parse(id)
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetUserByEmail busca um usuário pelo email
func (c *APIClientV2) GetUserByEmail(email string) (uuid.UUID, error) {
	resp, status, err := c.doRequest("GET", "/client-user", nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	users := extractArrayFromResponse(resp)
	for _, u := range users {
		if user, ok := u.(map[string]interface{}); ok {
			if userEmail, ok := user["email"].(string); ok && userEmail == email {
				if id, ok := user["id"].(string); ok {
					return uuid.Parse(id)
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

	if isAlreadyExistsError(resp, status) {
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

	customers := extractArrayFromResponse(resp)
	for _, cust := range customers {
		if customer, ok := cust.(map[string]interface{}); ok {
			if custEmail, ok := customer["email"].(string); ok && custEmail == email {
				if id, ok := customer["id"].(string); ok {
					return uuid.Parse(id)
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

	if isAlreadyExistsError(resp, respStatus) {
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

	reservations := extractArrayFromResponse(resp)
	for _, res := range reservations {
		if reservation, ok := res.(map[string]interface{}); ok {
			if key, ok := reservation["confirmation_key"].(string); ok && key == confirmationKey {
				if id, ok := reservation["id"].(string); ok {
					return uuid.Parse(id)
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

	if isAlreadyExistsError(resp, status) {
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

	tags := extractArrayFromResponse(resp)
	for _, t := range tags {
		if tag, ok := t.(map[string]interface{}); ok {
			if tagName, ok := tag["name"].(string); ok && tagName == name {
				if id, ok := tag["id"].(string); ok {
					return uuid.Parse(id)
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// AddCategoryToSubcategory vincula subcategoria a uma categoria (relacionamento N:M)
func (c *APIClientV2) AddCategoryToSubcategory(subcatID, catID string) error {
	path := fmt.Sprintf("/admin/subcategory/%s/category/%s", subcatID, catID)

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
	path := fmt.Sprintf("/product/%s/tags", productID)

	// Backend espera JSON body com tag_id (mesmo com path param)
	payload := map[string]interface{}{
		"tag_id": tagID,
	}

	resp, status, err := c.doRequest("POST", path, payload)
	if err != nil {
		return err
	}

	if status == 409 {
		return nil // Relacionamento já existe, ignorar
	}

	// Tratar duplicate key como "já existe"
	if status == 500 {
		if details, ok := resp["details"].(string); ok && strings.Contains(details, "duplicate key") {
			return nil
		}
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

	// Backend usa PUT /settings (não POST)
	_, status, err := c.doRequest("PUT", "/settings", payload)
	if err != nil {
		return err
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

	// Backend usa /notification/template (não /notification-template)
	resp, status, err := c.doRequest("POST", "/notification/template", payload)
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

	// Backend usa /project/settings/theme (não /theme-customization)
	_, status, err := c.doRequest("POST", "/project/settings/theme", payload)
	if err != nil {
		return err
	}

	if status == 409 {
		// Theme já existe, tentar atualizar
		_, status, err = c.doRequest("PUT", "/project/settings/theme", payload)
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
	path := fmt.Sprintf("/notification/templates/%s/%s", c.orgID, c.projID)
	resp, status, err := c.doRequest("GET", path, nil)
	if err != nil {
		return uuid.Nil, err
	}

	if status != 200 {
		return uuid.Nil, fmt.Errorf("status %d", status)
	}

	templates := extractArrayFromResponse(resp)
	for _, t := range templates {
		if template, ok := t.(map[string]interface{}); ok {
			if tName, ok := template["name"].(string); ok && tName == name {
				if id, ok := template["id"].(string); ok {
					return uuid.Parse(id)
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("não encontrado")
}

// GetPackages obtém lista de pacotes disponíveis
func (c *APIClientV2) GetPackages() ([]map[string]interface{}, error) {
	resp, status, err := c.doRequest("GET", "/package", nil)
	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, fmt.Errorf("status %d", status)
	}

	// Extrair array de pacotes
	data := extractArrayFromResponse(resp)
	packages := make([]map[string]interface{}, 0)
	for _, p := range data {
		if pkg, ok := p.(map[string]interface{}); ok {
			packages = append(packages, pkg)
		}
	}
	return packages, nil
}

// GetPackageByCodeName busca um pacote pelo code_name (enterprise, professional, starter, free)
func (c *APIClientV2) GetPackageByCodeName(codeName string) (string, error) {
	packages, err := c.GetPackages()
	if err != nil {
		return "", err
	}

	for _, pkg := range packages {
		if code, ok := pkg["code_name"].(string); ok && code == codeName {
			if id, ok := pkg["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("pacote '%s' não encontrado", codeName)
}

// SubscribeToPackage assina a organização em um pacote
// billingCycle: "monthly" ou "yearly"
func (c *APIClientV2) SubscribeToPackage(packageID, billingCycle string) error {
	payload := map[string]interface{}{
		"package_id":    packageID,
		"billing_cycle": billingCycle,
	}

	_, status, err := c.doRequest("POST", "/package/subscribe", payload)
	if err != nil {
		return err
	}

	if status == 409 {
		return nil // Já assinado
	}

	if status != 200 && status != 201 {
		return fmt.Errorf("status %d", status)
	}

	return nil
}

// GetCurrentSubscription obtém a assinatura atual da organização
func (c *APIClientV2) GetCurrentSubscription() (map[string]interface{}, error) {
	resp, status, err := c.doRequest("GET", "/package/subscription", nil)
	if err != nil {
		return nil, err
	}

	if status == 404 {
		return nil, fmt.Errorf("sem_assinatura")
	}

	if status != 200 {
		return nil, fmt.Errorf("status %d", status)
	}

	if data, ok := resp["data"].(map[string]interface{}); ok {
		return data, nil
	}

	return nil, fmt.Errorf("formato de resposta inválido")
}

// EnsureEnterpriseSubscription garante que a organização tem assinatura Enterprise
func (c *APIClientV2) EnsureEnterpriseSubscription() error {
	// 1. Verificar assinatura atual
	sub, err := c.GetCurrentSubscription()
	if err == nil && sub != nil {
		// Verificar se já é Enterprise
		if pkgData, ok := sub["package"].(map[string]interface{}); ok {
			if codeName, ok := pkgData["code_name"].(string); ok && codeName == "enterprise" {
				c.logger.Info("Organização já possui plano Enterprise")
				return nil
			}
		}
	}

	// 2. Buscar ID do pacote Enterprise
	enterpriseID, err := c.GetPackageByCodeName("enterprise")
	if err != nil {
		return fmt.Errorf("erro ao buscar pacote Enterprise: %w", err)
	}

	// 3. Assinar Enterprise
	err = c.SubscribeToPackage(enterpriseID, "yearly")
	if err != nil {
		return fmt.Errorf("erro ao assinar Enterprise: %w", err)
	}

	c.logger.Info("Assinatura Enterprise ativada com sucesso")
	return nil
}

// isAlreadyExistsError verifica se a resposta indica que o recurso já existe
// Trata tanto status 409 quanto status 500 com "already_exists" no body
func isAlreadyExistsError(resp map[string]interface{}, status int) bool {
	if status == 409 {
		return true
	}

	// Verificar se status 500 contém "already_exists" nos detalhes
	if status == 500 {
		if details, ok := resp["details"].(string); ok {
			if strings.Contains(details, "already_exists") {
				return true
			}
		}
		if msg, ok := resp["message"].(string); ok {
			if strings.Contains(strings.ToLower(msg), "already exists") {
				return true
			}
		}
		if errMsg, ok := resp["error"].(string); ok {
			if strings.Contains(strings.ToLower(errMsg), "already exists") {
				return true
			}
		}
	}

	return false
}

// isOrgAlreadyExistsError verifica se o erro indica que a organização já existe
func isOrgAlreadyExistsError(resp map[string]interface{}, status int) bool {
	if status == 400 || status == 500 {
		for _, key := range []string{"message", "details", "error"} {
			if msg, ok := resp[key].(string); ok {
				lower := strings.ToLower(msg)
				if strings.Contains(lower, "já existe") || strings.Contains(lower, "already exists") || strings.Contains(lower, "duplicate key") {
					return true
				}
			}
		}
	}
	return false
}

// extractArrayFromResponse extrai array da resposta JSON (pode vir em "data", "items" ou diretamente)
func extractArrayFromResponse(resp map[string]interface{}) []interface{} {
	// Tentar extrair do campo "data"
	if data, ok := resp["data"].([]interface{}); ok {
		return data
	}

	// Tentar extrair do campo "items" (usado quando backend retorna array direto)
	if items, ok := resp["items"].([]interface{}); ok {
		return items
	}

	// Tentar como array direto na resposta (fallback)
	for _, v := range resp {
		if arr, ok := v.([]interface{}); ok {
			return arr
		}
	}

	return nil
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

// ==================== Staff Sales CSV Import ====================

// ImportCSVResponse representa a resposta da importação de CSV
type ImportCSVResponse struct {
	BatchId      string   `json:"batch_id"`
	RecordsCount int      `json:"records_count"`
	Status       string   `json:"status"`
	Errors       []string `json:"errors,omitempty"`
}

// convertLatin1ToUTF8 converte conteúdo de Latin1/Windows-1252 para UTF-8
func convertLatin1ToUTF8(content []byte) ([]byte, error) {
	reader := transform.NewReader(
		bytes.NewReader(content),
		charmap.Windows1252.NewDecoder(),
	)
	return io.ReadAll(reader)
}

// ImportStaffSalesCSV importa registros de vendas via CSV
func (c *APIClientV2) ImportStaffSalesCSV(fileName string, csvContent []byte) (*ImportCSVResponse, error) {
	// Converter de Windows-1252 (Latin1) para UTF-8
	utf8Content, err := convertLatin1ToUTF8(csvContent)
	if err != nil {
		c.logger.Debug(fmt.Sprintf("Aviso: falha na conversão de encoding, usando original: %v", err))
		utf8Content = csvContent
	}

	base64Content := base64.StdEncoding.EncodeToString(utf8Content)

	payload := map[string]string{
		"file_name":    fileName,
		"file_content": base64Content,
	}

	resp, status, err := c.doRequest("POST", "/staff/dashboard/import", payload)
	if err != nil {
		return nil, err
	}

	if status != 200 {
		if errMsg, ok := resp["message"].(string); ok {
			return nil, fmt.Errorf("status %d: %s", status, errMsg)
		}
		return nil, fmt.Errorf("status %d", status)
	}

	result := &ImportCSVResponse{}
	if batchId, ok := resp["batch_id"].(string); ok {
		result.BatchId = batchId
	}
	if count, ok := resp["records_count"].(float64); ok {
		result.RecordsCount = int(count)
	}
	if statusStr, ok := resp["status"].(string); ok {
		result.Status = statusStr
	}
	if errors, ok := resp["errors"].([]interface{}); ok {
		for _, e := range errors {
			if s, ok := e.(string); ok {
				result.Errors = append(result.Errors, s)
			}
		}
	}

	return result, nil
}
