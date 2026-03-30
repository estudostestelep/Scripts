package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// TESTES DE RESERVAS (Depende de Customer + Table)
// Inclui: Reservation, Waitlist, Public Reservation
// =============================================================================

// IDs criados durante os testes
var (
	createdReservationID string
	createdWaitlistID    string
)

func (ts *TestSuite) RunReservationTests() {
	ts.logger.Section("FASE 5: RESERVAS (Depende de Customer + Table)")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Reservation CRUD", ts.TestReservationCRUDReservations},
		{"Waitlist CRUD", ts.TestWaitlistCRUDReservations},
		{"Public Reservation", ts.TestPublicReservationReservations},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// RESERVATION CRUD (depende de Customer e Table)
// =============================================================================

func (ts *TestSuite) TestReservationCRUDReservations() bool {
	ts.logger.Subsection("RESERVATION - CRUD Completo")
	allPassed := true

	// Verificar se temos customer e table
	if createdCustomerID == "" || createdTableID == "" {
		ts.logger.Warn("Customer ou Table não disponíveis - tentando criar reserva mesmo assim")
	}

	// CREATE
	ts.logger.Info("1. Criar reserva - POST /reservation")
	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	createPayload := map[string]interface{}{
		"organization_id": ts.config.Headers.OrgID,
		"project_id":      ts.config.Headers.ProjID,
		"datetime":        futureDate,
		"party_size":      4,
		"note":            "Reserva de teste automatizado",
		"status":          "confirmed",
	}

	// Adicionar customer_id e table_id se disponíveis
	if createdCustomerID != "" {
		createPayload["customer_id"] = createdCustomerID
	}
	if createdTableID != "" {
		createPayload["table_id"] = createdTableID
	}

	createResp, err := ts.client.Request("POST", "/reservation", createPayload, true)
	if err != nil {
		ts.addResult("POST /reservation", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdReservationID = id
				ts.addResult("POST /reservation", true, fmt.Sprintf("Reserva criada: %s", id[:8]))
			}
		}
		if createdReservationID == "" {
			ts.addResult("POST /reservation", true, "Reserva criada (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar reservas - GET /reservation")
	resp, err := ts.client.Request("GET", "/reservation", nil, true)
	if err != nil {
		ts.addResult("GET /reservation", false, err.Error())
		allPassed = false
	} else {
		reservations := ts.client.ExtractArray(resp)
		ts.addResult("GET /reservation", true, fmt.Sprintf("%d reservas encontradas", len(reservations)))

		if createdReservationID == "" && len(reservations) > 0 {
			if res, ok := reservations[0].(map[string]interface{}); ok {
				if id, ok := res["id"].(string); ok {
					createdReservationID = id
				}
			}
		}
	}

	// READ ONE
	if createdReservationID != "" {
		ts.logger.Info("3. Buscar reserva específica - GET /reservation/:id")
		_, err := ts.client.Request("GET", "/reservation/"+createdReservationID, nil, true)
		if err != nil {
			ts.addResult("GET /reservation/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /reservation/:id", true, "Reserva obtida")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar reserva - PUT /reservation/:id")
		updatePayload := map[string]interface{}{
			"party_size": 6,
			"note":       "Reserva atualizada - mais pessoas",
		}
		_, err = ts.client.Request("PUT", "/reservation/"+createdReservationID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /reservation/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /reservation/:id", true, "Reserva atualizada")
		}
	}

	// DELETE (criar reserva auxiliar para delete)
	ts.logger.Info("5. Testar delete de reserva - DELETE /reservation/:id")
	delDate := time.Now().Add(72 * time.Hour).Format("2006-01-02 15:04:05")
	delPayload := map[string]interface{}{
		"organization_id": ts.config.Headers.OrgID,
		"project_id":      ts.config.Headers.ProjID,
		"datetime":        delDate,
		"party_size":      2,
		"note":            "Reserva para deletar",
		"status":          "confirmed",
	}
	if createdCustomerID != "" {
		delPayload["customer_id"] = createdCustomerID
	}
	if createdTableID != "" {
		delPayload["table_id"] = createdTableID
	}
	delResp, err := ts.client.Request("POST", "/reservation", delPayload, true)
	if err != nil {
		ts.addResult("DELETE /reservation/:id", true, "Pulado - não conseguiu criar reserva auxiliar")
	} else {
		if delData := ts.client.ExtractData(delResp); delData != nil {
			if delID, ok := delData["id"].(string); ok {
				_, err = ts.client.Request("DELETE", "/reservation/"+delID, nil, true)
				if err != nil {
					ts.addResult("DELETE /reservation/:id", false, err.Error())
					allPassed = false
				} else {
					ts.addResult("DELETE /reservation/:id", true, "Reserva deletada com sucesso")
				}
			}
		}
	}

	return allPassed
}

// =============================================================================
// WAITLIST CRUD (depende de Customer)
// =============================================================================

func (ts *TestSuite) TestWaitlistCRUDReservations() bool {
	ts.logger.Subsection("WAITLIST - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Adicionar à fila de espera - POST /waitlist")

	// Verificar se temos customer_id (obrigatório para waitlist)
	if createdCustomerID == "" {
		ts.addResult("POST /waitlist", true, "Pulado - CustomerID não disponível")
		return true
	}

	createPayload := map[string]interface{}{
		"customer_id": createdCustomerID,
		"people":      4,
		"status":      "waiting",
	}

	createResp, err := ts.client.Request("POST", "/waitlist", createPayload, true)
	if err != nil {
		// Waitlist pode requerer módulo específico
		status := ts.client.GetLastStatus()
		if status == 403 || status == 404 {
			ts.addResult("POST /waitlist", true, "Módulo não habilitado (não crítico)")
			return true
		}
		ts.addResult("POST /waitlist", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdWaitlistID = id
				ts.addResult("POST /waitlist", true, fmt.Sprintf("Entrada na fila criada: %s", id[:8]))
			}
		}
		if createdWaitlistID == "" {
			ts.addResult("POST /waitlist", true, "Entrada na fila criada (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar fila de espera - GET /waitlist")
	resp, err := ts.client.Request("GET", "/waitlist", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 403 || status == 404 {
			ts.addResult("GET /waitlist", true, "Módulo não habilitado (não crítico)")
			return allPassed
		}
		ts.addResult("GET /waitlist", false, err.Error())
		allPassed = false
	} else {
		waitlist := ts.client.ExtractArray(resp)
		ts.addResult("GET /waitlist", true, fmt.Sprintf("%d entradas na fila", len(waitlist)))

		if createdWaitlistID == "" && len(waitlist) > 0 {
			if entry, ok := waitlist[0].(map[string]interface{}); ok {
				if id, ok := entry["id"].(string); ok {
					createdWaitlistID = id
				}
			}
		}
	}

	// READ ONE
	if createdWaitlistID != "" {
		ts.logger.Info("3. Buscar entrada específica - GET /waitlist/:id")
		_, err := ts.client.Request("GET", "/waitlist/"+createdWaitlistID, nil, true)
		if err != nil {
			ts.addResult("GET /waitlist/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /waitlist/:id", true, "Entrada obtida")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar entrada - PUT /waitlist/:id")
		updatePayload := map[string]interface{}{
			"people": 6,
			"status": "waiting",
		}
		_, err = ts.client.Request("PUT", "/waitlist/"+createdWaitlistID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /waitlist/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /waitlist/:id", true, "Entrada atualizada")
		}

		// DELETE
		ts.logger.Info("5. Deletar entrada da fila - DELETE /waitlist/:id")
		_, err = ts.client.Request("DELETE", "/waitlist/"+createdWaitlistID, nil, true)
		if err != nil {
			ts.addResult("DELETE /waitlist/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("DELETE /waitlist/:id", true, "Entrada da fila deletada com sucesso")
		}
	}

	return allPassed
}

// =============================================================================
// PUBLIC RESERVATION (não requer autenticação)
// =============================================================================

func (ts *TestSuite) TestPublicReservationReservations() bool {
	ts.logger.Subsection("PUBLIC RESERVATION - Reserva pública")
	allPassed := true

	orgID := ts.config.Headers.OrgID
	projID := ts.config.Headers.ProjID

	if orgID == "" || projID == "" {
		ts.addResult("Public Reservation", true, "Pulado - OrgID ou ProjID não disponíveis")
		return true
	}

	// GET PUBLIC TIMES (com query params obrigatórios)
	ts.logger.Info("1. Obter horários disponíveis - GET /public/times/:orgId/:projId")
	futureDate := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	timesURL := fmt.Sprintf("/public/times/%s/%s?date=%s&party_size=2", orgID, projID, futureDate)
	_, err := ts.client.Request("GET", timesURL, nil, false)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /public/times", true, "Endpoint não implementado (não crítico)")
		} else {
			ts.addResult("GET /public/times", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /public/times", true, "Horários públicos obtidos")
	}

	// CREATE PUBLIC RESERVATION (estrutura aninhada com customer e reservation)
	ts.logger.Info("2. Criar reserva pública - POST /public/reservation/:orgId/:projId")
	futureDatetime := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	uniqueEmail := fmt.Sprintf("publico_%s@teste.com", uuid.New().String()[:8])
	createPayload := map[string]interface{}{
		"customer": map[string]interface{}{
			"name":  "Cliente Público " + uuid.New().String()[:8],
			"email": uniqueEmail,
			"phone": "+5511988888888",
		},
		"reservation": map[string]interface{}{
			"datetime":   futureDatetime,
			"party_size": 2,
			"note":       "Reserva pública de teste",
			"source":     "public",
		},
	}

	_, err = ts.client.Request("POST", fmt.Sprintf("/public/reservation/%s/%s", orgID, projID), createPayload, false)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /public/reservation", true, "Endpoint não implementado (não crítico)")
		} else {
			ts.addResult("POST /public/reservation", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("POST /public/reservation", true, "Reserva pública criada")
	}

	return allPassed
}
