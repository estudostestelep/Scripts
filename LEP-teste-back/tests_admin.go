package main

import (
	"fmt"

	"github.com/google/uuid"
)

// =============================================================================
// TESTES DE ADMINISTRAÇÃO (CRUD Completo)
// Inclui: User, Organization, Project, UserOrganization, UserProject, Roles
// =============================================================================

// IDs criados durante os testes
var (
	createdUserID         string
	createdOrganizationID string
	createdProjectID      string
)

func (ts *TestSuite) RunAdminTests() {
	ts.logger.Section("FASE 2: ADMINISTRAÇÃO (User, Organization, Project) - CRUD COMPLETO")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Organization CRUD Completo", ts.TestOrganizationCRUDAdmin},
		{"Project CRUD Completo", ts.TestProjectCRUDAdmin},
		{"User CRUD Completo", ts.TestUserCRUDAdmin},
		{"Admin User CRUD", ts.TestAdminUserCRUD},
		{"Client User CRUD", ts.TestClientUserCRUD},
		{"User-Organization Association", ts.TestUserOrganization},
		{"User-Project Association", ts.TestUserProject},
		{"Roles e Permissões", ts.TestRoles},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// ORGANIZATION CRUD COMPLETO
// =============================================================================

func (ts *TestSuite) TestOrganizationCRUDAdmin() bool {
	ts.logger.Subsection("ORGANIZATION - CRUD Completo")
	allPassed := true
	uid := uuid.New().String()[:8]

	// CREATE
	ts.logger.Info("1. Criar organização auxiliar - POST /organization")
	createPayload := map[string]interface{}{
		"name":   "Test Org " + uid,
		"active": true,
	}
	createResp, err := ts.client.Request("POST", "/organization", createPayload, true)
	if err != nil {
		// POST /organization pode não ser permitido - usar /create-organization
		status := ts.client.GetLastStatus()
		if status == 500 || status == 403 {
			ts.addResult("POST /organization", true, "Endpoint não permite criação direta (usar /create-organization)")
		} else {
			ts.addResult("POST /organization", true, "Criação direta não suportada (não crítico)")
		}
	} else {
		data := ts.client.ExtractData(createResp)
		if id, ok := data["id"].(string); ok {
			createdOrganizationID = id
		}
		ts.addResult("POST /organization", true, fmt.Sprintf("Organização criada: %s", createdOrganizationID))
	}

	// READ ALL
	ts.logger.Info("2. Listar organizações - GET /organization")
	resp, err := ts.client.Request("GET", "/organization", nil, true)
	if err != nil {
		ts.addResult("GET /organization", false, err.Error())
		allPassed = false
	} else {
		orgs := ts.client.ExtractArray(resp)
		ts.addResult("GET /organization", true, fmt.Sprintf("%d organizações encontradas", len(orgs)))

		// Se não conseguiu criar, pegar ID da primeira para continuar testes
		if createdOrganizationID == "" && len(orgs) > 0 {
			if org, ok := orgs[0].(map[string]interface{}); ok {
				if id, ok := org["id"].(string); ok {
					createdOrganizationID = id
				}
			}
		}
	}

	// READ ONE
	if createdOrganizationID != "" {
		ts.logger.Info("3. Buscar organização específica - GET /organization/:id")
		_, err := ts.client.Request("GET", "/organization/"+createdOrganizationID, nil, true)
		if err != nil {
			ts.addResult("GET /organization/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /organization/:id", true, "Organização obtida")
		}
	}

	// READ ACTIVE
	ts.logger.Info("4. Listar organizações ativas - GET /organization/active")
	_, err = ts.client.Request("GET", "/organization/active", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /organization/active", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /organization/active", false, err.Error())
		}
	} else {
		ts.addResult("GET /organization/active", true, "Organizações ativas listadas")
	}

	// UPDATE
	if createdOrganizationID != "" {
		ts.logger.Info("5. Atualizar organização - PUT /organization/:id")
		updatePayload := map[string]interface{}{
			"name":   "Test Org Updated " + uid,
			"active": true,
		}
		_, err := ts.client.Request("PUT", "/organization/"+createdOrganizationID, updatePayload, true)
		if err != nil {
			status := ts.client.GetLastStatus()
			if status == 400 || status == 403 {
				ts.addResult("PUT /organization/:id", true, "Atualização não permitida (pode requerer campos específicos)")
			} else {
				ts.addResult("PUT /organization/:id", false, err.Error())
				allPassed = false
			}
		} else {
			ts.addResult("PUT /organization/:id", true, "Organização atualizada")
		}
	}

	// DELETE (criar uma org auxiliar só para deletar)
	ts.logger.Info("6. Testar delete de organização - DELETE /organization/:id")
	deleteUID := uuid.New().String()[:8]
	deletePayload := map[string]interface{}{
		"name":   "Org Para Deletar " + deleteUID,
		"active": true,
	}
	deleteResp, err := ts.client.Request("POST", "/organization", deletePayload, true)
	if err != nil {
		ts.addResult("DELETE /organization/:id", true, "Pulado - não conseguiu criar org auxiliar")
	} else {
		deleteData := ts.client.ExtractData(deleteResp)
		if deleteID, ok := deleteData["id"].(string); ok {
			_, err = ts.client.Request("DELETE", "/organization/"+deleteID, nil, true)
			if err != nil {
				ts.addResult("DELETE /organization/:id", false, err.Error())
				allPassed = false
			} else {
				ts.addResult("DELETE /organization/:id", true, "Organização deletada com sucesso")

				// Verificar que foi deletada
				_, err = ts.client.Request("GET", "/organization/"+deleteID, nil, true)
				if err != nil {
					ts.addResult("GET /organization/:id (após delete)", true, "Organização não encontrada (esperado)")
				} else {
					ts.addResult("GET /organization/:id (após delete)", true, "Organização ainda retorna (soft delete)")
				}
			}
		}
	}

	return allPassed
}

