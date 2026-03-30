package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// DumpService gerencia o processo de extração de dados do backend de origem
type DumpService struct {
	client *MigrationClient
	logger *Logger
}

// NewDumpService cria um novo DumpService
func NewDumpService(client *MigrationClient, logger *Logger) *DumpService {
	return &DumpService{client: client, logger: logger}
}

// Run executa o dump completo: faz login, extrai todos os dados e salva no arquivo de saída
func (s *DumpService) Run(email, password, sourceURL, outputFile string) error {
	s.logger.Section("INICIANDO DUMP")
	s.logger.Info("Source: %s", sourceURL)

	// Login para obter todas as combinações org/projeto do usuário
	s.logger.Info("Fazendo login em %s...", sourceURL)
	loginProjects, err := s.client.Login(email, password)
	if err != nil {
		return fmt.Errorf("falha no login: %w", err)
	}
	if len(loginProjects) == 0 {
		return fmt.Errorf("nenhuma organização/projeto encontrado para o usuário %s", email)
	}
	s.logger.Success("Autenticado — %d combinações org/projeto encontradas", len(loginProjects))

	// Agrupar projetos por organização
	orgOrder := []string{}
	orgMap := make(map[string]*OrganizationDump)
	for _, lp := range loginProjects {
		if _, ok := orgMap[lp.OrganizationID]; !ok {
			orgOrder = append(orgOrder, lp.OrganizationID)
			orgMap[lp.OrganizationID] = &OrganizationDump{
				ID:    lp.OrganizationID,
				Name:  lp.OrganizationName,
				Email: lp.OrganizationEmail,
			}
		}
		orgMap[lp.OrganizationID].Projects = append(orgMap[lp.OrganizationID].Projects, ProjectDump{
			ID:   lp.ProjectID,
			Name: lp.ProjectName,
		})
	}

	dump := MigrationDump{
		Metadata: DumpMetadata{
			DumpedAt:    time.Now().Format(time.RFC3339),
			SourceURL:   sourceURL,
			ToolVersion: "1.0.0",
		},
	}

	// Dump de cada organização em ordem
	for orgIdx, orgID := range orgOrder {
		org := orgMap[orgID]
		s.logger.Subsection(fmt.Sprintf("Organização %d/%d: %s", orgIdx+1, len(orgOrder), org.Name))

		var dumpedProjects []ProjectDump
		for pi, proj := range org.Projects {
			s.logger.Info("  Projeto %d/%d: %s (%s)", pi+1, len(org.Projects), proj.Name, proj.ID)
			s.client.SetOrgProject(org.ID, proj.ID)

			data, err := s.dumpProjectData()
			if err != nil {
				s.logger.Error("  Erro ao fazer dump do projeto %s: %v", proj.ID, err)
				data = ProjectData{}
			}

			dumpedProjects = append(dumpedProjects, ProjectDump{
				ID:   proj.ID,
				Name: proj.Name,
				Data: data,
			})
		}

		dump.Organizations = append(dump.Organizations, OrganizationDump{
			ID:       org.ID,
			Name:     org.Name,
			Email:    org.Email,
			Projects: dumpedProjects,
		})
	}

	s.client.ClearOrgProject()

	// Salvar arquivo JSON
	s.logger.Section("SALVANDO DUMP")
	s.logger.Info("Serializando dados...")

	fileData, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar dump: %w", err)
	}

	if err := os.WriteFile(outputFile, fileData, 0644); err != nil {
		return fmt.Errorf("erro ao salvar arquivo '%s': %w", outputFile, err)
	}

	totalProjects := 0
	for _, org := range dump.Organizations {
		totalProjects += len(org.Projects)
	}

	s.logger.Success("Dump concluído!")
	s.logger.Info("  Organizações: %d", len(dump.Organizations))
	s.logger.Info("  Projetos: %d", totalProjects)
	s.logger.Info("  Arquivo: %s (%.1f KB)", outputFile, float64(len(fileData))/1024)

	return nil
}

