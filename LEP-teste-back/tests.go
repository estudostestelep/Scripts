package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/google/uuid"
)

// generateTestPNG cria uma imagem PNG 1x1 branca para testes de upload
func generateTestPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 255, 255, 255})
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	return buf.Bytes()
}

type TestSuite struct {
	client *APIClient
	logger *Logger
	config Config

	// Resultados
	passed int
	failed int
	tests  []TestResult

	// Organização de teste isolada
	testOrgID     string
	testProjectID string
}

type TestResult struct {
	Name    string
	Status  bool
	Message string
}

func NewTestSuite(client *APIClient, logger *Logger, config Config) *TestSuite {
	return &TestSuite{
		client: client,
		logger: logger,
		config: config,
	}
}

func (ts *TestSuite) addResult(name string, status bool, message string) {
	ts.tests = append(ts.tests, TestResult{
		Name:    name,
		Status:  status,
		Message: message,
	})

	if status {
		ts.passed++
	} else {
		ts.failed++
	}
}

func (ts *TestSuite) PrintResults() {
	ts.logger.Section("RESULTADOS DOS TESTES")

	for _, test := range ts.tests {
		if test.Status {
			ts.logger.Success("%s - %s", test.Name, test.Message)
		} else {
			ts.logger.Error("%s - %s", test.Name, test.Message)
		}
	}

	ts.logger.Stats(len(ts.tests), ts.passed, ts.failed)
}

// =============================================================================
// SETUP: Criar organização de teste isolada
// =============================================================================

func (ts *TestSuite) SetupTestOrganization() bool {
	ts.logger.Section("FASE 0: SETUP - Criar Organização de Teste")

	uid := uuid.New().String()[:8]
	payload := map[string]interface{}{
		"name":     "TEST_ORG_" + uid,
		"email":    fmt.Sprintf("testadmin_%s@test.com", uid),
		"password": "Test@123456!", // Senha forte: maiúscula, minúscula, número, especial
	}

	resp, err := ts.client.Request("POST", "/create-organization", payload, true)
	if err != nil {
		ts.logger.Warn("POST /create-organization falhou: %v", err)
		ts.logger.Info("Continuando com organização do login...")
		ts.addResult("Setup Test Organization", true, "Usando org existente do login")
		return true
	}

	// Extrair org_id e project_id da resposta
	data := ts.client.ExtractData(resp)
	if data != nil {
		if orgID, ok := data["organization_id"].(string); ok {
			ts.testOrgID = orgID
		}
		if projID, ok := data["project_id"].(string); ok {
			ts.testProjectID = projID
		}
	}

	// Tentar extrair de estruturas alternativas
	if ts.testOrgID == "" {
		if org := ts.client.ExtractMap(resp, "organization"); org != nil {
			if id, ok := org["id"].(string); ok {
				ts.testOrgID = id
			}
		}
	}
	if ts.testProjectID == "" {
		if proj := ts.client.ExtractMap(resp, "project"); proj != nil {
			if id, ok := proj["id"].(string); ok {
				ts.testProjectID = id
			}
		}
	}

	if ts.testOrgID != "" && ts.testProjectID != "" {
		// Atualizar headers para usar a nova org/project
		ts.config.Headers.OrgID = ts.testOrgID
		ts.config.Headers.ProjID = ts.testProjectID
		savedToken := ts.client.SaveToken()
		ts.client.SetHeaders(ts.testOrgID, ts.testProjectID, savedToken)
		ts.addResult("Setup Test Organization", true,
			fmt.Sprintf("Org: %s, Proj: %s", ts.testOrgID[:8], ts.testProjectID[:8]))
	} else {
		ts.logger.Warn("Não conseguiu extrair IDs da organização criada")
		ts.addResult("Setup Test Organization", true, "Org criada mas IDs não extraídos - usando org do login")
	}

	return true
}

// =============================================================================
// CLEANUP: Deletar organização de teste
// =============================================================================