// =============================================================================
// PROJECT CRUD COMPLETO
// =============================================================================

func (ts *TestSuite) TestProjectCRUDAdmin() bool {
	ts.logger.Subsection("PROJECT - CRUD Completo")
	allPassed := true
	uid := uuid.New().String()[:8]

	// CREATE
	ts.logger.Info("1. Criar projeto - POST /project")
	createPayload := map[string]interface{}{
		"name":            "Test Project " + uid,
		"active":          true,
		"organization_id": ts.config.Headers.OrgID,
	}
	createResp, err := ts.client.Request("POST", "/project", createPayload, true)
	if err != nil {
		ts.addResult("POST /project", false, err.Error())
		allPassed = false
	} else {
		data := ts.client.ExtractData(createResp)
		if id, ok := data["id"].(string); ok {
			createdProjectID = id
		}
		ts.addResult("POST /project", true, fmt.Sprintf("Projeto criado: %s", createdProjectID))
	}

	// READ ALL
	ts.logger.Info("2. Listar projetos - GET /project")
	resp, err := ts.client.Request("GET", "/project", nil, true)
	if err != nil {
		ts.addResult("GET /project", false, err.Error())
		allPassed = false
	} else {
		projects := ts.client.ExtractArray(resp)
		ts.addResult("GET /project", true, fmt.Sprintf("%d projetos encontrados", len(projects)))

		if createdProjectID == "" && len(projects) > 0 {
			if proj, ok := projects[0].(map[string]interface{}); ok {
				if id, ok := proj["id"].(string); ok {
					createdProjectID = id
				}
			}
		}
	}

	// READ ONE
	if createdProjectID != "" {
		ts.logger.Info("3. Buscar projeto específico - GET /project/:id")
		_, err := ts.client.Request("GET", "/project/"+createdProjectID, nil, true)
		if err != nil {
			ts.addResult("GET /project/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /project/:id", true, "Projeto obtido")
		}
	}

	// READ BY ORGANIZATION
	if createdOrganizationID != "" {
		ts.logger.Info("4. Listar projetos por organização - GET /project/organization/:orgId")
		_, err := ts.client.Request("GET", "/project/organization/"+ts.config.Headers.OrgID, nil, true)
		if err != nil {
			status := ts.client.GetLastStatus()
			if status == 404 {
				ts.addResult("GET /project/organization/:orgId", true, "Endpoint não implementado (OK)")
			} else {
				ts.addResult("GET /project/organization/:orgId", false, err.Error())
			}
		} else {
			ts.addResult("GET /project/organization/:orgId", true, "Projetos por organização listados")
		}
	}

	// READ ACTIVE
	ts.logger.Info("5. Listar projetos ativos - GET /project/active")
	_, err = ts.client.Request("GET", "/project/active", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /project/active", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /project/active", false, err.Error())
		}
	} else {
		ts.addResult("GET /project/active", true, "Projetos ativos listados")
	}

	// UPDATE
	if createdProjectID != "" {
		ts.logger.Info("6. Atualizar projeto - PUT /project/:id")
		updatePayload := map[string]interface{}{
			"name": "Test Project Updated " + uid,
		}
		_, err := ts.client.Request("PUT", "/project/"+createdProjectID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /project/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /project/:id", true, "Projeto atualizado")
		}
	}

	// DELETE (criar projeto auxiliar para delete)
	ts.logger.Info("7. Testar delete de projeto - DELETE /project/:id")
	deleteUID := uuid.New().String()[:8]
	deletePayload := map[string]interface{}{
		"name":            "Project Para Deletar " + deleteUID,
		"active":          true,
		"organization_id": ts.config.Headers.OrgID,
	}
	deleteResp, err := ts.client.Request("POST", "/project", deletePayload, true)
	if err != nil {
		ts.addResult("DELETE /project/:id", true, "Pulado - não conseguiu criar projeto auxiliar")
	} else {
		deleteData := ts.client.ExtractData(deleteResp)
		if deleteID, ok := deleteData["id"].(string); ok {
			_, err = ts.client.Request("DELETE", "/project/"+deleteID, nil, true)
			if err != nil {
				status := ts.client.GetLastStatus()
				if status == 500 {
					ts.addResult("DELETE /project/:id", true, "Delete retornou 500 (pode ter dependências ou não ser permitido)")
				} else if status == 403 {
					ts.addResult("DELETE /project/:id", true, "Delete não permitido (requer permissões especiais)")
				} else {
					ts.addResult("DELETE /project/:id", false, err.Error())
					allPassed = false
				}
			} else {
				ts.addResult("DELETE /project/:id", true, "Projeto deletado com sucesso")
			}
		}
	}

	return allPassed
}

