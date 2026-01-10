package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL    string
	token      string
	orgID      string
	projID     string
	logger     *Logger
	client     *http.Client
	lastStatus int // Armazenar último status HTTP
}

func NewAPIClient(baseURL string, logger *Logger) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		logger:  logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) SetHeaders(orgID, projID, token string) {
	c.orgID = orgID
	c.projID = projID
	c.token = token
}

// GetLastStatus retorna o último status HTTP recebido
func (c *APIClient) GetLastStatus() int {
	return c.lastStatus
}

// Request faz requisição HTTP com logging detalhado
func (c *APIClient) Request(method, path string, body interface{}, requiresAuth bool) (map[string]interface{}, error) {
	url := c.baseURL + path

	// Preparar body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("Erro ao marshal body: %v", err)
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
		c.logger.Debug("Request body: %s", string(jsonBody))
	}

	// Criar request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		c.logger.Error("Erro ao criar request: %v", err)
		return nil, err
	}

	// Adicionar headers
	req.Header.Set("Content-Type", "application/json")

	if requiresAuth {
		if c.token == "" {
			c.logger.Error("Token não disponível para requisição autenticada")
			return nil, fmt.Errorf("token não disponível")
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		c.logger.Debug("Headers: Authorization: Bearer %s...", c.token[:20])
	}

	// Headers multi-tenant (exceto para login)
	if c.orgID != "" && c.projID != "" && requiresAuth {
		req.Header.Set("X-Lpe-Organization-Id", c.orgID)
		req.Header.Set("X-Lpe-Project-Id", c.projID)
		c.logger.Debug("Headers: Org-Id: %s, Proj-Id: %s", c.orgID[:8], c.projID[:8])
	}

	// Log da requisição
	c.logger.Info("%s %s", method, path)

	// Executar request
	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("Erro ao executar request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Ler response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Erro ao ler response: %v", err)
		return nil, err
	}

	// Parse JSON response - pode ser objeto, array ou string
	var result map[string]interface{}
	if len(respBody) > 0 {
		// Tentar como objeto primeiro
		if err := json.Unmarshal(respBody, &result); err != nil {
			// Se for array, armazenar como _array
			var arrayResult []interface{}
			if err := json.Unmarshal(respBody, &arrayResult); err != nil {
				// Se não é JSON válido, tentar como string
				result = map[string]interface{}{"raw": string(respBody)}
			} else {
				// Sucesso ao parsear como array - armazenar em campo especial
				result = map[string]interface{}{"_array": arrayResult}
			}
		}
	}

	// Armazenar status para acesso posterior
	c.lastStatus = resp.StatusCode

	// Log do resultado
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if statusOK {
		c.logger.Success("%s %s (status: %d, %dms)", method, path, resp.StatusCode, duration.Milliseconds())
	} else {
		c.logger.Warn("%s %s (status: %d, %dms)", method, path, resp.StatusCode, duration.Milliseconds())
	}

	c.logger.Debug("Response: %s", string(respBody))

	// Se não foi sucesso, retornar erro com a mensagem da API
	if !statusOK {
		if errMsg, ok := result["message"].(string); ok {
			return result, fmt.Errorf("status %d: %s", resp.StatusCode, errMsg)
		}
		return result, fmt.Errorf("status %d", resp.StatusCode)
	}

	return result, nil
}

// Helper para extrair dados específicos da resposta
func (c *APIClient) ExtractData(resp map[string]interface{}) map[string]interface{} {
	if data, ok := resp["data"].(map[string]interface{}); ok {
		return data
	}
	return resp
}

