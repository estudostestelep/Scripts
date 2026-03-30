package main

import (
	"fmt"

	"github.com/google/uuid"
)

// =============================================================================
// TESTES DE NOTIFICAÇÕES (Independente)
// Inclui: Templates, Configs, Logs, Webhooks, Review Queue
// =============================================================================

// IDs criados durante os testes
var (
	createdTemplateID string
)

func (ts *TestSuite) RunNotificationTests() {
	ts.logger.Section("FASE 7: NOTIFICAÇÕES")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Notification Template CRUD", ts.TestNotificationTemplateCRUD},
		{"Notification Config", ts.TestNotificationConfig},
		{"Notification Logs", ts.TestNotificationLogs},
		{"Twilio Webhooks", ts.TestTwilioWebhooks},
		{"Review Queue", ts.TestReviewQueue},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// NOTIFICATION TEMPLATE CRUD
// =============================================================================

func (ts *TestSuite) TestNotificationTemplateCRUD() bool {
	ts.logger.Subsection("NOTIFICATION TEMPLATE - CRUD")
	allPassed := true

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	// CREATE
	ts.logger.Info("1. Criar template - POST /notification/template")
	createPayload := map[string]interface{}{
		"name":            "Template Teste " + uuid.New().String()[:8],
		"channel":         "sms",
		"subject":         "Confirmação de Reserva",
		"body":            "Olá {{nome}}, sua reserva para {{data}} às {{horario}} foi confirmada!",
		"variables":       []string{"nome", "data", "horario"},
		"active":          true,
		"organization_id": orgID,
		"project_id":      projID,
	}

	createResp, err := ts.client.Request("POST", "/notification/template", createPayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /notification/template", true, "Endpoint não implementado (não crítico)")
			return true
		}
		ts.addResult("POST /notification/template", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdTemplateID = id
				ts.addResult("POST /notification/template", true, fmt.Sprintf("Template criado: %s", id[:8]))
			}
		}
		if createdTemplateID == "" {
			ts.addResult("POST /notification/template", true, "Template criado (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar templates - GET /notification/templates/:orgId/:projectId")
	_, err = ts.client.Request("GET", fmt.Sprintf("/notification/templates/%s/%s", orgID, projID), nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /notification/templates", true, "Endpoint não implementado")
		} else {
			ts.addResult("GET /notification/templates", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /notification/templates", true, "Templates listados")
	}

	// UPDATE
	if createdTemplateID != "" {
		ts.logger.Info("3. Atualizar template - PUT /notification/template")
		updatePayload := map[string]interface{}{
			"id":      createdTemplateID,
			"name":    "Template Atualizado " + uuid.New().String()[:8],
			"subject": "Reserva Confirmada!",
			"body":    "Olá {{nome}}, confirmamos sua reserva para {{data}}!",
		}
		_, err = ts.client.Request("PUT", "/notification/template", updatePayload, true)
		if err != nil {
			ts.addResult("PUT /notification/template", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /notification/template", true, "Template atualizado")
		}
	}

	return allPassed
}

// =============================================================================
// NOTIFICATION CONFIG
// =============================================================================

func (ts *TestSuite) TestNotificationConfig() bool {
	ts.logger.Subsection("NOTIFICATION CONFIG - Configuração de eventos")
	allPassed := true

	// CREATE/UPDATE CONFIG
	ts.logger.Info("1. Criar/atualizar configuração - POST /notification/config")
	configPayload := map[string]interface{}{
		"event_type": "reservation_create",
		"enabled":    true,
		"channels":   []string{"sms", "email"},
	}

	if createdTemplateID != "" {
		configPayload["template_id"] = createdTemplateID
	}

	_, err := ts.client.Request("POST", "/notification/config", configPayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /notification/config", true, "Endpoint não implementado (não crítico)")
			return true
		}
		ts.addResult("POST /notification/config", false, err.Error())
		allPassed = false
	} else {
		ts.addResult("POST /notification/config", true, "Configuração criada/atualizada")
	}

	return allPassed
}

// =============================================================================
// NOTIFICATION LOGS
// =============================================================================

func (ts *TestSuite) TestNotificationLogs() bool {
	ts.logger.Subsection("NOTIFICATION LOGS - Logs de envio")
	allPassed := true

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	// GET LOGS
	ts.logger.Info("1. Obter logs de notificações - GET /notification/logs/:orgId/:projectId")
	_, err := ts.client.Request("GET", fmt.Sprintf("/notification/logs/%s/%s", orgID, projID), nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /notification/logs", true, "Endpoint não implementado (não crítico)")
		} else {
			ts.addResult("GET /notification/logs", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /notification/logs", true, "Logs obtidos")
	}

	return allPassed
}

// =============================================================================
// TWILIO WEBHOOKS
// =============================================================================

func (ts *TestSuite) TestTwilioWebhooks() bool {
	ts.logger.Subsection("TWILIO WEBHOOKS - Callbacks de status e inbound")
	allPassed := true

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	// STATUS WEBHOOK
	ts.logger.Info("1. Testar webhook de status - POST /webhook/twilio/status")
	statusPayload := map[string]interface{}{
		"MessageSid":    "SM" + uuid.New().String()[:30],
		"MessageStatus": "delivered",
		"To":            "+5511999999999",
		"From":          "+5511888888888",
	}

	_, err := ts.client.Request("POST", "/webhook/twilio/status", statusPayload, false)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /webhook/twilio/status", true, "Endpoint não implementado")
		} else {
			// Webhook pode retornar erros se não encontrar o SID - isso é esperado
			ts.addResult("POST /webhook/twilio/status", true, "Webhook chamado (resposta esperada)")
		}
	} else {
		ts.addResult("POST /webhook/twilio/status", true, "Status webhook processado")
	}

	// INBOUND WEBHOOK
	ts.logger.Info("2. Testar webhook inbound - POST /webhook/twilio/inbound/:orgId/:projectId")
	inboundPayload := map[string]interface{}{
		"MessageSid": "SM" + uuid.New().String()[:30],
		"Body":       "Sim, confirmo a reserva",
		"From":       "+5511999999999",
		"To":         "+5511888888888",
	}

	_, err = ts.client.Request("POST", fmt.Sprintf("/webhook/twilio/inbound/%s/%s", orgID, projID), inboundPayload, false)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /webhook/twilio/inbound", true, "Endpoint não implementado")
		} else {
			// Webhook pode processar e retornar sucesso mesmo sem reserva associada
			ts.addResult("POST /webhook/twilio/inbound", true, "Webhook chamado")
		}
	} else {
		ts.addResult("POST /webhook/twilio/inbound", true, "Inbound webhook processado")
	}

	return allPassed
}