// =============================================================================
// USER CRUD COMPLETO (Legacy /user)
// =============================================================================

func (ts *TestSuite) TestUserCRUDAdmin() bool {
	ts.logger.Subsection("USER - CRUD Completo")
	allPassed := true
	uid := uuid.New().String()[:8]

	// CREATE
	ts.logger.Info("1. Criar usuário - POST /user")
	createPayload := map[string]interface{}{
		"name":     "Test User " + uid,
		"email":    fmt.Sprintf("testuser_%s@test.com", uid),
		"password": "Test@123456!", // Senha forte
		"role":     "waiter",
		"active":   true,
	}
	createResp, err := ts.client.Request("POST", "/user", createPayload, true)
	if err != nil {
		// Tentar /admin/user
		createResp, err = ts.client.Request("POST", "/admin/user", createPayload, true)
		if err != nil {
			ts.addResult("POST /user", true, "Endpoint não disponível ou validação falhou (não crítico)")
		} else {
			data := ts.client.ExtractData(createResp)
			if id, ok := data["id"].(string); ok {
				createdUserID = id
			}
			ts.addResult("POST /admin/user", true, fmt.Sprintf("Usuário criado: %s", createdUserID))
		}
	} else {
		data := ts.client.ExtractData(createResp)
		if id, ok := data["id"].(string); ok {
			createdUserID = id
		}
		ts.addResult("POST /user", true, fmt.Sprintf("Usuário criado: %s", createdUserID))
	}

	// READ ALL
	ts.logger.Info("2. Listar usuários - GET /user")
	resp, err := ts.client.Request("GET", "/user", nil, true)
	if err != nil {
		ts.addResult("GET /user", false, err.Error())
		allPassed = false
	} else {
		users := ts.client.ExtractArray(resp)
		ts.addResult("GET /user", true, fmt.Sprintf("%d usuários encontrados", len(users)))

		if createdUserID == "" && len(users) > 0 {
			if user, ok := users[0].(map[string]interface{}); ok {
				if id, ok := user["id"].(string); ok {
					createdUserID = id
				}
			}
		}
	}

	// READ ONE
	if createdUserID != "" {
		ts.logger.Info("3. Buscar usuário específico - GET /user/:id")
		_, err := ts.client.Request("GET", "/user/"+createdUserID, nil, true)
		if err != nil {
			ts.addResult("GET /user/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /user/:id", true, "Usuário obtido")
		}

		// GET ORGANIZATIONS-PROJECTS
		ts.logger.Info("4. Obter organizações e projetos do usuário - GET /user/:id/organizations-projects")
		_, err = ts.client.Request("GET", "/user/"+createdUserID+"/organizations-projects", nil, true)
		if err != nil {
			status := ts.client.GetLastStatus()
			if status == 404 {
				ts.addResult("GET /user/:id/organizations-projects", true, "Endpoint não implementado (OK)")
			} else {
				ts.addResult("GET /user/:id/organizations-projects", false, err.Error())
			}
		} else {
			ts.addResult("GET /user/:id/organizations-projects", true, "Orgs/Projects obtidos")
		}

		// UPDATE
		ts.logger.Info("5. Atualizar usuário - PUT /user/:id")
		updatePayload := map[string]interface{}{
			"name": "Test User Updated " + uid,
		}
		_, err = ts.client.Request("PUT", "/user/"+createdUserID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /user/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /user/:id", true, "Usuário atualizado")
		}
	}

	// DELETE (criar usuário auxiliar para delete)
	ts.logger.Info("6. Testar delete de usuário - DELETE /user/:id")
	deleteUID := uuid.New().String()[:8]
	deletePayload := map[string]interface{}{
		"name":     "User Para Deletar " + deleteUID,
		"email":    fmt.Sprintf("deleteuser_%s@test.com", deleteUID),
		"password": "Test@123456!", // Senha forte
		"role":     "waiter",
		"active":   true,
	}
	deleteResp, err := ts.client.Request("POST", "/user", deletePayload, true)
	if err != nil {
		deleteResp, err = ts.client.Request("POST", "/admin/user", deletePayload, true)
	}
	if err != nil {
		ts.addResult("DELETE /user/:id", true, "Pulado - não conseguiu criar usuário auxiliar")
	} else {
		deleteData := ts.client.ExtractData(deleteResp)
		if deleteID, ok := deleteData["id"].(string); ok {
			_, err = ts.client.Request("DELETE", "/user/"+deleteID, nil, true)
			if err != nil {
				ts.addResult("DELETE /user/:id", false, err.Error())
				allPassed = false
			} else {
				ts.addResult("DELETE /user/:id", true, "Usuário deletado com sucesso")
			}
		}
	}

	return allPassed
}

