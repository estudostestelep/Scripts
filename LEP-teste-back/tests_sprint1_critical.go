package main

import (
	"fmt"
)

// ============================================================================
// SPRINT 1: TESTES CRÍTICOS (34 testes)
// ============================================================================

// TestCheckTokenExtended valida um token JWT (versão estendida)
func (ts *TestSuite) TestCheckTokenExtended() bool {
	ts.logger.Subsection("POST /checkToken - Validar token JWT estendido")

	payload := map[string]interface{}{
		"token": ts.client.token,
	}

	_, err := ts.client.Request("POST", "/checkToken", payload, true)
	if err != nil {
		ts.addResult("POST /checkToken (extended)", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /checkToken (extended)", true, "Token validado com sucesso")
	return true
}

// TestTwilioStatusWebhook testa callback de status do Twilio
func (ts *TestSuite) TestTwilioStatusWebhook() bool {
	ts.logger.Subsection("POST /webhook/twilio/status - Status callback")

	payload := map[string]interface{}{
		"MessageSid":    "SMxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"MessageStatus": "delivered",
		"ErrorCode":     nil,
	}

	_, err := ts.client.Request("POST", "/webhook/twilio/status", payload, false)
	if err != nil {
		ts.addResult("POST /webhook/twilio/status", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /webhook/twilio/status", true, "Webhook de status processado")
	return true
}

// TestTwilioInboundWebhook testa mensagem inbound do Twilio
func (ts *TestSuite) TestTwilioInboundWebhook() bool {
	ts.logger.Subsection("POST /webhook/twilio/inbound - Inbound message")

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	payload := map[string]interface{}{
		"MessageSid": "SMxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"AccountSid": "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"From":       "+5511999999999",
		"To":         "+5511888888888",
		"Body":       "Mensagem de teste",
		"NumMedia":   "0",
	}

	path := fmt.Sprintf("/webhook/twilio/inbound/%s/%s", orgID, projID)
_,err := ts.client.Request("POST", path, payload, false)
	if err != nil {
		ts.addResult("POST /webhook/twilio/inbound", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /webhook/twilio/inbound", true, "Webhook inbound processado")
	return true
}

// TestCreateNotificationTemplate cria um template de notificação
func (ts *TestSuite) TestCreateNotificationTemplate() bool {
	ts.logger.Subsection("POST /notification/template - Criar template")

	payload := map[string]interface{}{
		"name":       "Template Teste",
		"type":       "sms",
		"content":    "Bem-vindo ao {restaurant_name}!",
		"variables": []string{"restaurant_name"},
	}

_,err := ts.client.Request("POST", "/notification/template", payload, true)
	if err != nil {
		ts.addResult("POST /notification/template", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /notification/template", true, "Template criado com sucesso")
	return true
}

// TestUpdateNotificationTemplate atualiza um template de notificação
func (ts *TestSuite) TestUpdateNotificationTemplate() bool {
	ts.logger.Subsection("PUT /notification/template - Atualizar template")

	payload := map[string]interface{}{
		"id":       "template-id-123",
		"name":     "Template Atualizado",
		"content":  "Conteúdo atualizado",
	}

_,err := ts.client.Request("PUT", "/notification/template", payload, true)
	if err != nil {
		ts.addResult("PUT /notification/template", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("PUT /notification/template", true, "Template atualizado")
	return true
}

// TestSendNotification envia uma notificação
func (ts *TestSuite) TestSendNotification() bool {
	ts.logger.Subsection("POST /notification/send - Enviar notificação")

	payload := map[string]interface{}{
		"type":    "sms",
		"to":      "+5511999999999",
		"message": "Sua reserva foi confirmada!",
	}

_,err := ts.client.Request("POST", "/notification/send", payload, true)
	if err != nil {
		ts.addResult("POST /notification/send", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /notification/send", true, "Notificação enviada")
	return true
}

// TestNotificationEvent registra um evento de notificação
func (ts *TestSuite) TestNotificationEvent() bool {
	ts.logger.Subsection("POST /notification/event - Registrar evento")

	payload := map[string]interface{}{
		"event_type": "reservation_confirmed",
		"entity_id":  "reservation-123",
		"data": map[string]interface{}{
			"guest_name": "João Silva",
			"date":       "2025-11-15",
		},
	}

_,err := ts.client.Request("POST", "/notification/event", payload, true)
	if err != nil {
		ts.addResult("POST /notification/event", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /notification/event", true, "Evento registrado")
	return true
}

// TestGetNotificationLogs busca logs de notificação
func (ts *TestSuite) TestGetNotificationLogs() bool {
	ts.logger.Subsection("GET /notification/logs - Obter logs")

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	path := fmt.Sprintf("/notification/logs/%s/%s", orgID, projID)
_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /notification/logs", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /notification/logs", true, "Logs obtidos com sucesso")
	return true
}

// TestGetNotificationTemplates lista templates de notificação
func (ts *TestSuite) TestGetNotificationTemplates() bool {
	ts.logger.Subsection("GET /notification/templates - Listar templates")

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	path := fmt.Sprintf("/notification/templates/%s/%s", orgID, projID)
_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /notification/templates", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /notification/templates", true, "Templates listados")
	return true
}

// TestCreateNotificationConfig cria configuração de notificação
func (ts *TestSuite) TestCreateNotificationConfig() bool {
	ts.logger.Subsection("POST /notification/config - Criar configuração")

	payload := map[string]interface{}{
		"event_type": "reservation_confirmed",
		"channels":   []string{"sms", "email"},
		"enabled":    true,
	}

_,err := ts.client.Request("POST", "/notification/config", payload, true)
	if err != nil {
		ts.addResult("POST /notification/config", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /notification/config", true, "Configuração criada")
	return true
}

// TestGetOrderProgress obter progresso do pedido
func (ts *TestSuite) TestGetOrderProgress() bool {
	ts.logger.Subsection("GET /order/:id/progress - Progresso do pedido")

	// Criar pedido de teste
	orderPayload := map[string]interface{}{
		"customer_id":     "customer-123",
		"status":          "pending",
		"total_amount":    50.0,
		"organization_id": ts.config.Headers.OrgID,
		"project_id":      ts.config.Headers.ProjID,
		"items": []map[string]interface{}{
			{
				"product_id": "product-123",
				"quantity":   1,
				"price":      50.0,
			},
		},
	}

	createResp, _ := ts.client.Request("POST", "/order", orderPayload, true)
	orderID := ts.client.ExtractString(ts.client.ExtractData(createResp), "id")

	if orderID == "" {
		ts.addResult("GET /order/:id/progress", false, "Não conseguiu criar pedido")
		return false
	}

	path := fmt.Sprintf("/order/%s/progress", orderID)
_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /order/:id/progress", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /order/:id/progress", true, "Progresso obtido")
	return true
}

// TestUpdateOrderStatus atualiza status do pedido
func (ts *TestSuite) TestUpdateOrderStatus() bool {
	ts.logger.Subsection("PUT /order/:id/status - Atualizar status")

	// Criar pedido primeiro
	orderPayload := map[string]interface{}{
		"customer_id":     "customer-123",
		"status":          "pending",
		"total_amount":    50.0,
		"organization_id": ts.config.Headers.OrgID,
		"project_id":      ts.config.Headers.ProjID,
		"items": []map[string]interface{}{
			{
				"product_id": "product-123",
				"quantity":   1,
				"price":      50.0,
			},
		},
	}

	createResp, _ := ts.client.Request("POST", "/order", orderPayload, true)
	orderID := ts.client.ExtractString(ts.client.ExtractData(createResp), "id")

	if orderID == "" {
		ts.addResult("PUT /order/:id/status", false, "Não conseguiu criar pedido")
		return false
	}

	statusPayload := map[string]interface{}{
		"status": "completed",
	}

	path := fmt.Sprintf("/order/%s/status", orderID)
_,err := ts.client.Request("PUT", path, statusPayload, true)
	if err != nil {
		ts.addResult("PUT /order/:id/status", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("PUT /order/:id/status", true, "Status atualizado")
	return true
}

// TestCreateOrganizationWithSetup cria organização com setup completo
func (ts *TestSuite) TestCreateOrganizationWithSetup() bool {
	ts.logger.Subsection("POST /create-organization - Bootstrap completo")

	payload := map[string]interface{}{
		"name":     "Novo Restaurante Teste",
		"password": "senha123456",
	}

_,err := ts.client.Request("POST", "/create-organization", payload, false)
	if err != nil {
		ts.addResult("POST /create-organization", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /create-organization", true, "Organização criada com setup")
	return true
}

// TestOrganizationSeeding testa seeding de organização
func (ts *TestSuite) TestOrganizationSeeding() bool {
	ts.logger.Subsection("POST /organization - Seeding")

	payload := map[string]interface{}{
		"name":        "Org Seeding Teste",
		"description": "Teste de seeding",
		"email":       "org-seed@test.com",
	}

	_, err := ts.client.Request("POST", "/organization", payload, true)
	if err != nil {
		ts.addResult("POST /organization (seed)", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /organization (seed)", true, "Organização seeded")
	return true
}

// TestProjectSeeding testa seeding de projeto
func (ts *TestSuite) TestProjectSeeding() bool {
	ts.logger.Subsection("POST /project - Seeding")

	payload := map[string]interface{}{
		"name":        "Projeto Seeding Teste",
		"description": "Teste de seeding",
	}

	_, err := ts.client.Request("POST", "/project", payload, true)
	if err != nil {
		ts.addResult("POST /project (seed)", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /project (seed)", true, "Projeto seeded")
	return true
}

// TestUserOrganizationSeeding testa seeding user-org
func (ts *TestSuite) TestUserOrganizationSeeding() bool {
	ts.logger.Subsection("POST /user-organization - Seeding")

	// Use the authenticated user's ID
	userID := ts.config.TestUser.Email
	path := fmt.Sprintf("/user-organization/user/%s", userID)

	payload := map[string]interface{}{
		"role": "admin",
	}

	_, err := ts.client.Request("POST", path, payload, true)
	if err != nil {
		ts.addResult("POST /user-organization (seed)", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /user-organization (seed)", true, "User-org seeded")
	return true
}

// TestUserProjectSeeding testa seeding user-project
func (ts *TestSuite) TestUserProjectSeeding() bool {
	ts.logger.Subsection("POST /user-project - Seeding")

	// Use the authenticated user's ID
	userID := ts.config.TestUser.Email
	path := fmt.Sprintf("/user-project/user/%s", userID)

	payload := map[string]interface{}{
		"role": "editor",
	}

	_, err := ts.client.Request("POST", path, payload, true)
	if err != nil {
		ts.addResult("POST /user-project (seed)", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /user-project (seed)", true, "User-project seeded")
	return true
}

// TestGetUserOrganizations lista organizações do usuário
func (ts *TestSuite) TestGetUserOrganizations() bool {
	ts.logger.Subsection("GET /user-organization/user/:userId - Listar orgs")

	userID := ts.config.TestUser.Email // Usar email como fallback
	path := fmt.Sprintf("/user-organization/user/%s", userID)

_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /user-organization/user/:userId", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /user-organization/user/:userId", true, "Orgs listadas")
	return true
}

// TestGetOrganizationUsers lista usuários da organização
func (ts *TestSuite) TestGetOrganizationUsers() bool {
	ts.logger.Subsection("GET /user-organization/org/:orgId - Listar usuários")

	orgID := ts.config.Headers.OrgID
	path := fmt.Sprintf("/user-organization/org/%s", orgID)

_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /user-organization/org/:orgId", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /user-organization/org/:orgId", true, "Usuários listados")
	return true
}

// TestRemoveUserFromOrganization remove usuário da organização
func (ts *TestSuite) TestRemoveUserFromOrganization() bool {
	ts.logger.Subsection("DELETE /user-organization - Remover acesso")

	userID := "user-123"
	orgID := ts.config.Headers.OrgID
	path := fmt.Sprintf("/user-organization/user/%s/org/%s", userID, orgID)

_,err := ts.client.Request("DELETE", path, nil, true)
	if err != nil {
		ts.addResult("DELETE /user-organization", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("DELETE /user-organization", true, "Acesso removido")
	return true
}

// TestUpdateUserOrganization atualiza relacionamento user-org
func (ts *TestSuite) TestUpdateUserOrganization() bool {
	ts.logger.Subsection("PUT /user-organization/:id - Atualizar acesso")

	payload := map[string]interface{}{
		"id":   "user-org-123",
		"role": "viewer",
	}

_,err := ts.client.Request("PUT", "/user-organization/user-org-123", payload, true)
	if err != nil {
		ts.addResult("PUT /user-organization/:id", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("PUT /user-organization/:id", true, "Acesso atualizado")
	return true
}

// TestGetUserProjects lista projetos do usuário
func (ts *TestSuite) TestGetUserProjects() bool {
	ts.logger.Subsection("GET /user-project/user/:userId - Listar projetos")

	userID := ts.config.TestUser.Email
	path := fmt.Sprintf("/user-project/user/%s", userID)

_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /user-project/user/:userId", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /user-project/user/:userId", true, "Projetos listados")
	return true
}

// TestGetUserProjectsByOrg lista projetos do usuário em uma org
func (ts *TestSuite) TestGetUserProjectsByOrg() bool {
	ts.logger.Subsection("GET /user-project/user/:userId/org/:orgId")

	userID := ts.config.TestUser.Email
	orgID := ts.config.Headers.OrgID
	path := fmt.Sprintf("/user-project/user/%s/org/%s", userID, orgID)

_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /user-project/user/:userId/org/:orgId", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /user-project/user/:userId/org/:orgId", true, "Projetos por org listados")
	return true
}

// TestGetProjectUsers lista usuários do projeto
func (ts *TestSuite) TestGetProjectUsers() bool {
	ts.logger.Subsection("GET /user-project/proj/:projectId")

	projID := ts.config.Headers.ProjID
	path := fmt.Sprintf("/user-project/proj/%s", projID)

_,err := ts.client.Request("GET", path, nil, true)
	if err != nil {
		ts.addResult("GET /user-project/proj/:projectId", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /user-project/proj/:projectId", true, "Usuários do projeto listados")
	return true
}

// TestRemoveUserFromProject remove usuário do projeto
func (ts *TestSuite) TestRemoveUserFromProject() bool {
	ts.logger.Subsection("DELETE /user-project - Remover acesso")

	userID := "user-123"
	projID := ts.config.Headers.ProjID
	path := fmt.Sprintf("/user-project/user/%s/proj/%s", userID, projID)

_,err := ts.client.Request("DELETE", path, nil, true)
	if err != nil {
		ts.addResult("DELETE /user-project", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("DELETE /user-project", true, "Acesso ao projeto removido")
	return true
}

// TestUpdateUserProject atualiza relacionamento user-project
func (ts *TestSuite) TestUpdateUserProject() bool {
	ts.logger.Subsection("PUT /user-project/:id")

	payload := map[string]interface{}{
		"id":   "user-proj-123",
		"role": "viewer",
	}

_,err := ts.client.Request("PUT", "/user-project/user-proj-123", payload, true)
	if err != nil {
		ts.addResult("PUT /user-project/:id", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("PUT /user-project/:id", true, "Acesso ao projeto atualizado")
	return true
}

// TestAdminResetPasswords testa reset de senhas de admins
func (ts *TestSuite) TestAdminResetPasswords() bool {
	ts.logger.Subsection("POST /admin/reset-passwords")

	payload := map[string]interface{}{
		"organization_id": ts.config.Headers.OrgID,
	}

_,err := ts.client.Request("POST", "/admin/reset-passwords", payload, true)
	if err != nil {
		ts.addResult("POST /admin/reset-passwords", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("POST /admin/reset-passwords", true, "Reset de senhas executado")
	return true
}

// TestPermanentDeleteOrganization testa hard delete de organização
func (ts *TestSuite) TestPermanentDeleteOrganization() bool {
	ts.logger.Subsection("DELETE /organization/:id/permanent")

	orgID := "test-org-to-delete"
	path := fmt.Sprintf("/organization/%s/permanent", orgID)

_,err := ts.client.Request("DELETE", path, nil, true)
	if err != nil {
		ts.addResult("DELETE /organization/:id/permanent", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("DELETE /organization/:id/permanent", true, "Organização permanentemente deletada")
	return true
}

// TestOccupancyReport testa relatório de ocupação
func (ts *TestSuite) TestOccupancyReport() bool {
	ts.logger.Subsection("GET /reports/occupancy")

_,err := ts.client.Request("GET", "/reports/occupancy", nil, true)
	if err != nil {
		ts.addResult("GET /reports/occupancy", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /reports/occupancy", true, "Relatório de ocupação obtido")
	return true
}

// TestReservationsReport testa relatório de reservas
func (ts *TestSuite) TestReservationsReport() bool {
	ts.logger.Subsection("GET /reports/reservations")

_,err := ts.client.Request("GET", "/reports/reservations", nil, true)
	if err != nil {
		ts.addResult("GET /reports/reservations", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /reports/reservations", true, "Relatório de reservas obtido")
	return true
}

// TestWaitlistReport testa relatório de fila de espera
func (ts *TestSuite) TestWaitlistReport() bool {
	ts.logger.Subsection("GET /reports/waitlist")

_,err := ts.client.Request("GET", "/reports/waitlist", nil, true)
	if err != nil {
		ts.addResult("GET /reports/waitlist", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /reports/waitlist", true, "Relatório de fila obtido")
	return true
}

// TestLeadsReport testa relatório de leads
func (ts *TestSuite) TestLeadsReport() bool {
	ts.logger.Subsection("GET /reports/leads")

_,err := ts.client.Request("GET", "/reports/leads", nil, true)
	if err != nil {
		ts.addResult("GET /reports/leads", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	ts.addResult("GET /reports/leads", true, "Relatório de leads obtido")
	return true
}

// RunSprintOneTests executa todos os testes críticos do Sprint 1
func (ts *TestSuite) RunSprintOneTests() {
	ts.logger.Section("🔴 SPRINT 1: TESTES CRÍTICOS (34 TESTES)")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"POST /checkToken (extended)", ts.TestCheckTokenExtended},
		{"POST /webhook/twilio/status", ts.TestTwilioStatusWebhook},
		{"POST /webhook/twilio/inbound", ts.TestTwilioInboundWebhook},
		{"POST /notification/template", ts.TestCreateNotificationTemplate},
		{"PUT /notification/template", ts.TestUpdateNotificationTemplate},
		{"POST /notification/send", ts.TestSendNotification},
		{"POST /notification/event", ts.TestNotificationEvent},
		{"GET /notification/logs", ts.TestGetNotificationLogs},
		{"GET /notification/templates", ts.TestGetNotificationTemplates},
		{"POST /notification/config", ts.TestCreateNotificationConfig},
		{"GET /order/:id/progress", ts.TestGetOrderProgress},
		{"PUT /order/:id/status", ts.TestUpdateOrderStatus},
		{"POST /create-organization", ts.TestCreateOrganizationWithSetup},
		{"POST /organization (seed)", ts.TestOrganizationSeeding},
		{"POST /project (seed)", ts.TestProjectSeeding},
		{"POST /user-organization (seed)", ts.TestUserOrganizationSeeding},
		{"POST /user-project (seed)", ts.TestUserProjectSeeding},
		{"GET /user-organization/user/:userId", ts.TestGetUserOrganizations},
		{"GET /user-organization/org/:orgId", ts.TestGetOrganizationUsers},
		{"DELETE /user-organization", ts.TestRemoveUserFromOrganization},
		{"PUT /user-organization/:id", ts.TestUpdateUserOrganization},
		{"GET /user-project/user/:userId", ts.TestGetUserProjects},
		{"GET /user-project/user/:userId/org/:orgId", ts.TestGetUserProjectsByOrg},
		{"GET /user-project/proj/:projectId", ts.TestGetProjectUsers},
		{"DELETE /user-project", ts.TestRemoveUserFromProject},
		{"PUT /user-project/:id", ts.TestUpdateUserProject},
		{"POST /admin/reset-passwords", ts.TestAdminResetPasswords},
		{"DELETE /organization/:id/permanent", ts.TestPermanentDeleteOrganization},
		{"GET /reports/occupancy", ts.TestOccupancyReport},
		{"GET /reports/reservations", ts.TestReservationsReport},
		{"GET /reports/waitlist", ts.TestWaitlistReport},
		{"GET /reports/leads", ts.TestLeadsReport},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		if test.fn() {
			passed++
		} else {
			failed++
		}
	}

	ts.logger.Stats(len(tests), passed, failed)
}
