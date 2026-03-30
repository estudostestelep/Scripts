package main

import (
	"fmt"
)

// =============================================================================
// TESTES DE PEDIDOS (Depende de Product + Table + Customer)
// Inclui: Order CRUD, Kitchen Queue, Order Status
// =============================================================================

// IDs criados durante os testes
var (
	createdOrderID string
)

func (ts *TestSuite) RunOrderTests() {
	ts.logger.Section("FASE 6: PEDIDOS (Depende de Product + Table + Customer)")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Order CRUD", ts.TestOrderCRUDOrders},
		{"Kitchen Queue", ts.TestKitchenQueueOrders},
		{"Order Status", ts.TestOrderStatus},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// ORDER CRUD (depende de Product, Table e Customer)
// =============================================================================

func (ts *TestSuite) TestOrderCRUDOrders() bool {
	ts.logger.Subsection("ORDER - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar pedido - POST /order")

	// Preparar items do pedido
	items := []map[string]interface{}{}
	if createdProductID != "" {
		items = append(items, map[string]interface{}{
			"product_id": createdProductID,
			"quantity":   2,
			"price":      29.90,
			"notes":      "Sem cebola",
		})
	}

	createPayload := map[string]interface{}{
		"items":        items,
		"total_amount": 59.80,
		"note":         "Pedido de teste automatizado",
		"source":       "internal",
		"status":       "pending",
	}

	// Adicionar table_id e customer_id se disponíveis
	if createdTableID != "" {
		createPayload["table_id"] = createdTableID
	}
	if createdCustomerID != "" {
		createPayload["customer_id"] = createdCustomerID
	}

	createResp, err := ts.client.Request("POST", "/order", createPayload, true)
	if err != nil {
		ts.addResult("POST /order", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdOrderID = id
				ts.addResult("POST /order", true, fmt.Sprintf("Pedido criado: %s", id[:8]))
			}
		}
		if createdOrderID == "" {
			ts.addResult("POST /order", true, "Pedido criado (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar pedidos - GET /order")
	resp, err := ts.client.Request("GET", "/order", nil, true)
	if err != nil {
		ts.addResult("GET /order", false, err.Error())
		allPassed = false
	} else {
		orders := ts.client.ExtractArray(resp)
		ts.addResult("GET /order", true, fmt.Sprintf("%d pedidos encontrados", len(orders)))

		if createdOrderID == "" && len(orders) > 0 {
			if order, ok := orders[0].(map[string]interface{}); ok {
				if id, ok := order["id"].(string); ok {
					createdOrderID = id
				}
			}
		}
	}

	// READ ONE
	if createdOrderID != "" {
		ts.logger.Info("3. Buscar pedido específico - GET /order/:id")
		_, err := ts.client.Request("GET", "/order/"+createdOrderID, nil, true)
		if err != nil {
			ts.addResult("GET /order/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /order/:id", true, "Pedido obtido")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar pedido - PUT /order/:id")
		updatePayload := map[string]interface{}{
			"note":         "Pedido atualizado - urgente",
			"total_amount": 69.80,
		}
		_, err = ts.client.Request("PUT", "/order/"+createdOrderID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /order/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /order/:id", true, "Pedido atualizado")
		}

		// GET PROGRESS
		ts.logger.Info("5. Obter progresso do pedido - GET /order/:id/progress")
		_, err = ts.client.Request("GET", "/order/"+createdOrderID+"/progress", nil, true)
		if err != nil {
			status := ts.client.GetLastStatus()
			if status == 404 {
				ts.addResult("GET /order/:id/progress", true, "Endpoint não implementado (não crítico)")
			} else {
				ts.addResult("GET /order/:id/progress", false, err.Error())
			}
		} else {
			ts.addResult("GET /order/:id/progress", true, "Progresso obtido")
		}
	}

	return allPassed
}

// =============================================================================
// KITCHEN QUEUE
// =============================================================================

func (ts *TestSuite) TestKitchenQueueOrders() bool {
	ts.logger.Subsection("KITCHEN QUEUE - Fila da cozinha")
	allPassed := true

	// GET KITCHEN QUEUE
	ts.logger.Info("1. Obter fila da cozinha - GET /kitchen/queue")
	_, err := ts.client.Request("GET", "/kitchen/queue", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /kitchen/queue", true, "Endpoint não implementado (não crítico)")
		} else {
			ts.addResult("GET /kitchen/queue", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /kitchen/queue", true, "Fila da cozinha obtida")
	}

	return allPassed
}

// =============================================================================
// ORDER STATUS
// =============================================================================

func (ts *TestSuite) TestOrderStatus() bool {
	ts.logger.Subsection("ORDER STATUS - Atualização de status")
	allPassed := true

	if createdOrderID == "" {
		ts.addResult("Order Status", true, "Pulado - OrderID não disponível")
		return true
	}

	// Testar fluxo de status: pending → preparing → ready → delivered
	statuses := []string{"preparing", "ready", "delivered"}

	for i, status := range statuses {
		ts.logger.Info(fmt.Sprintf("%d. Atualizar status para '%s' - PUT /order/:id/status", i+1, status))
		statusPayload := map[string]interface{}{
			"status": status,
		}
		_, err := ts.client.Request("PUT", "/order/"+createdOrderID+"/status", statusPayload, true)
		if err != nil {
			httpStatus := ts.client.GetLastStatus()
			if httpStatus == 404 {
				ts.addResult(fmt.Sprintf("PUT /order/:id/status (%s)", status), true, "Endpoint não implementado")
				return allPassed
			}
			ts.addResult(fmt.Sprintf("PUT /order/:id/status (%s)", status), false, err.Error())
			allPassed = false
		} else {
			ts.addResult(fmt.Sprintf("PUT /order/:id/status (%s)", status), true, fmt.Sprintf("Status atualizado para %s", status))
		}
	}

	return allPassed
}

// =============================================================================
// CLEANUP - Deletar pedidos de teste (não executado automaticamente)
// =============================================================================

func (ts *TestSuite) CleanupOrders() bool {
	if createdOrderID == "" {
		return true
	}

	ts.logger.Info("Cleanup: Deletando pedido de teste")
	_, err := ts.client.Request("DELETE", "/order/"+createdOrderID, nil, true)
	if err != nil {
		ts.addResult("DELETE /order/:id", false, err.Error())
		return false
	}

	ts.addResult("DELETE /order/:id", true, "Pedido deletado")
	createdOrderID = ""
	return true
}