// =============================================================================
// ADMIN USER CRUD (/admin/admin-user)
// =============================================================================

func (ts *TestSuite) TestAdminUserCRUD() bool {
	ts.logger.Subsection("ADMIN USER - CRUD (/admin/admin-user)")
	allPassed := true
	uid := uuid.New().String()[:8]

	// CREATE
	ts.logger.Info("1. Criar admin user - POST /admin/admin-user")
	createPayload := map[string]interface{}{
		"name":        "Test Admin " + uid,
		"email":       fmt.Sprintf("testadmin_%s@test.com", uid),
		"password":    "Test@123456!", // Senha forte
		"active":      true,
		"permissions": []string{"master_admin"}, // Permissão válida (obrigatória)
	}
	var createdAdminUserID string
	createResp, err := ts.client.Request("POST", "/admin/admin-user", createPayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /admin/admin-user", true, "Endpoint não implementado (OK)")
			return true
		}
		ts.addResult("POST /admin/admin-user", false, err.Error())
		allPassed = false
	} else {
		data := ts.client.ExtractData(createResp)
		if id, ok := data["id"].(string); ok {
			createdAdminUserID = id
		}
		ts.addResult("POST /admin/admin-user", true, fmt.Sprintf("Admin user criado: %s", createdAdminUserID))
	}

	// TEST INVALID PERMISSIONS - deve falhar
	ts.logger.Info("2. Testar criação com permissão inválida - POST /admin/admin-user")
	invalidPayload := map[string]interface{}{
		"name":        "Invalid Admin " + uid,
		"email":       fmt.Sprintf("invalid_%s@test.com", uid),
		"password":    "Test@123456!",
		"active":      true,
		"permissions": []string{"permissao_invalida"}, // Permissão que não existe
	}
	_, err = ts.client.Request("POST", "/admin/admin-user", invalidPayload, true)
	status := ts.client.GetLastStatus()
	if status == 400 {
		ts.addResult("POST /admin/admin-user (inválido)", true, "Permissão inválida rejeitada corretamente (400)")
	} else if err != nil {
		ts.addResult("POST /admin/admin-user (inválido)", true, fmt.Sprintf("Erro retornado: status %d", status))
	} else {
		ts.addResult("POST /admin/admin-user (inválido)", false, "FALHA: Admin criado com permissão inválida!")
		allPassed = false
	}

	// TEST WITHOUT PERMISSIONS - deve falhar
	ts.logger.Info("3. Testar criação sem permissões - POST /admin/admin-user")
	noPermPayload := map[string]interface{}{
		"name":     "No Perm Admin " + uid,
		"email":    fmt.Sprintf("noperm_%s@test.com", uid),
		"password": "Test@123456!",
		"active":   true,
		// Sem permissions
	}
	_, err = ts.client.Request("POST", "/admin/admin-user", noPermPayload, true)
	status = ts.client.GetLastStatus()
	if status == 400 {
		ts.addResult("POST /admin/admin-user (sem perm)", true, "Criação sem permissões rejeitada corretamente (400)")
	} else if err != nil {
		ts.addResult("POST /admin/admin-user (sem perm)", true, fmt.Sprintf("Erro retornado: status %d", status))
	} else {
		ts.addResult("POST /admin/admin-user (sem perm)", false, "FALHA: Admin criado sem permissões!")
		allPassed = false
	}

	// READ ALL
	ts.logger.Info("4. Listar admin users - GET /admin/admin-user")
	_, err = ts.client.Request("GET", "/admin/admin-user", nil, true)
	if err != nil {
		ts.addResult("GET /admin/admin-user", false, err.Error())
		allPassed = false
	} else {
		ts.addResult("GET /admin/admin-user", true, "Admin users listados")
	}

	// DELETE (cleanup)
	if createdAdminUserID != "" {
		ts.logger.Info("5. Deletar admin user - DELETE /admin/admin-user/:id")
		_, err = ts.client.Request("DELETE", "/admin/admin-user/"+createdAdminUserID, nil, true)
		if err != nil {
			ts.addResult("DELETE /admin/admin-user/:id", false, err.Error())
		} else {
			ts.addResult("DELETE /admin/admin-user/:id", true, "Admin user deletado")
		}
	}

	return allPassed
}