func (c *APIClient) ExtractString(resp map[string]interface{}, key string) string {
	if val, ok := resp[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (c *APIClient) ExtractMap(resp map[string]interface{}, key string) map[string]interface{} {
	if val, ok := resp[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// ExtractBool retorna um valor booleano da resposta
func (c *APIClient) ExtractBool(resp map[string]interface{}, key string) bool {
	if val, ok := resp[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// ExtractArray retorna um array da resposta
// Pode vir de três formas:
// 1. Array direto (armazenado em "_array")
// 2. Campo "data" contendo array
// 3. Campo específico contendo array
func (c *APIClient) ExtractArray(resp map[string]interface{}, keys ...string) []interface{} {
	// Primeiro, verificar se é um array direto (armazenado como _array)
	if arr, ok := resp["_array"].([]interface{}); ok {
		return arr
	}

	// Se nenhuma chave especificada, tentar "data"
	if len(keys) == 0 {
		keys = []string{"data"}
	}

	// Procurar a chave especificada
	for _, key := range keys {
		if val, ok := resp[key]; ok {
			if arr, ok := val.([]interface{}); ok {
				return arr
			}
		}
	}

	return nil
}

// RequestWithFile faz requisição HTTP com arquivo (multipart/form-data)
func (c *APIClient) RequestWithFile(method, path string, fileData []byte, fileName string, contentType string, requiresAuth bool) (map[string]interface{}, error) {
	url := c.baseURL + path

	// Criar buffer para multipart
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Adicionar arquivo com header de content-type
	header := make(map[string][]string)
	header["Content-Disposition"] = []string{`form-data; name="image"; filename="` + fileName + `"`}
	header["Content-Type"] = []string{contentType}

	part, err := writer.CreatePart(header)
	if err != nil {
		c.logger.Error("Erro ao criar form part: %v", err)
		return nil, err
	}

	_, err = part.Write(fileData)
	if err != nil {
		c.logger.Error("Erro ao escrever arquivo: %v", err)
		return nil, err
	}

	// Fechar writer (importante para finalizar multipart)
	err = writer.Close()
	if err != nil {
		c.logger.Error("Erro ao fechar writer: %v", err)
		return nil, err
	}

	// Criar request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		c.logger.Error("Erro ao criar request: %v", err)
		return nil, err
	}

	// Adicionar headers
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if requiresAuth {
		if c.token == "" {
			c.logger.Error("Token não disponível para requisição autenticada")
			return nil, fmt.Errorf("token não disponível")
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		c.logger.Debug("Headers: Authorization: Bearer %s...", c.token[:20])
	}

	// Headers multi-tenant
	if c.orgID != "" && c.projID != "" && requiresAuth {
		req.Header.Set("X-Lpe-Organization-Id", c.orgID)
		req.Header.Set("X-Lpe-Project-Id", c.projID)
		c.logger.Debug("Headers: Org-Id: %s, Proj-Id: %s", c.orgID[:8], c.projID[:8])
	}

	// Log da requisição
	c.logger.Info("%s %s (file: %s)", method, path, fileName)

	// Executar request
	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("Erro ao executar request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Ler response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Erro ao ler response: %v", err)
		return nil, err
	}

	// Parse JSON response
	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			var arrayResult []interface{}
			if err := json.Unmarshal(respBody, &arrayResult); err != nil {
				result = map[string]interface{}{"raw": string(respBody)}
			} else {
				result = map[string]interface{}{"_array": arrayResult}
			}
		}
	}

	// Armazenar status
	c.lastStatus = resp.StatusCode

	// Log do resultado
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if statusOK {
		c.logger.Success("%s %s (status: %d, %dms)", method, path, resp.StatusCode, duration.Milliseconds())
	} else {
		c.logger.Warn("%s %s (status: %d, %dms)", method, path, resp.StatusCode, duration.Milliseconds())
	}

	c.logger.Debug("Response: %s", string(respBody))

	// Se não foi sucesso, retornar erro
	if !statusOK {
		if errMsg, ok := result["message"].(string); ok {
			return result, fmt.Errorf("status %d: %s", resp.StatusCode, errMsg)
		}
		return result, fmt.Errorf("status %d", resp.StatusCode)
	}

	return result, nil
}