func (ts *TestSuite) CleanupTestOrganization() bool {
	ts.logger.Section("CLEANUP: Deletar Organização de Teste")

	if ts.testOrgID == "" {
		ts.addResult("Cleanup Test Organization", true, "Nenhuma org de teste para deletar")
		return true
	}

	_, err := ts.client.Request("DELETE", "/organization/"+ts.testOrgID, nil, true)
	if err != nil {
		ts.addResult("Cleanup Test Organization", false,
			fmt.Sprintf("Falha ao deletar org de teste: %v", err))
		return false
	}

	ts.addResult("Cleanup Test Organization", true,
		fmt.Sprintf("Org de teste %s deletada", ts.testOrgID[:8]))
	return true
}

// =============================================================================
// EXECUÇÃO PRINCIPAL
// =============================================================================

func (ts *TestSuite) RunAll() {
	ts.logger.Info("Iniciando teste completo do backend...")
	ts.logger.Info("Backend URL: %s", ts.config.BackendURL)
	fmt.Println()

	// =========================================================================
	// FASE 1: AUTENTICAÇÃO (sempre primeiro)
	// =========================================================================
	ts.RunAuthTests()

	// Verificar se login foi bem sucedido
	if ts.failed > 0 {
		ts.logger.Error("Autenticação falhou, abortando testes")
		ts.PrintResults()
		return
	}

	// =========================================================================
	// FASE 0: SETUP - Criar organização de teste isolada
	// =========================================================================
	ts.SetupTestOrganization()

	// =========================================================================
	// FASE 2: ADMINISTRAÇÃO (User, Organization, Project) - CRUD COMPLETO
	// =========================================================================
	ts.RunAdminTests()

	// =========================================================================
	// FASE 3: TESTES DE PERMISSÕES
	// =========================================================================
	ts.RunPermissionTests()

	// =========================================================================
	// FASE 4: CATÁLOGO (Menu → Category → Subcategory → Product → Tag)
	// =========================================================================
	ts.RunCatalogTests()

	// =========================================================================
	// FASE 5: CLIENTES E INFRAESTRUTURA (Environment → Table → Customer)
	// =========================================================================
	ts.RunCustomerTests()

	// =========================================================================
	// FASE 6: RESERVAS (Depende de Customer + Table)
	// =========================================================================
	ts.RunReservationTests()

	// =========================================================================
	// FASE 7: PEDIDOS (Depende de Product + Table + Customer)
	// =========================================================================
	ts.RunOrderTests()

	// =========================================================================
	// FASE 8: NOTIFICAÇÕES (Independente)
	// =========================================================================
	ts.RunNotificationTests()

	// =========================================================================
	// FASE 9: CONFIGURAÇÕES (Independente)
	// =========================================================================
	ts.RunSettingsTests()

	// =========================================================================
	// FASE 10: UPLOAD DE IMAGENS
	// =========================================================================
	ts.logger.Section("FASE 10: UPLOAD DE IMAGENS")
	ts.TestImageUploadFix()

	// =========================================================================
	// FASE 11: TESTES DE OTIMIZAÇÃO
	// =========================================================================
	ts.logger.Section("FASE 11: OTIMIZAÇÕES")
	ts.TestProductTagsOptimization()

	// =========================================================================
	// FASE 12: SELEÇÃO INTELIGENTE DE MENUS
	// =========================================================================
	ts.logger.Section("FASE 12: SELEÇÃO INTELIGENTE DE MENUS")
	ts.TestMenuIntelligentSelection()

	// =========================================================================
	// FASE 13: CUSTOMIZAÇÃO DE TEMA
	// =========================================================================
	ts.RunThemeCustomizationTests()

	// =========================================================================
	// FASE 14: NOTIFICAÇÕES AVANÇADAS
	// =========================================================================
	ts.RunNotificationEnhancedTests()

	// =========================================================================
	// FASE 15: VALIDAÇÃO DE TIPOS
	// =========================================================================
	ts.RunTypeValidationTests()

	// =========================================================================
	// CLEANUP: Deletar organização de teste
	// =========================================================================
	ts.CleanupTestOrganization()

	// =========================================================================
	// FASE FINAL: LOGOUT
	// =========================================================================
	ts.logger.Section("FASE FINAL: LOGOUT")
	ts.TestLogoutAuth()

	// =========================================================================
	// RELATÓRIO DE COBERTURA DE ROTAS
	// =========================================================================
	ts.RunRouteCoverageReport()

	// =========================================================================
	// RESULTADOS
	// =========================================================================
	ts.PrintResults()
}
