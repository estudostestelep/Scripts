package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"
)

// detectMimeType retorna o MIME type com base na extensão do arquivo
func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// MigrationClient é o cliente HTTP para operações de migração
type MigrationClient struct {
	baseURL      string
	token        string
	orgID        string
	projID       string
	logger       *Logger
	httpClient   *http.Client
	showPayloads bool
}

// NewMigrationClient cria um novo cliente de migração
func NewMigrationClient(baseURL string, timeout int, logger *Logger, showPayloads bool) *MigrationClient {
	return &MigrationClient{
		baseURL:      baseURL,
		logger:       logger,
		showPayloads: showPayloads,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// SetOrgProject define os headers multi-tenant
func (c *MigrationClient) SetOrgProject(orgID, projID string) {
	c.orgID = orgID
	c.projID = projID
}

// ClearOrgProject remove os headers multi-tenant (para chamadas admin sem contexto)
func (c *MigrationClient) ClearOrgProject() {
	c.orgID = ""
	c.projID = ""
}

// ClearAuth remove token e headers multi-tenant (para criar nova organização como público)
func (c *MigrationClient) ClearAuth() {
	c.token = ""
	c.orgID = ""
	c.projID = ""
}

// doRequest executa uma requisição HTTP com headers de autenticação e multi-tenant
func (c *MigrationClient) doRequest(method, path string, body interface{}) (map[string]interface{}, int, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("erro ao serializar body: %w", err)
		}
		if c.showPayloads {
			c.logger.Debug("[%s %s] Payload: %s", method, path, string(data))
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.orgID != "" {
		req.Header.Set("X-Lpe-Organization-Id", c.orgID)
	}
	if c.projID != "" {
		req.Header.Set("X-Lpe-Project-Id", c.projID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("erro na requisição %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			var arrResult []interface{}
			if arrErr := json.Unmarshal(respBody, &arrResult); arrErr == nil {
				result = map[string]interface{}{"items": arrResult}
			} else {
				result = map[string]interface{}{"raw": string(respBody)}
			}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			c.logger.Debug("[%s %s] Status: %d, Body: %s", method, path, resp.StatusCode, string(respBody))
		}
	}

	return result, resp.StatusCode, nil
}

// Get executa GET e retorna o resultado
func (c *MigrationClient) Get(path string) (map[string]interface{}, error) {
	result, status, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, nil
	}
	if status < 200 || status >= 300 {
		if msg, ok := result["message"].(string); ok {
			return nil, fmt.Errorf("GET %s retornou %d: %s", path, status, msg)
		}
		return nil, fmt.Errorf("GET %s retornou status %d", path, status)
	}
	return result, nil
}

// Post executa POST e retorna o resultado
func (c *MigrationClient) Post(path string, body interface{}) (map[string]interface{}, error) {
	result, status, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		if msg, ok := result["message"].(string); ok {
			return nil, fmt.Errorf("POST %s retornou %d: %s", path, status, msg)
		}
		return nil, fmt.Errorf("POST %s retornou status %d", path, status)
	}
	return result, nil
}

// Put executa PUT e retorna o resultado
func (c *MigrationClient) Put(path string, body interface{}) (map[string]interface{}, error) {
	result, status, err := c.doRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		if msg, ok := result["message"].(string); ok {
			return nil, fmt.Errorf("PUT %s retornou %d: %s", path, status, msg)
		}
		return nil, fmt.Errorf("PUT %s retornou status %d", path, status)
	}
	return result, nil
}

// Login faz login via /login e retorna todas as organizations/projects do usuário
func (c *MigrationClient) Login(email, password string) ([]LoginProject, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/login", payload)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		if msg, ok := resp["message"].(string); ok {
			return nil, fmt.Errorf("login falhou (%d): %s", status, msg)
		}
		return nil, fmt.Errorf("login falhou com status %d", status)
	}

	// Extrair token
	if tkn, ok := resp["token"].(string); ok && tkn != "" {
		c.token = tkn
	}

	// Extrair lista de projetos (cada um tem organization_id e project_id)
	var projects []LoginProject
	rawProjects := extractArray(resp, "projects")
	for _, raw := range rawProjects {
		p, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		lp := LoginProject{}
		if v, ok := p["organization_id"].(string); ok {
			lp.OrganizationID = v
		}
		if v, ok := p["project_id"].(string); ok {
			lp.ProjectID = v
		}
		if v, ok := p["organization_name"].(string); ok {
			lp.OrganizationName = v
		}
		if v, ok := p["project_name"].(string); ok {
			lp.ProjectName = v
		}
		if v, ok := p["organization_email"].(string); ok {
			lp.OrganizationEmail = v
		}
		if lp.OrganizationID != "" && lp.ProjectID != "" {
			projects = append(projects, lp)
		}
	}

	return projects, nil
}