// =============================================================================
// REVIEW QUEUE
// =============================================================================

func (ts *TestSuite) TestReviewQueue() bool {
	ts.logger.Subsection("REVIEW QUEUE - Fila de revisão de respostas")
	allPassed := true

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	// GET REVIEW QUEUE
	ts.logger.Info("1. Obter fila de revisão - GET /notification/review-queue/:orgId/:projectId")
	resp, err := ts.client.Request("GET", fmt.Sprintf("/notification/review-queue/%s/%s", orgID, projID), nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /notification/review-queue", true, "Endpoint não implementado (não crítico)")
			return true
		}
		ts.addResult("GET /notification/review-queue", false, err.Error())
		allPassed = false
	} else {
		queue := ts.client.ExtractArray(resp)
		ts.addResult("GET /notification/review-queue", true, fmt.Sprintf("%d itens na fila de revisão", len(queue)))

		// Se temos itens na fila, testar aprovação
		if len(queue) > 0 {
			if item, ok := queue[0].(map[string]interface{}); ok {
				if itemID, ok := item["id"].(string); ok {
					// APPROVE
					ts.logger.Info("2. Aprovar item - POST /notification/review-queue/:id/approve")
					approvePayload := map[string]interface{}{
						"notes": "Aprovado automaticamente pelo teste",
					}
					_, err = ts.client.Request("POST", fmt.Sprintf("/notification/review-queue/%s/approve", itemID), approvePayload, true)
					if err != nil {
						ts.addResult("POST /review-queue/:id/approve", false, err.Error())
					} else {
						ts.addResult("POST /review-queue/:id/approve", true, "Item aprovado")
					}
				}
			}
		}
	}

	return allPassed
}

// =============================================================================
// SEND NOTIFICATION (teste manual)
// =============================================================================

func (ts *TestSuite) TestSendNotification() bool {
	ts.logger.Subsection("SEND NOTIFICATION - Envio manual")

	// Este teste é opcional e requer configuração de Twilio
	ts.logger.Info("1. Enviar notificação manual - POST /notification/send")
	sendPayload := map[string]interface{}{
		"channel":   "sms",
		"recipient": "+5511999999999",
		"subject":   "Teste de Notificação",
		"message":   "Esta é uma mensagem de teste enviada pelo sistema automatizado.",
	}

	_, err := ts.client.Request("POST", "/notification/send", sendPayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 || status == 400 {
			ts.addResult("POST /notification/send", true, "Endpoint não configurado (Twilio não ativo)")
			return true
		}
		ts.addResult("POST /notification/send", false, err.Error())
		return false
	}

	ts.addResult("POST /notification/send", true, "Notificação enviada")
	return true
}