// dumpProjectData extrai todos os dados do projeto cujo contexto já foi definido no client
func (s *DumpService) dumpProjectData() (ProjectData, error) {
	data := ProjectData{}

	data.Menus = s.getList("/menu")
	s.logger.Info("    menus: %d", len(data.Menus))

	data.Categories = s.getList("/category")
	s.logger.Info("    categorias: %d", len(data.Categories))

	data.Subcategories = s.getList("/subcategory")
	s.logger.Info("    subcategorias: %d", len(data.Subcategories))

	data.Tags = s.getList("/tag")
	s.logger.Info("    tags: %d", len(data.Tags))

	data.Products = s.getList("/product?includeTags=true")
	s.logger.Info("    produtos: %d", len(data.Products))

	data.Environments = s.getList("/environment")
	s.logger.Info("    ambientes: %d", len(data.Environments))

	data.Tables = s.getList("/table")
	s.logger.Info("    mesas: %d", len(data.Tables))

	data.Customers = s.getList("/customer")
	s.logger.Info("    clientes: %d", len(data.Customers))

	data.Reservations = s.getList("/reservation")
	s.logger.Info("    reservas: %d", len(data.Reservations))

	data.Orders = s.getList("/order")
	s.logger.Info("    pedidos: %d", len(data.Orders))

	data.Waitlist = s.getList("/waitlist")
	s.logger.Info("    waitlist: %d", len(data.Waitlist))

	data.Campaigns = s.getList("/campaign")
	s.logger.Info("    campanhas: %d", len(data.Campaigns))

	data.Roles = s.getList("/role")
	s.logger.Info("    roles: %d", len(data.Roles))

	data.StaffAvailability = s.getList("/staff/availability")
	data.StaffSchedules = s.getList("/staff/schedule")
	data.StaffAttendance = s.getList("/staff/attendance")
	data.StaffCommissions = s.getList("/staff/commission")
	data.StaffStock = s.getList("/staff/stock")
	s.logger.Info("    staff (av/sc/at/co/st): %d/%d/%d/%d/%d",
		len(data.StaffAvailability), len(data.StaffSchedules),
		len(data.StaffAttendance), len(data.StaffCommissions), len(data.StaffStock))

	// Buscar relações N:N subcategory → categories
	var subcatCategories []map[string]interface{}
	for _, sub := range data.Subcategories {
		subID, _ := sub["id"].(string)
		if subID == "" {
			continue
		}
		cats := s.getList(fmt.Sprintf("/subcategory/%s/categories", subID))
		if len(cats) > 0 {
			var catIDs []interface{}
			for _, cat := range cats {
				if catID, _ := cat["id"].(string); catID != "" {
					catIDs = append(catIDs, catID)
				}
			}
			subcatCategories = append(subcatCategories, map[string]interface{}{
				"subcategory_id": subID,
				"category_ids":   catIDs,
			})
		}
	}
	data.SubcategoryCategories = subcatCategories
	s.logger.Info("    subcategory-category links: %d", len(data.SubcategoryCategories))

	data.Settings = s.getObject("/settings")
	data.DisplaySettings = s.getObject("/project/settings/display")
	data.Theme = s.getObject("/project/settings/theme")
	s.logger.Info("    settings/display/theme: obtidos")

	return data, nil
}

// getList faz GET em um endpoint e retorna uma lista de entidades.
// Trata múltiplos padrões de resposta da API LEP.
func (s *DumpService) getList(path string) []map[string]interface{} {
	resp, err := s.client.Get(path)
	if err != nil {
		s.logger.Debug("GET %s: %v", path, err)
		return nil
	}
	if resp == nil {
		return nil
	}

	// Padrão 1: {"data": [...]}
	if data, ok := resp["data"]; ok {
		if arr, ok := data.([]interface{}); ok {
			return toMapSlice(arr)
		}
		if dataMap, ok := data.(map[string]interface{}); ok {
			for _, v := range dataMap {
				if arr, ok := v.([]interface{}); ok {
					return toMapSlice(arr)
				}
			}
		}
	}

	// Padrão 2: {"items": [...]}
	if items, ok := resp["items"]; ok {
		if arr, ok := items.([]interface{}); ok {
			return toMapSlice(arr)
		}
	}

	// Padrão 3: chave específica com array (ex: {"products": [...]})
	for _, v := range resp {
		if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
			if _, ok := arr[0].(map[string]interface{}); ok {
				return toMapSlice(arr)
			}
		}
	}

	s.logger.Debug("GET %s: resposta não contém lista reconhecível", path)
	return nil
}

// getObject faz GET e retorna um objeto de configuração
func (s *DumpService) getObject(path string) map[string]interface{} {
	resp, err := s.client.Get(path)
	if err != nil {
		s.logger.Debug("GET %s: %v", path, err)
		return nil
	}
	if resp == nil {
		return nil
	}
	if data, ok := resp["data"].(map[string]interface{}); ok {
		return data
	}
	return resp
}

// toMapSlice converte []interface{} para []map[string]interface{}
func toMapSlice(arr []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

// LoadDumpFile carrega e parseia o arquivo de dump do disco
func LoadDumpFile(path string) (*MigrationDump, error) {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de dump '%s': %w", path, err)
	}

	var dump MigrationDump
	if err := json.Unmarshal(fileData, &dump); err != nil {
		return nil, fmt.Errorf("erro ao parsear dump: %w", err)
	}

	return &dump, nil
}
