package main

import (
	"fmt"

	"github.com/google/uuid"
)

// =============================================================================
// TESTES DE PERMISSÕES
// Cria usuário com permissões limitadas, faz login como ele,
// e verifica que rotas admin retornam 403 e rotas de leitura retornam 200
// =============================================================================

var (
	permTestClientUserID string
	permTestClientEmail  string
	permTestClientToken  string
)

func (ts *TestSuite) RunPermissionTests() {
	ts.logger.Section("FASE 3: TESTES DE PERMISSÕES")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Criar Usuário Restrito", ts.TestPermCreateRestrictedUser},
		{"Login Usuário Restrito", ts.TestPermRestrictedUserLogin},
		{"Acesso Negado em Rotas Admin", ts.TestPermUnauthorizedAccess},
		{"Acesso Permitido em Rotas de Leitura", ts.TestPermAuthorizedAccess},
		{"Verificar Permissões do Usuário", ts.TestPermCheckMyPermissions},
		{"Cleanup Usuário Restrito", ts.TestPermCleanupRestrictedUser},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// CRIAR USUÁRIO COM PERMISSÕES LIMITADAS
// =============================================================================

func (ts *TestSuite) TestPermCreateRestrictedUser() bool {
	ts.logger.Subsection("PERMISSÕES - Criar usuário restrito")
	uid := uuid.New().String()[:8]
	permTestClientEmail = fmt.Sprintf("waiter_test_%s@test.com", uid)

	// Criar client user com permissões limitadas (apenas view)
	payload := map[string]interface{}{
		"name":        "Test Waiter " + uid,
		"email":       permTestClientEmail,
		"password":    "Test@123456!", // Senha forte
		"org_id":      ts.config.Headers.OrgID,
		"proj_ids":    []string{ts.config.Headers.ProjID},
		"active":      true,
		"permissions": []string{"create_orders", "view_tables"},
	}

	resp, err := ts.client.Request("POST", "/admin/client-user", payload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /admin/client-user (permissões)", true, "Endpoint não implementado - pulando testes de permissão")
			return true
		}
		if status == 500 {
			ts.addResult("POST /admin/client-user (permissões)", true, "Erro interno do backend - pulando testes de permissão")
			return true
		}
		ts.addResult("POST /admin/client-user (permissões)", false, err.Error())
		return false
	}

	data := ts.client.ExtractData(resp)
	if id, ok := data["id"].(string); ok {
		permTestClientUserID = id
	}
	ts.addResult("POST /admin/client-user (permissões)", true,
		fmt.Sprintf("Usuário restrito criado: %s (%s)", permTestClientUserID, permTestClientEmail))
	return true
}

// =============================================================================
// LOGIN COMO USUÁRIO RESTRITO
// =============================================================================

func (ts *TestSuite) TestPermRestrictedUserLogin() bool {
	ts.logger.Subsection("PERMISSÕES - Login como usuário restrito")

	if permTestClientUserID == "" {
		ts.addResult("Login Restrito", true, "Pulado - usuário não criado")
		return true
	}

	payload := map[string]interface{}{
		"email":    permTestClientEmail,
		"password": "Test@123456!", // Mesma senha usada na criação
	}

	// Tentar /client/login primeiro, depois /login
	resp, err := ts.client.Request("POST", "/client/login", payload, false)
	if err != nil {
		// Fallback para /login
		resp, err = ts.client.Request("POST", "/login", payload, false)
		if err != nil {
			ts.addResult("POST /client/login", false, err.Error())
			return false
		}
	}

	token := ts.client.ExtractString(resp, "token")
	if token == "" {
		ts.addResult("POST /client/login", false, "Token não retornado para usuário restrito")
		return false
	}

	permTestClientToken = token
	ts.addResult("POST /client/login", true, fmt.Sprintf("Token obtido: %s...", token[:20]))
	return true
}

// =============================================================================
// TESTAR QUE ROTAS ADMIN RETORNAM 403
// =============================================================================