// AdminLogin faz login via /admin/login e armazena o token
func (c *MigrationClient) AdminLogin(email, password string) error {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/admin/login", payload)
	if err != nil {
		return err
	}
	if status != 200 {
		if msg, ok := resp["message"].(string); ok {
			return fmt.Errorf("admin login falhou (%d): %s", status, msg)
		}
		if msg, ok := resp["error"].(string); ok {
			return fmt.Errorf("admin login falhou (%d): %s", status, msg)
		}
		return fmt.Errorf("admin login falhou com status %d", status)
	}

	if tkn, ok := resp["token"].(string); ok && tkn != "" {
		c.token = tkn
	}

	return nil
}

// CreateOrganization cria uma organização no destino.
// Se já existir, tenta encontrar o orgID/projID via login ou admin API.
// Retorna orgID, projID mesmo que o login falhe — o caller faz admin login como fallback.
func (c *MigrationClient) CreateOrganization(name, email, password string) (orgID, projID string, err error) {
	payload := map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	}

	resp, status, err := c.doRequest("POST", "/create-organization", payload)
	if err != nil {
		return "", "", err
	}

	// Organização já existe: tentar login para obter orgID/projID
	if status == 409 || isConflictError(resp, status) {
		c.logger.Info("Organização '%s' já existe no destino, fazendo login...", name)
		oID, pID, loginErr := c.loginForOrg(email, password)
		if loginErr != nil {
			c.logger.Warn("Login com email da org falhou: %v — será necessário admin login", loginErr)
			return "", "", fmt.Errorf("org já existe e login falhou: %w", loginErr)
		}
		return oID, pID, nil
	}

	if status != 200 && status != 201 {
		if msg, ok := resp["message"].(string); ok {
			return "", "", fmt.Errorf("criar organização falhou (%d): %s", status, msg)
		}
		return "", "", fmt.Errorf("criar organização falhou com status %d", status)
	}

	orgID, projID, err = extractOrgProjFromCreateResponse(resp)
	if err != nil {
		return "", "", err
	}

	// Tentar login para obter token — se falhar, o caller faz admin login
	c.logger.Info("Organização criada, fazendo login para obter token...")
	if _, loginErr := c.Login(email, password); loginErr != nil {
		c.logger.Warn("Login pós-criação falhou: %v — será necessário admin login", loginErr)
	}

	return orgID, projID, nil
}

// loginForOrg faz login e retorna o primeiro orgID/projID encontrado
func (c *MigrationClient) loginForOrg(email, password string) (orgID, projID string, err error) {
	projects, err := c.Login(email, password)
	if err != nil {
		return "", "", err
	}
	if len(projects) == 0 {
		return "", "", fmt.Errorf("login retornou sem projetos")
	}
	return projects[0].OrganizationID, projects[0].ProjectID, nil
}

// FindOrganizationByName busca uma organização pelo nome via GET /organization/
// Requer token admin. Retorna orgID, projID do primeiro projeto encontrado.
func (c *MigrationClient) FindOrganizationByName(name string) (orgID, projID string, err error) {
	resp, err := c.Get("/organization/")
	if err != nil {
		return "", "", fmt.Errorf("erro ao listar organizações: %w", err)
	}

	// Resposta pode ser {"items": [...]} ou {"data": [...]} ou array direto
	var orgs []interface{}
	if items, ok := resp["items"].([]interface{}); ok {
		orgs = items
	} else if data, ok := resp["data"].([]interface{}); ok {
		orgs = data
	}

	for _, raw := range orgs {
		org, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		orgName, _ := org["name"].(string)
		if strings.EqualFold(orgName, name) {
			oID, _ := org["id"].(string)
			if oID == "" {
				continue
			}
			// Buscar projetos da org
			projResp, projErr := c.Get("/project/organization/" + oID)
			if projErr == nil {
				var projects []interface{}
				if items, ok := projResp["items"].([]interface{}); ok {
					projects = items
				} else if data, ok := projResp["data"].([]interface{}); ok {
					projects = data
				}
				for _, rawProj := range projects {
					proj, ok := rawProj.(map[string]interface{})
					if !ok {
						continue
					}
					pID, _ := proj["id"].(string)
					if pID != "" {
						return oID, pID, nil
					}
				}
			}
			return oID, "", fmt.Errorf("organização '%s' encontrada (ID: %s) mas sem projetos", name, oID)
		}
	}

	return "", "", fmt.Errorf("organização '%s' não encontrada", name)
}

// isConflictError detecta erros de duplicidade na resposta
func isConflictError(resp map[string]interface{}, status int) bool {
	if status == 400 || status == 500 {
		msg, _ := resp["message"].(string)
		errStr, _ := resp["error"].(string)
		for _, s := range []string{strings.ToLower(msg), strings.ToLower(errStr)} {
			if contains(s, "já existe") || contains(s, "duplicate") || contains(s, "already exists") || contains(s, "unique") {
				return true
			}
		}
	}
	return false
}

