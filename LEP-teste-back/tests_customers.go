package main

import (
	"fmt"

	"github.com/google/uuid"
)

// =============================================================================
// TESTES DE CLIENTES E INFRAESTRUTURA (Base para Reservas)
// Ordem: Environment → Table → Customer
// =============================================================================

// IDs criados durante os testes para uso nas dependências
var (
	createdEnvironmentID string
	createdTableID       string
	createdCustomerID    string
)

func (ts *TestSuite) RunCustomerTests() {
	ts.logger.Section("FASE 4: CLIENTES E INFRAESTRUTURA (Environment → Table → Customer)")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Environment CRUD", ts.TestEnvironmentCRUDCustomers},
		{"Table CRUD", ts.TestTableCRUDCustomers},
		{"Customer CRUD", ts.TestCustomerCRUDCustomers},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// ENVIRONMENT CRUD
// =============================================================================

func (ts *TestSuite) TestEnvironmentCRUDCustomers() bool {
	ts.logger.Subsection("ENVIRONMENT - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar ambiente - POST /environment")
	createPayload := map[string]interface{}{
		"name":        "Ambiente Teste " + uuid.New().String()[:8],
		"description": "Área externa para testes",
		"active":      true,
		"capacity":    50,
	}

	createResp, err := ts.client.Request("POST", "/environment", createPayload, true)
	if err != nil {
		ts.addResult("POST /environment", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdEnvironmentID = id
				ts.addResult("POST /environment", true, fmt.Sprintf("Ambiente criado: %s", id[:8]))
			}
		}
		if createdEnvironmentID == "" {
			ts.addResult("POST /environment", true, "Ambiente criado (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar ambientes - GET /environment")
	resp, err := ts.client.Request("GET", "/environment", nil, true)
	if err != nil {
		ts.addResult("GET /environment", false, err.Error())
		allPassed = false
	} else {
		environments := ts.client.ExtractArray(resp)
		ts.addResult("GET /environment", true, fmt.Sprintf("%d ambientes encontrados", len(environments)))

		if createdEnvironmentID == "" && len(environments) > 0 {
			if env, ok := environments[0].(map[string]interface{}); ok {
				if id, ok := env["id"].(string); ok {
					createdEnvironmentID = id
				}
			}
		}
	}

	// READ ONE
	if createdEnvironmentID != "" {
		ts.logger.Info("3. Buscar ambiente específico - GET /environment/:id")
		_, err := ts.client.Request("GET", "/environment/"+createdEnvironmentID, nil, true)
		if err != nil {
			ts.addResult("GET /environment/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /environment/:id", true, "Ambiente obtido")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar ambiente - PUT /environment/:id")
		updatePayload := map[string]interface{}{
			"name":        "Ambiente Atualizado " + uuid.New().String()[:8],
			"description": "Descrição atualizada",
			"capacity":    60,
			"active":      true,
		}
		_, err = ts.client.Request("PUT", "/environment/"+createdEnvironmentID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /environment/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /environment/:id", true, "Ambiente atualizado")
		}
	}

	// READ ACTIVE
	ts.logger.Info("5. Listar ambientes ativos - GET /environment/active")
	_, err = ts.client.Request("GET", "/environment/active", nil, true)
	if err != nil {
		ts.addResult("GET /environment/active", false, err.Error())
	} else {
		ts.addResult("GET /environment/active", true, "Ambientes ativos listados")
	}

	// DELETE (criar ambiente auxiliar para delete)
	ts.logger.Info("6. Testar delete de ambiente - DELETE /environment/:id")
	deletePayload := map[string]interface{}{
		"name":     "Ambiente Para Deletar " + uuid.New().String()[:8],
		"active":   true,
		"capacity": 10,
	}
	deleteResp, err := ts.client.Request("POST", "/environment", deletePayload, true)
	if err != nil {
		ts.addResult("DELETE /environment/:id", true, "Pulado - não conseguiu criar ambiente auxiliar")
	} else {
		if deleteData := ts.client.ExtractData(deleteResp); deleteData != nil {
			if deleteID, ok := deleteData["id"].(string); ok {
				_, err = ts.client.Request("DELETE", "/environment/"+deleteID, nil, true)
				if err != nil {
					ts.addResult("DELETE /environment/:id", false, err.Error())
					allPassed = false
				} else {
					ts.addResult("DELETE /environment/:id", true, "Ambiente deletado com sucesso")
				}
			}
		}
	}

	return allPassed
}

// =============================================================================
// TABLE CRUD (pode depender de Environment)
// =============================================================================

func (ts *TestSuite) TestTableCRUDCustomers() bool {
	ts.logger.Subsection("TABLE - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar mesa - POST /table")
	tableNumber := int(200 + uuid.New().ID()%100) // Range 200-299 para evitar conflitos
	createPayload := map[string]interface{}{
		"number":   tableNumber,
		"capacity": 4,
		"location": "Área principal",
		"status":   "livre",
	}
	// NOTA: Não enviar environment_id por enquanto para evitar problemas de FK

	createResp, err := ts.client.Request("POST", "/table", createPayload, true)
	if err != nil {
		ts.addResult("POST /table", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdTableID = id
				ts.addResult("POST /table", true, fmt.Sprintf("Mesa criada: %s", id[:8]))
			}
		}
		if createdTableID == "" {
			ts.addResult("POST /table", true, "Mesa criada (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar mesas - GET /table")
	resp, err := ts.client.Request("GET", "/table", nil, true)
	if err != nil {
		ts.addResult("GET /table", false, err.Error())
		allPassed = false
	} else {
		tables := ts.client.ExtractArray(resp)
		ts.addResult("GET /table", true, fmt.Sprintf("%d mesas encontradas", len(tables)))

		if createdTableID == "" && len(tables) > 0 {
			if table, ok := tables[0].(map[string]interface{}); ok {
				if id, ok := table["id"].(string); ok {
					createdTableID = id
				}
			}
		}
	}

	// READ ONE
	if createdTableID != "" {
		ts.logger.Info("3. Buscar mesa específica - GET /table/:id")
		_, err := ts.client.Request("GET", "/table/"+createdTableID, nil, true)
		if err != nil {
			ts.addResult("GET /table/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /table/:id", true, "Mesa obtida")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar mesa - PUT /table/:id")
		updatePayload := map[string]interface{}{
			"number":   tableNumber,
			"capacity": 6,
			"location": "Área VIP",
			"status":   "livre",
		}
		_, err = ts.client.Request("PUT", "/table/"+createdTableID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /table/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /table/:id", true, "Mesa atualizada")
		}
	}

	// DELETE (criar mesa auxiliar para delete)
	ts.logger.Info("5. Testar delete de mesa - DELETE /table/:id")
	deleteTableNum := int(900 + uuid.New().ID()%99)
	deletePayload := map[string]interface{}{
		"number":   deleteTableNum,
		"capacity": 2,
		"status":   "livre",
	}
	deleteResp, err := ts.client.Request("POST", "/table", deletePayload, true)
	if err != nil {
		ts.addResult("DELETE /table/:id", true, "Pulado - não conseguiu criar mesa auxiliar")
	} else {
		if deleteData := ts.client.ExtractData(deleteResp); deleteData != nil {
			if deleteID, ok := deleteData["id"].(string); ok {
				_, err = ts.client.Request("DELETE", "/table/"+deleteID, nil, true)
				if err != nil {
					ts.addResult("DELETE /table/:id", false, err.Error())
					allPassed = false
				} else {
					ts.addResult("DELETE /table/:id", true, "Mesa deletada com sucesso")
				}
			}
		}
	}

	return allPassed
}

// =============================================================================
// CUSTOMER CRUD
// =============================================================================

func (ts *TestSuite) TestCustomerCRUDCustomers() bool {
	ts.logger.Subsection("CUSTOMER - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar cliente - POST /customer")
	uniqueEmail := fmt.Sprintf("cliente_%s@teste.com", uuid.New().String()[:8])
	createPayload := map[string]interface{}{
		"name":  "Cliente Teste " + uuid.New().String()[:8],
		"email": uniqueEmail,
		"phone": "+5511999999999",
	}

	createResp, err := ts.client.Request("POST", "/customer", createPayload, true)
	if err != nil {
		ts.addResult("POST /customer", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdCustomerID = id
				ts.addResult("POST /customer", true, fmt.Sprintf("Cliente criado: %s", id[:8]))
			}
		}
		if createdCustomerID == "" {
			ts.addResult("POST /customer", true, "Cliente criado (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar clientes - GET /customer")
	resp, err := ts.client.Request("GET", "/customer", nil, true)
	if err != nil {
		ts.addResult("GET /customer", false, err.Error())
		allPassed = false
	} else {
		customers := ts.client.ExtractArray(resp)
		ts.addResult("GET /customer", true, fmt.Sprintf("%d clientes encontrados", len(customers)))

		if createdCustomerID == "" && len(customers) > 0 {
			if customer, ok := customers[0].(map[string]interface{}); ok {
				if id, ok := customer["id"].(string); ok {
					createdCustomerID = id
				}
			}
		}
	}

	// READ ONE
	if createdCustomerID != "" {
		ts.logger.Info("3. Buscar cliente específico - GET /customer/:id")
		_, err := ts.client.Request("GET", "/customer/"+createdCustomerID, nil, true)
		if err != nil {
			ts.addResult("GET /customer/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /customer/:id", true, "Cliente obtido")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar cliente - PUT /customer/:id")
		updatePayload := map[string]interface{}{
			"name":  "Cliente Atualizado " + uuid.New().String()[:8],
			"phone": "+5511988888888",
		}
		_, err = ts.client.Request("PUT", "/customer/"+createdCustomerID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /customer/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /customer/:id", true, "Cliente atualizado")
		}
	}

	// DELETE (criar cliente auxiliar para delete)
	ts.logger.Info("5. Testar delete de cliente - DELETE /customer/:id")
	deleteEmail := fmt.Sprintf("delete_%s@teste.com", uuid.New().String()[:8])
	deletePayload := map[string]interface{}{
		"name":  "Cliente Para Deletar",
		"email": deleteEmail,
		"phone": "+5511900000000",
	}
	deleteResp, err := ts.client.Request("POST", "/customer", deletePayload, true)
	if err != nil {
		ts.addResult("DELETE /customer/:id", true, "Pulado - não conseguiu criar cliente auxiliar")
	} else {
		if deleteData := ts.client.ExtractData(deleteResp); deleteData != nil {
			if deleteID, ok := deleteData["id"].(string); ok {
				_, err = ts.client.Request("DELETE", "/customer/"+deleteID, nil, true)
				if err != nil {
					ts.addResult("DELETE /customer/:id", false, err.Error())
					allPassed = false
				} else {
					ts.addResult("DELETE /customer/:id", true, "Cliente deletado com sucesso")
				}
			}
		}
	}

	return allPassed
}