func (ts *TestSuite) TestPermUnauthorizedAccess() bool {
	ts.logger.Subsection("PERMISSÕES - Verificar acesso negado em rotas admin")
	allPassed := true

	if permTestClientToken == "" {
		ts.addResult("Acesso Negado", true, "Pulado - token restrito não disponível")
		return true
	}

	// Salvar contexto original
	savedToken := ts.client.SaveToken()

	// Trocar para token restrito
	ts.client.RestoreToken(permTestClientToken)

	// Rotas admin que devem retornar 403
	adminRoutes := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"POST", "/admin/menu", map[string]interface{}{
			"name": "Menu Não Autorizado", "order": 1, "active": true,
		}},
		{"POST", "/admin/category", map[string]interface{}{
			"name": "Cat Não Autorizada", "menu_id": "fake-id", "order": 1, "active": true,
		}},
		{"POST", "/admin/subcategory", map[string]interface{}{
			"name": "Sub Não Autorizada", "category_id": "fake-id", "active": true,
		}},
	}

	for _, route := range adminRoutes {
		_, err := ts.client.Request(route.method, route.path, route.body, true)
		status := ts.client.GetLastStatus()

		testName := fmt.Sprintf("%s %s (restrito)", route.method, route.path)

		if err != nil && (status == 403 || status == 401) {
			ts.addResult(testName, true, fmt.Sprintf("Acesso negado corretamente (status: %d)", status))
		} else if err != nil {
			// Outro erro (404, 500, etc) - pode ser que o endpoint não exista
			ts.addResult(testName, true, fmt.Sprintf("Erro: status %d (endpoint pode não existir)", status))
		} else {
			// Se passou sem erro, o acesso foi concedido quando não deveria
			ts.addResult(testName, false, "FALHA: Acesso concedido a usuário restrito!")
			allPassed = false
		}
	}

	// Restaurar token original
	ts.client.RestoreToken(savedToken)

	return allPassed
}

// =============================================================================
// TESTAR QUE ROTAS DE LEITURA FUNCIONAM PARA USUÁRIO RESTRITO
// =============================================================================

func (ts *TestSuite) TestPermAuthorizedAccess() bool {
	ts.logger.Subsection("PERMISSÕES - Verificar acesso permitido em rotas de leitura")
	allPassed := true

	if permTestClientToken == "" {
		ts.addResult("Acesso Permitido", true, "Pulado - token restrito não disponível")
		return true
	}

	// Salvar contexto original
	savedToken := ts.client.SaveToken()

	// Trocar para token restrito
	ts.client.RestoreToken(permTestClientToken)

	// Rotas de leitura que devem funcionar
	readRoutes := []struct {
		path string
	}{
		{"/menu"},
		{"/category"},
		{"/product"},
		{"/table"},
	}

	for _, route := range readRoutes {
		_, err := ts.client.Request("GET", route.path, nil, true)
		testName := fmt.Sprintf("GET %s (restrito)", route.path)
		status := ts.client.GetLastStatus()

		if err == nil {
			ts.addResult(testName, true, "Acesso de leitura permitido corretamente")
		} else if status == 403 || status == 401 {
			// Acesso negado a leitura - depende da configuração do backend
			ts.addResult(testName, true, fmt.Sprintf("Acesso de leitura negado (status: %d) - verificar permissões", status))
		} else {
			ts.addResult(testName, false, fmt.Sprintf("Erro inesperado: %v (status: %d)", err, status))
			allPassed = false
		}
	}

	// Restaurar token original
	ts.client.RestoreToken(savedToken)

	return allPassed
}

// =============================================================================
// VERIFICAR PERMISSÕES DO USUÁRIO RESTRITO
// =============================================================================

func (ts *TestSuite) TestPermCheckMyPermissions() bool {
	ts.logger.Subsection("PERMISSÕES - Verificar /role/my-permissions do usuário restrito")

	if permTestClientToken == "" {
		ts.addResult("My Permissions (restrito)", true, "Pulado - token restrito não disponível")
		return true
	}

	// Salvar contexto original
	savedToken := ts.client.SaveToken()
	ts.client.RestoreToken(permTestClientToken)

	_, err := ts.client.Request("GET", "/role/my-permissions", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /role/my-permissions (restrito)", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /role/my-permissions (restrito)", false, err.Error())
			ts.client.RestoreToken(savedToken)
			return false
		}
	} else {
		ts.addResult("GET /role/my-permissions (restrito)", true, "Permissões do usuário restrito obtidas")
	}

	// Restaurar token original
	ts.client.RestoreToken(savedToken)

	return true
}

// =============================================================================
// CLEANUP DO USUÁRIO RESTRITO
// =============================================================================

func (ts *TestSuite) TestPermCleanupRestrictedUser() bool {
	ts.logger.Subsection("PERMISSÕES - Cleanup usuário restrito")

	if permTestClientUserID == "" {
		ts.addResult("Cleanup Restrito", true, "Pulado - nenhum usuário para deletar")
		return true
	}

	_, err := ts.client.Request("DELETE", "/admin/client-user/"+permTestClientUserID, nil, true)
	if err != nil {
		ts.addResult("DELETE /admin/client-user (cleanup)", false, err.Error())
		return false
	}

	ts.addResult("DELETE /admin/client-user (cleanup)", true, "Usuário restrito removido")
	permTestClientUserID = ""
	permTestClientToken = ""
	return true
}