// extractOrgProjFromCreateResponse extrai orgID e projID da resposta de /create-organization.
// Formato esperado: {"organization": {"id": "..."}, "project": {"id": "..."}, ...}
func extractOrgProjFromCreateResponse(resp map[string]interface{}) (orgID, projID string, err error) {
	// Formato principal: resp["organization"]["id"] e resp["project"]["id"]
	if org, ok := resp["organization"].(map[string]interface{}); ok {
		orgID, _ = org["id"].(string)
	}
	if proj, ok := resp["project"].(map[string]interface{}); ok {
		projID, _ = proj["id"].(string)
	}

	// Fallback: dentro de resp["data"]
	if orgID == "" || projID == "" {
		if data, ok := resp["data"].(map[string]interface{}); ok {
			if orgID == "" {
				if org, ok := data["organization"].(map[string]interface{}); ok {
					orgID, _ = org["id"].(string)
				}
			}
			if projID == "" {
				if proj, ok := data["project"].(map[string]interface{}); ok {
					projID, _ = proj["id"].(string)
				}
			}
		}
	}

	if orgID == "" || projID == "" {
		return "", "", fmt.Errorf("não foi possível extrair IDs da resposta de criação de organização")
	}

	return orgID, projID, nil
}

// extractArray extrai um slice de interface{} de um mapa em uma chave específica
func extractArray(resp map[string]interface{}, key string) []interface{} {
	if v, ok := resp[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

// contains verifica se s contém substr
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// DownloadImage baixa os bytes de uma imagem a partir de uma URL absoluta
func DownloadImage(imageURL string, timeout int) ([]byte, error) {
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("download falhou: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("imagem não encontrada (404)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download retornou status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler bytes da imagem: %w", err)
	}
	return data, nil
}

// UploadImage envia uma imagem via multipart/form-data para POST /upload/:category/image
// e retorna a nova image_url gerada pelo servidor de destino.
func (c *MigrationClient) UploadImage(imageBytes []byte, filename, category string) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	mimeType := detectMimeType(filename)
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, filename))
	partHeader.Set("Content-Type", mimeType)

	fw, err := w.CreatePart(partHeader)
	if err != nil {
		return "", fmt.Errorf("erro ao criar form file: %w", err)
	}
	if _, err := fw.Write(imageBytes); err != nil {
		return "", fmt.Errorf("erro ao escrever bytes: %w", err)
	}
	w.Close()

	uploadPath := "/upload/" + category + "/image"
	req, err := http.NewRequest("POST", c.baseURL+uploadPath, &buf)
	if err != nil {
		return "", fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.orgID != "" {
		req.Header.Set("X-Lpe-Organization-Id", c.orgID)
	}
	if c.projID != "" {
		req.Header.Set("X-Lpe-Project-Id", c.projID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro no upload: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload retornou status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("erro ao parsear resposta do upload: %w", err)
	}

	// Resposta: {"data": {"image_url": "..."}}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if url, ok := data["image_url"].(string); ok && url != "" {
			return url, nil
		}
	}
	// Fallback direto na raiz
	if url, ok := result["image_url"].(string); ok && url != "" {
		return url, nil
	}

	return "", fmt.Errorf("image_url não encontrada na resposta do upload")
}

// imageCategoryFromURL extrai a categoria a partir de uma URL de imagem LEP.
// Formato esperado: /uploads/{orgId}/{projId}/{category}/{filename}
func imageCategoryFromURL(imageURL string) string {
	// Remover query string se houver
	if idx := strings.Index(imageURL, "?"); idx >= 0 {
		imageURL = imageURL[:idx]
	}
	parts := strings.Split(strings.Trim(imageURL, "/"), "/")
	// /uploads/{orgId}/{projId}/{category}/{filename} → índice 3
	if len(parts) >= 5 && parts[0] == "uploads" {
		return parts[3]
	}
	// /static/{category}/{filename} → índice 1
	if len(parts) >= 3 && parts[0] == "static" {
		return parts[1]
	}
	return "products" // fallback padrão
}

// imageFilenameFromURL extrai o nome do arquivo de uma URL de imagem
func imageFilenameFromURL(imageURL string) string {
	if idx := strings.Index(imageURL, "?"); idx >= 0 {
		imageURL = imageURL[:idx]
	}
	return filepath.Base(imageURL)
}

// LoginProject representa uma entrada de org/projeto da resposta do login
type LoginProject struct {
	OrganizationID    string
	OrganizationName  string
	OrganizationEmail string
	ProjectID         string
	ProjectName       string
}