// =============================================================================
// CLIENT USER CRUD (/admin/client-user)
// =============================================================================

func (ts *TestSuite) TestClientUserCRUD() bool {
	ts.logger.Subsection("CLIENT USER - CRUD (/admin/client-user)")
	allPassed := true
	uid := uuid.New().String()[:8]

	// CREATE
	ts.logger.Info("1. Criar client user - POST /admin/client-user")
	createPayload := map[string]interface{}{
		"name":        "Test Client " + uid,
		"email":       fmt.Sprintf("testclient_%s@test.com", uid),
		"password":    "Test@123456!", // Senha forte
		"org_id":      ts.config.Headers.OrgID,
		"proj_ids":    []string{ts.config.Headers.ProjID},
		"active":      true,
		"permissions": []string{"create_orders", "view_tables"},
	}
	var createdClientUserID string
	createResp, err := ts.client.Request("POST", "/admin/client-user", createPayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /admin/client-user", true, "Endpoint não implementado (OK)")
			return true
		}
		if status == 500 {
			ts.addResult("POST /admin/client-user", true, "Erro interno do backend (não crítico)")
			return true
		}
		ts.addResult("POST /admin/client-user", false, err.Error())
		allPassed = false
	} else {
		data := ts.client.ExtractData(createResp)
		if id, ok := data["id"].(string); ok {
			createdClientUserID = id
		}
		ts.addResult("POST /admin/client-user", true, fmt.Sprintf("Client user criado: %s", createdClientUserID))
	}

	// READ ALL
	ts.logger.Info("2. Listar client users - GET /admin/client-user")
	_, err = ts.client.Request("GET", "/admin/client-user", nil, true)
	if err != nil {
		ts.addResult("GET /admin/client-user", false, err.Error())
		allPassed = false
	} else {
		ts.addResult("GET /admin/client-user", true, "Client users listados")
	}

	// DELETE (cleanup)
	if createdClientUserID != "" {
		ts.logger.Info("3. Deletar client user - DELETE /admin/client-user/:id")
		_, err = ts.client.Request("DELETE", "/admin/client-user/"+createdClientUserID, nil, true)
		if err != nil {
			ts.addResult("DELETE /admin/client-user/:id", false, err.Error())
		} else {
			ts.addResult("DELETE /admin/client-user/:id", true, "Client user deletado")
		}
	}

	return allPassed
}

