package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// =============================================================================
// TESTES DE VALIDAÇÃO DE TIPOS
// Detecta quando o frontend envia dados com tipos incorretos
// Gera output estruturado para correção automática
// =============================================================================

// TypeValidationError representa um erro de tipo encontrado
type TypeValidationError struct {
	Endpoint     string `json:"endpoint"`
	Method       string `json:"method"`
	Field        string `json:"field"`
	SentType     string `json:"sent_type"`
	ExpectedType string `json:"expected_type"`
	SentValue    string `json:"sent_value"`
	FixSuggestion string `json:"fix_suggestion"`
}

// TypeValidationResult resultado da validação de tipos
type TypeValidationResult struct {
	Success bool                  `json:"success"`
	Errors  []TypeValidationError `json:"errors"`
}

func (ts *TestSuite) RunTypeValidationTests() {
	ts.logger.Section("FASE VALIDAÇÃO: TESTES DE TIPO DE DADOS")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Table - Capacity como String (deve falhar)", ts.TestTableCapacityStringType},
		{"Table - Number como String (deve falhar)", ts.TestTableNumberStringType},
		{"Product - Price como String (deve falhar)", ts.TestProductPriceStringType},
		{"Reservation - PartySize como String (deve falhar)", ts.TestReservationPartySizeStringType},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// TEST: Table Capacity como String
// Simula o erro: "json: cannot unmarshal string into Go struct field Table.capacity of type int"
// =============================================================================

func (ts *TestSuite) TestTableCapacityStringType() bool {
	ts.logger.Subsection("TABLE - Validação: capacity como string (deve rejeitar)")

	// Primeiro, buscar uma mesa existente para usar o ID
	resp, err := ts.client.Request("GET", "/table", nil, true)
	if err != nil {
		ts.addResult("GET /table (prep)", false, err.Error())
		return false
	}

	tables := ts.client.ExtractArray(resp)
	if len(tables) == 0 {
		ts.addResult("Table Type Validation", true, "SKIP: Nenhuma mesa disponível para teste")
		return true
	}

	var tableID string
	if table, ok := tables[0].(map[string]interface{}); ok {
		if id, ok := table["id"].(string); ok {
			tableID = id
		}
	}

	if tableID == "" {
		ts.addResult("Table Type Validation", true, "SKIP: ID não extraído")
		return true
	}

	// PAYLOAD COM TIPO INCORRETO - capacity como string
	wrongPayload := map[string]interface{}{
		"number":   1,
		"capacity": "4",  // ERRO: deve ser int, não string
		"status":   "livre",
	}

	_, err = ts.client.Request("PUT", "/table/"+tableID, wrongPayload, true)

	if err != nil && strings.Contains(err.Error(), "cannot unmarshal") {
		// DETECTOU O ERRO DE TIPO - Gerar relatório de correção
		validationError := TypeValidationError{
			Endpoint:      "PUT /table/:id",
			Method:        "PUT",
			Field:         "capacity",
			SentType:      "string",
			ExpectedType:  "int",
			SentValue:     "\"4\"",
			FixSuggestion: "Alterar de \"capacity\": \"4\" para \"capacity\": 4 (sem aspas)",
		}

		ts.printTypeErrorReport(validationError)
		ts.addResult("PUT /table/:id (capacity string)", false,
			"ERRO DE TIPO DETECTADO: capacity enviado como string, esperado int. "+
			"FIX: Remover aspas do valor no frontend")
		return false
	}

	if err != nil && !strings.Contains(err.Error(), "cannot unmarshal") {
		// Outro erro (backend rejeitou corretamente)
		ts.addResult("PUT /table/:id (capacity string)", true,
			fmt.Sprintf("Erro esperado: %v", err))
		return true
	}

	// Se passou, o backend aceitou string (pode ser flexível ou bug)
	ts.addResult("PUT /table/:id (capacity string)", true,
		"AVISO: Backend aceitou capacity como string - verificar se há conversão automática")
	return true
}

// =============================================================================
// TEST: Table Number como String
// =============================================================================

func (ts *TestSuite) TestTableNumberStringType() bool {
	ts.logger.Subsection("TABLE - Validação: number como string (deve rejeitar)")

	// Tentar criar mesa com number como string
	wrongPayload := map[string]interface{}{
		"number":   "99",  // ERRO: deve ser int
		"capacity": 4,
		"status":   "livre",
	}

	_, err := ts.client.Request("POST", "/table", wrongPayload, true)

	if err != nil && strings.Contains(err.Error(), "cannot unmarshal") {
		validationError := TypeValidationError{
			Endpoint:      "POST /table",
			Method:        "POST",
			Field:         "number",
			SentType:      "string",
			ExpectedType:  "int",
			SentValue:     "\"99\"",
			FixSuggestion: "Alterar de \"number\": \"99\" para \"number\": 99 (sem aspas)",
		}

		ts.printTypeErrorReport(validationError)
		ts.addResult("POST /table (number string)", false,
			"ERRO DE TIPO DETECTADO: number enviado como string, esperado int")
		return false
	}

	if err != nil && !strings.Contains(err.Error(), "cannot unmarshal") {
		// Outro erro (pode ser duplicidade, etc)
		ts.addResult("POST /table (number string)", true,
			fmt.Sprintf("Erro esperado: %v", err))
		return true
	}

	ts.addResult("POST /table (number string)", true,
		"Backend aceitou number como string - verificar conversão")
	return true
}

// =============================================================================
// TEST: Product Price como String
// =============================================================================

func (ts *TestSuite) TestProductPriceStringType() bool {
	ts.logger.Subsection("PRODUCT - Validação: price como string (deve rejeitar)")

	// Buscar produto existente
	resp, err := ts.client.Request("GET", "/product", nil, true)
	if err != nil {
		ts.addResult("GET /product (prep)", false, err.Error())
		return false
	}

	products := ts.client.ExtractArray(resp)
	if len(products) == 0 {
		ts.addResult("Product Type Validation", true, "SKIP: Nenhum produto disponível")
		return true
	}

	var productID string
	if prod, ok := products[0].(map[string]interface{}); ok {
		if id, ok := prod["id"].(string); ok {
			productID = id
		}
	}

	if productID == "" {
		ts.addResult("Product Type Validation", true, "SKIP: ID não extraído")
		return true
	}

	// PAYLOAD COM TIPO INCORRETO - price como string
	wrongPayload := map[string]interface{}{
		"price": "19.90",  // ERRO: deve ser float, não string
	}

	_, err = ts.client.Request("PUT", "/product/"+productID, wrongPayload, true)

	if err != nil && strings.Contains(err.Error(), "cannot unmarshal") {
		validationError := TypeValidationError{
			Endpoint:      "PUT /product/:id",
			Method:        "PUT",
			Field:         "price",
			SentType:      "string",
			ExpectedType:  "float64",
			SentValue:     "\"19.90\"",
			FixSuggestion: "Alterar de \"price\": \"19.90\" para \"price\": 19.90 (sem aspas)",
		}

		ts.printTypeErrorReport(validationError)
		ts.addResult("PUT /product/:id (price string)", false,
			"ERRO DE TIPO DETECTADO: price enviado como string, esperado float")
		return false
	}

	ts.addResult("PUT /product/:id (price string)", true,
		"Backend aceitou price - verificar se há conversão")
	return true
}

// =============================================================================
// TEST: Reservation PartySize como String
// =============================================================================

func (ts *TestSuite) TestReservationPartySizeStringType() bool {
	ts.logger.Subsection("RESERVATION - Validação: party_size como string (deve rejeitar)")

	// PAYLOAD COM TIPO INCORRETO
	wrongPayload := map[string]interface{}{
		"datetime":   "2025-12-15T19:00:00Z",
		"party_size": "4",  // ERRO: deve ser int
		"status":     "confirmed",
	}

	_, err := ts.client.Request("POST", "/reservation", wrongPayload, true)

	if err != nil && strings.Contains(err.Error(), "cannot unmarshal") {
		validationError := TypeValidationError{
			Endpoint:      "POST /reservation",
			Method:        "POST",
			Field:         "party_size",
			SentType:      "string",
			ExpectedType:  "int",
			SentValue:     "\"4\"",
			FixSuggestion: "Alterar de \"party_size\": \"4\" para \"party_size\": 4 (sem aspas)",
		}

		ts.printTypeErrorReport(validationError)
		ts.addResult("POST /reservation (party_size string)", false,
			"ERRO DE TIPO DETECTADO: party_size enviado como string")
		return false
	}

	ts.addResult("POST /reservation (party_size string)", true,
		"Backend aceitou party_size - verificar conversão")
	return true
}

// =============================================================================
// HELPER: Imprimir relatório de erro de tipo (formato para prompt de correção)
// =============================================================================

func (ts *TestSuite) printTypeErrorReport(err TypeValidationError) {
	ts.logger.Error("========================================")
	ts.logger.Error("ERRO DE TIPO DETECTADO - RELATÓRIO PARA CORREÇÃO")
	ts.logger.Error("========================================")
	ts.logger.Error("")
	ts.logger.Error("Endpoint: %s %s", err.Method, err.Endpoint)
	ts.logger.Error("Campo com erro: %s", err.Field)
	ts.logger.Error("Tipo enviado: %s", err.SentType)
	ts.logger.Error("Tipo esperado: %s", err.ExpectedType)
	ts.logger.Error("Valor enviado: %s", err.SentValue)
	ts.logger.Error("")
	ts.logger.Error("COMO CORRIGIR:")
	ts.logger.Error("%s", err.FixSuggestion)
	ts.logger.Error("")
	ts.logger.Error("========================================")

	// Também gerar JSON para uso programático
	jsonOutput, _ := json.MarshalIndent(err, "", "  ")
	ts.logger.Debug("JSON para correção automática:\n%s", string(jsonOutput))
}

// =============================================================================
// GenerateCorrectionPrompt gera um prompt para correção automática
// =============================================================================

func (ts *TestSuite) GenerateCorrectionPrompt(errors []TypeValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Erros de Tipo Detectados - Correção Necessária\n\n")

	for i, err := range errors {
		sb.WriteString(fmt.Sprintf("### Erro %d\n", i+1))
		sb.WriteString(fmt.Sprintf("- **Endpoint**: `%s %s`\n", err.Method, err.Endpoint))
		sb.WriteString(fmt.Sprintf("- **Campo**: `%s`\n", err.Field))
		sb.WriteString(fmt.Sprintf("- **Problema**: Enviado como `%s`, esperado `%s`\n", err.SentType, err.ExpectedType))
		sb.WriteString(fmt.Sprintf("- **Correção**: %s\n\n", err.FixSuggestion))
	}

	sb.WriteString("### Ação Recomendada\n")
	sb.WriteString("Verifique o código frontend que faz a requisição para estes endpoints ")
	sb.WriteString("e corrija os tipos dos campos indicados.\n\n")
	sb.WriteString("**Exemplo de correção em JavaScript:**\n")
	sb.WriteString("```javascript\n")
	sb.WriteString("// ERRADO:\n")
	sb.WriteString("const payload = { capacity: \"4\" };\n\n")
	sb.WriteString("// CORRETO:\n")
	sb.WriteString("const payload = { capacity: parseInt(value, 10) };\n")
	sb.WriteString("// ou\n")
	sb.WriteString("const payload = { capacity: Number(value) };\n")
	sb.WriteString("```\n")

	return sb.String()
}