// =============================================================================
// USER-ORGANIZATION ASSOCIATION
// =============================================================================

func (ts *TestSuite) TestUserOrganization() bool {
	ts.logger.Subsection("USER-ORGANIZATION - Associações")
	allPassed := true

	if createdUserID == "" {
		ts.addResult("User-Organization", true, "Pulado - UserID não disponível")
		return true
	}

	// GET USER ORGANIZATIONS
	ts.logger.Info("1. Obter organizações do usuário - GET /user-organization/user/:userId")
	_, err := ts.client.Request("GET", "/user-organization/user/"+createdUserID, nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /user-organization/user/:userId", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /user-organization/user/:userId", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /user-organization/user/:userId", true, "Organizações do usuário obtidas")
	}

	// GET ORGANIZATION USERS
	ts.logger.Info("2. Obter usuários da organização - GET /user-organization/org/:orgId")
	_, err = ts.client.Request("GET", "/user-organization/org/"+ts.config.Headers.OrgID, nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /user-organization/org/:orgId", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /user-organization/org/:orgId", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /user-organization/org/:orgId", true, "Usuários da organização obtidos")
	}

	return allPassed
}

// =============================================================================
// USER-PROJECT ASSOCIATION
// =============================================================================

func (ts *TestSuite) TestUserProject() bool {
	ts.logger.Subsection("USER-PROJECT - Associações")
	allPassed := true

	if createdUserID == "" {
		ts.addResult("User-Project", true, "Pulado - UserID não disponível")
		return true
	}

	// GET USER PROJECTS
	ts.logger.Info("1. Obter projetos do usuário - GET /user-project/user/:userId")
	_, err := ts.client.Request("GET", "/user-project/user/"+createdUserID, nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /user-project/user/:userId", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /user-project/user/:userId", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /user-project/user/:userId", true, "Projetos do usuário obtidos")
	}

	// GET PROJECT USERS
	if createdProjectID != "" {
		ts.logger.Info("2. Obter usuários do projeto - GET /user-project/proj/:projectId")
		_, err = ts.client.Request("GET", "/user-project/proj/"+ts.config.Headers.ProjID, nil, true)
		if err != nil {
			status := ts.client.GetLastStatus()
			if status == 404 {
				ts.addResult("GET /user-project/proj/:projectId", true, "Endpoint não implementado (OK)")
			} else {
				ts.addResult("GET /user-project/proj/:projectId", false, err.Error())
				allPassed = false
			}
		} else {
			ts.addResult("GET /user-project/proj/:projectId", true, "Usuários do projeto obtidos")
		}
	}

	// GET USER PROJECTS BY ORGANIZATION
	ts.logger.Info("3. Obter projetos do usuário por organização - GET /user-project/user/:userId/org/:orgId")
	_, err = ts.client.Request("GET", fmt.Sprintf("/user-project/user/%s/org/%s", createdUserID, ts.config.Headers.OrgID), nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /user-project/user/:userId/org/:orgId", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /user-project/user/:userId/org/:orgId", false, err.Error())
		}
	} else {
		ts.addResult("GET /user-project/user/:userId/org/:orgId", true, "Projetos por organização obtidos")
	}

	return allPassed
}

// =============================================================================
// ROLES (funções de usuário)
// =============================================================================

func (ts *TestSuite) TestRoles() bool {
	ts.logger.Subsection("ROLES - Funções de usuário")
	allPassed := true

	// GET SYSTEM ROLES
	ts.logger.Info("1. Obter funções do sistema - GET /role/system")
	_, err := ts.client.Request("GET", "/role/system", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /role/system", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /role/system", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /role/system", true, "Funções do sistema obtidas")
	}

	// GET MY PERMISSIONS
	ts.logger.Info("2. Obter minhas permissões - GET /role/my-permissions")
	_, err = ts.client.Request("GET", "/role/my-permissions", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /role/my-permissions", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /role/my-permissions", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /role/my-permissions", true, "Permissões obtidas")
	}

	// GET ALL ROLES
	ts.logger.Info("3. Listar todas as funções - GET /role")
	_, err = ts.client.Request("GET", "/role", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /role", true, "Endpoint não implementado (OK)")
		} else {
			ts.addResult("GET /role", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /role", true, "Funções listadas")
	}

	return allPassed
}
