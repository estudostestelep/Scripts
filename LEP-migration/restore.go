package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// RestoreService gerencia o processo de recriação dos dados no backend de destino
type RestoreService struct {
	client        *MigrationClient
	logger        *Logger
	sourceURL     string // URL base da origem, para re-download de imagens
	timeout       int    // timeout HTTP em segundos
	adminEmail    string // email do admin para fallback de login
	adminPassword string // senha do admin para fallback de login
	stats         RestoreStats
}

// NewRestoreService cria um novo RestoreService
func NewRestoreService(client *MigrationClient, logger *Logger, sourceURL string, timeout int, adminEmail, adminPassword string) *RestoreService {
	return &RestoreService{
		client:        client,
		logger:        logger,
		sourceURL:     sourceURL,
		timeout:       timeout,
		adminEmail:    adminEmail,
		adminPassword: adminPassword,
	}
}

// Run executa o restore completo a partir de um dump.
// orgPassword é a senha usada para criar as organizações no destino.
func (s *RestoreService) Run(dump *MigrationDump, orgPassword string) error {
	s.logger.Section("INICIANDO RESTORE")
	s.logger.Info("Destino: %s", s.client.baseURL)
	s.logger.Info("Organizações no dump: %d", len(dump.Organizations))

	s.logger.Warn("ATENÇÃO: imagens (image_url) apontam para o storage de origem e podem não funcionar no destino.")
	s.logger.Warn("ATENÇÃO: senhas de usuários/clientes não são migradas. Senha padrão das orgs: [configurada].")

	for orgIdx, org := range dump.Organizations {
		s.logger.Subsection(fmt.Sprintf("Organização %d/%d: %s", orgIdx+1, len(dump.Organizations), org.Name))

		// Limpar auth antes de cada org: /create-organization é endpoint público
		s.client.ClearAuth()

		if err := s.restoreOrganization(org, orgPassword); err != nil {
			s.logger.Error("Erro na organização '%s': %v", org.Name, err)
			s.stats.Failed++
		}
	}

	s.logger.Summary("RESTORE CONCLUÍDO", s.stats.Created, s.stats.Skipped, s.stats.Failed)
	return nil
}

// doAdminLogin tenta login admin, com fallback para login normal
func (s *RestoreService) doAdminLogin() error {
	// Tentar /admin/login primeiro
	if err := s.client.AdminLogin(s.adminEmail, s.adminPassword); err != nil {
		s.logger.Warn("Admin login falhou, tentando login normal: %v", err)
		if _, loginErr := s.client.Login(s.adminEmail, s.adminPassword); loginErr != nil {
			return fmt.Errorf("admin login: %v, login normal: %v", err, loginErr)
		}
	}
	s.logger.Success("Login admin realizado com sucesso")
	return nil
}

// orgEmailFallback retorna o email da org ou gera um a partir do nome quando vazio
func orgEmailFallback(name, email string) string {
	if email != "" {
		return email
	}
	slug := strings.ToLower(name)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug + "@migration.local"
}

// restoreOrganization cria a organização no destino e restaura seus projetos
func (s *RestoreService) restoreOrganization(org OrganizationDump, orgPassword string) error {
	if orgPassword == "" {
		orgPassword = "senha123"
	}

	email := orgEmailFallback(org.Name, org.Email)
	s.logger.Info("Criando organização '%s' com email '%s'...", org.Name, email)
	orgID, projID, err := s.client.CreateOrganization(org.Name, email, orgPassword)
	if err != nil {
		// CreateOrganization falhou (org já existe e login com email gerado falhou)
		// Tentar admin login e buscar a org pelo nome
		s.logger.Warn("CreateOrganization falhou: %v", err)
		if s.adminEmail == "" {
			return fmt.Errorf("erro ao criar organização e sem credenciais admin para fallback: %w", err)
		}
		s.logger.Info("Fazendo admin login para encontrar organização existente...")
		if loginErr := s.doAdminLogin(); loginErr != nil {
			return fmt.Errorf("admin login falhou: %w (erro original: %v)", loginErr, err)
		}
		orgID, projID, err = s.client.FindOrganizationByName(org.Name)
		if err != nil {
			return fmt.Errorf("erro ao encontrar organização pelo nome: %w", err)
		}
		s.logger.Success("Organização encontrada via admin (ID: %s, Projeto: %s)", orgID, projID)
	} else {
		s.logger.Success("Organização criada (ID: %s, Projeto padrão: %s)", orgID, projID)
	}

	// Se o token não foi obtido, fazer login admin como fallback
	if s.client.token == "" && s.adminEmail != "" {
		s.logger.Info("Token não disponível, fazendo login admin como fallback...")
		if loginErr := s.doAdminLogin(); loginErr != nil {
			return fmt.Errorf("fallback admin login falhou: %w", loginErr)
		}
	}

	if len(org.Projects) == 0 {
		s.logger.Warn("Nenhum projeto no dump para esta organização")
		return nil
	}

	// O primeiro projeto do dump mapeia para o projeto padrão criado com a org
	firstProject := org.Projects[0]
	s.client.SetOrgProject(orgID, projID)

	s.logger.Info("Restaurando projeto '%s' (mapeado para projeto padrão %s)...", firstProject.Name, projID)
	projStats, err := s.restoreProject(firstProject.Data)
	if err != nil {
		s.logger.Error("Erro no projeto '%s': %v", firstProject.Name, err)
	}
	s.stats.Add(projStats)

	// Projetos adicionais: avisar que não serão migrados nesta versão
	if len(org.Projects) > 1 {
		s.logger.Warn("Organização '%s' tem %d projetos, mas apenas o primeiro foi migrado (criação de projetos adicionais não suportada nesta versão)",
			org.Name, len(org.Projects))
	}

	s.client.ClearOrgProject()
	return nil
}

// restoreProject restaura todos os dados de um projeto na ordem correta de dependência
func (s *RestoreService) restoreProject(data ProjectData) (RestoreStats, error) {
	stats := RestoreStats{}
	idMap := NewIDMap()

	// 1. Menus (sem dependências)
	s.logger.Subsection("Menus")
	c, sk, f := s.restoreMenus(data.Menus, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 2. Categorias (dependem de menu_id)
	s.logger.Subsection("Categorias")
	c, sk, f = s.restoreCategories(data.Categories, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 3. Subcategorias (sem FK direta para category)
	s.logger.Subsection("Subcategorias")
	c, sk, f = s.restoreSubcategories(data.Subcategories, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 3b. Vínculos N:N subcategoria → categoria
	if len(data.SubcategoryCategories) > 0 {
		s.logger.Subsection("Subcategoria-Categoria (N:N)")
		c, sk, f = s.restoreSubcategoryCategories(data.SubcategoryCategories, idMap)
		stats.Created += c; stats.Skipped += sk; stats.Failed += f
	}

	// 4. Tags (sem dependências)
	s.logger.Subsection("Tags")
	c, sk, f = s.restoreTags(data.Tags, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 5. Produtos (dependem de menu, category, subcategory) + associação de tags
	s.logger.Subsection("Produtos")
	c, sk, f = s.restoreProducts(data.Products, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 6. Ambientes (sem dependências)
	s.logger.Subsection("Ambientes")
	c, sk, f = s.restoreEnvironments(data.Environments, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 7. Mesas (dependem de environment_id)
	s.logger.Subsection("Mesas")
	c, sk, f = s.restoreTables(data.Tables, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 8. Clientes (sem dependências)
	s.logger.Subsection("Clientes")
	c, sk, f = s.restoreCustomers(data.Customers, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 9. Reservas (dependem de customer_id, table_id)
	s.logger.Subsection("Reservas")
	c, sk, f = s.restoreReservations(data.Reservations, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 10. Pedidos (dependem de table_id, customer_id)
	s.logger.Subsection("Pedidos")
	c, sk, f = s.restoreOrders(data.Orders, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 11. Waitlist (dependem de customer_id)
	s.logger.Subsection("Waitlist")
	c, sk, f = s.restoreWaitlist(data.Waitlist, idMap)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 12. Configurações (não possuem dependências externas, apenas PUT)
	s.logger.Subsection("Configurações")
	s.restoreSettings(data.Settings, data.DisplaySettings, data.Theme)

	// 13. Campanhas
	s.logger.Subsection("Campanhas")
	c, sk, f = s.restoreSimpleList("/campaign", data.Campaigns, "campanha")
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	// 14. Staff
	s.logger.Subsection("Staff")
	c, sk, f = s.restoreStaff(data)
	stats.Created += c; stats.Skipped += sk; stats.Failed += f

	return stats, nil
}

// ---- Restore por entidade ----

func (s *RestoreService) restoreMenus(menus []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, menu := range menus {
		oldID, _ := menu["id"].(string)
		payload := stripServerFields(menu)

		resp, err := s.client.Post("/manage/menu", payload)
		if err != nil {
			s.logger.Error("  Menu '%v': %v", menu["name"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Menus[oldID] = newID
		}
		s.logger.Success("  Menu '%v' criado (ID: %s)", menu["name"], newID)
		created++
	}
	return
}

func (s *RestoreService) restoreCategories(categories []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, cat := range categories {
		oldID, _ := cat["id"].(string)
		payload := stripServerFields(cat)

		s.migrateImageInPayload(payload)

		// Remap menu_id
		if oldMenuID, ok := payload["menu_id"].(string); ok {
			if newMenuID, ok := idMap.Menus[oldMenuID]; ok {
				payload["menu_id"] = newMenuID
			}
		}

		resp, err := s.client.Post("/manage/category", payload)
		if err != nil {
			s.logger.Error("  Categoria '%v': %v", cat["name"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Categories[oldID] = newID
		}
		s.logger.Success("  Categoria '%v' criada (ID: %s)", cat["name"], newID)
		created++
	}
	return
}

func (s *RestoreService) restoreSubcategories(subcats []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, sub := range subcats {
		oldID, _ := sub["id"].(string)
		payload := stripServerFields(sub)

		resp, err := s.client.Post("/manage/subcategory", payload)
		if err != nil {
			s.logger.Error("  Subcategoria '%v': %v", sub["name"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Subcategories[oldID] = newID
		}
		s.logger.Success("  Subcategoria '%v' criada (ID: %s)", sub["name"], newID)
		created++
	}
	return
}

// restoreSubcategoryCategories recria os vínculos N:N entre subcategorias e categorias
func (s *RestoreService) restoreSubcategoryCategories(links []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, link := range links {
		oldSubID, _ := link["subcategory_id"].(string)
		newSubID, ok := idMap.Subcategories[oldSubID]
		if !ok || newSubID == "" {
			s.logger.Debug("  Subcategoria %s não mapeada, pulando links", oldSubID)
			skipped++
			continue
		}

		catIDs, _ := link["category_ids"].([]interface{})
		for _, rawCatID := range catIDs {
			oldCatID, _ := rawCatID.(string)
			newCatID, ok := idMap.Categories[oldCatID]
			if !ok || newCatID == "" {
				s.logger.Debug("  Categoria %s não mapeada para subcategoria %s", oldCatID, oldSubID)
				skipped++
				continue
			}

			path := fmt.Sprintf("/manage/subcategory/%s/category/%s", newSubID, newCatID)
			if _, err := s.client.Post(path, nil); err != nil {
				s.logger.Error("  Link subcat %s → cat %s: %v", newSubID, newCatID, err)
				failed++
				continue
			}
			s.logger.Success("  Link subcategoria %s → categoria %s criado", newSubID, newCatID)
			created++
		}
	}
	return
}

func (s *RestoreService) restoreTags(tags []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, tag := range tags {
		oldID, _ := tag["id"].(string)
		payload := stripServerFields(tag)

		resp, err := s.client.Post("/tag", payload)
		if err != nil {
			s.logger.Error("  Tag '%v': %v", tag["name"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Tags[oldID] = newID
		}
		s.logger.Success("  Tag '%v' criada (ID: %s)", tag["name"], newID)
		created++
	}
	return
}

func (s *RestoreService) restoreProducts(products []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, prod := range products {
		oldID, _ := prod["id"].(string)
		payload := stripServerFields(prod)

		// Extrair tags antes de criar o produto
		var oldTagIDs []string
		if tags, ok := prod["tags"].([]interface{}); ok {
			for _, t := range tags {
				if tm, ok := t.(map[string]interface{}); ok {
					if tid, ok := tm["id"].(string); ok {
						oldTagIDs = append(oldTagIDs, tid)
					}
				}
			}
		}
		// Remover tags do payload (associação é feita separadamente)
		delete(payload, "tags")

		s.migrateImageInPayload(payload)

		// Remap IDs de referência
		remapField(payload, "menu_id", idMap.Menus)
		remapField(payload, "category_id", idMap.Categories)
		remapField(payload, "subcategory_id", idMap.Subcategories)

		resp, err := s.client.Post("/product", payload)
		if err != nil {
			s.logger.Error("  Produto '%v': %v", prod["name"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Products[oldID] = newID
		}
		s.logger.Success("  Produto '%v' criado (ID: %s)", prod["name"], newID)
		created++

		// Associar tags ao produto
		for _, oldTagID := range oldTagIDs {
			newTagID, ok := idMap.Tags[oldTagID]
			if !ok {
				s.logger.Debug("  Tag %s não mapeada para produto '%v'", oldTagID, prod["name"])
				continue
			}
			if _, err := s.client.Post(fmt.Sprintf("/product/%s/tags/%s", newID, newTagID), nil); err != nil {
				s.logger.Warn("  Erro ao associar tag %s ao produto %s: %v", newTagID, newID, err)
			}
		}
	}
	return
}

func (s *RestoreService) restoreEnvironments(environments []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, env := range environments {
		oldID, _ := env["id"].(string)
		payload := stripServerFields(env)

		resp, err := s.client.Post("/environment", payload)
		if err != nil {
			s.logger.Error("  Ambiente '%v': %v", env["name"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Environments[oldID] = newID
		}
		s.logger.Success("  Ambiente '%v' criado (ID: %s)", env["name"], newID)
		created++
	}
	return
}

func (s *RestoreService) restoreTables(tables []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, table := range tables {
		oldID, _ := table["id"].(string)
		payload := stripServerFields(table)

		remapField(payload, "environment_id", idMap.Environments)

		resp, err := s.client.Post("/table", payload)
		if err != nil {
			s.logger.Error("  Mesa '%v': %v", table["number"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Tables[oldID] = newID
		}
		s.logger.Success("  Mesa '%v' criada (ID: %s)", table["number"], newID)
		created++
	}
	return
}

func (s *RestoreService) restoreCustomers(customers []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, customer := range customers {
		oldID, _ := customer["id"].(string)
		payload := stripServerFields(customer)

		resp, err := s.client.Post("/customer", payload)
		if err != nil {
			s.logger.Error("  Cliente '%v': %v", customer["name"], err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Customers[oldID] = newID
		}
		s.logger.Success("  Cliente '%v' criado (ID: %s)", customer["name"], newID)
		created++
	}
	return
}

func (s *RestoreService) restoreReservations(reservations []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, res := range reservations {
		payload := stripServerFields(res)

		remapField(payload, "customer_id", idMap.Customers)
		remapField(payload, "table_id", idMap.Tables)

		if _, err := s.client.Post("/reservation", payload); err != nil {
			s.logger.Error("  Reserva: %v", err)
			failed++
			continue
		}
		s.logger.Success("  Reserva criada")
		created++
	}
	return
}

func (s *RestoreService) restoreOrders(orders []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, order := range orders {
		oldID, _ := order["id"].(string)
		payload := stripServerFields(order)

		remapField(payload, "table_id", idMap.Tables)
		remapField(payload, "customer_id", idMap.Customers)

		// Remap IDs de itens de pedido, se existirem
		if items, ok := payload["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					remapField(itemMap, "product_id", idMap.Products)
				}
			}
		}

		resp, err := s.client.Post("/order", payload)
		if err != nil {
			s.logger.Error("  Pedido: %v", err)
			failed++
			continue
		}

		newID := extractNewID(resp)
		if newID != "" && oldID != "" {
			idMap.Orders[oldID] = newID
		}
		s.logger.Success("  Pedido criado (ID: %s)", newID)
		created++
	}
	return
}

func (s *RestoreService) restoreWaitlist(waitlist []map[string]interface{}, idMap *IDMap) (created, skipped, failed int) {
	for _, entry := range waitlist {
		payload := stripServerFields(entry)
		remapField(payload, "customer_id", idMap.Customers)

		if _, err := s.client.Post("/waitlist", payload); err != nil {
			s.logger.Error("  Waitlist: %v", err)
			failed++
			continue
		}
		s.logger.Success("  Waitlist entry criada")
		created++
	}
	return
}

func (s *RestoreService) restoreSettings(settings, displaySettings, theme map[string]interface{}) {
	if settings != nil {
		payload := stripServerFields(settings)
		if _, err := s.client.Put("/settings", payload); err != nil {
			s.logger.Error("  Settings: %v", err)
		} else {
			s.logger.Success("  Settings atualizadas")
		}
	}

	if displaySettings != nil {
		payload := stripServerFields(displaySettings)
		if _, err := s.client.Put("/project/settings/display", payload); err != nil {
			s.logger.Error("  Display Settings: %v", err)
		} else {
			s.logger.Success("  Display Settings atualizadas")
		}
	}

	if theme != nil {
		payload := stripServerFields(theme)
		if _, err := s.client.Put("/project/settings/theme", payload); err != nil {
			s.logger.Error("  Theme: %v", err)
		} else {
			s.logger.Success("  Theme atualizado")
		}
	}
}

func (s *RestoreService) restoreSimpleList(endpoint string, items []map[string]interface{}, label string) (created, skipped, failed int) {
	for _, item := range items {
		payload := stripServerFields(item)
		if _, err := s.client.Post(endpoint, payload); err != nil {
			s.logger.Error("  %s '%v': %v", label, item["name"], err)
			failed++
			continue
		}
		s.logger.Success("  %s '%v' criado", label, item["name"])
		created++
	}
	return
}

func (s *RestoreService) restoreStaff(data ProjectData) (created, skipped, failed int) {
	type staffEntry struct {
		path  string
		items []map[string]interface{}
		label string
	}
	entries := []staffEntry{
		{"/staff/availability", data.StaffAvailability, "disponibilidade"},
		{"/staff/schedule", data.StaffSchedules, "escala"},
		{"/staff/attendance", data.StaffAttendance, "presença"},
		{"/staff/commission", data.StaffCommissions, "comissão"},
		{"/staff/stock", data.StaffStock, "estoque"},
	}
	for _, e := range entries {
		c, sk, f := s.restoreSimpleList(e.path, e.items, e.label)
		created += c; skipped += sk; failed += f
	}
	return
}

// ---- Migração de imagens ----

// migrateImageURL lê a imagem (local ou remota) e faz upload no destino.
// Retorna a nova URL no destino. Em caso de falha, retorna a URL original com aviso.
func (s *RestoreService) migrateImageURL(originalURL string) string {
	if originalURL == "" {
		return ""
	}

	category := imageCategoryFromURL(originalURL)
	filename := imageFilenameFromURL(originalURL)

	var imageBytes []byte
	var err error

	if isLocalPath(originalURL) {
		// Ler imagem do disco local
		s.logger.Debug("  Lendo imagem local: %s", originalURL)
		imageBytes, err = os.ReadFile(originalURL)
		if err != nil {
			s.logger.Warn("  Imagem local não encontrada (%s): %v — mantendo URL original", originalURL, err)
			return originalURL
		}
	} else {
		// Baixar via HTTP
		downloadURL := originalURL
		if len(originalURL) > 0 && originalURL[0] == '/' {
			downloadURL = s.sourceURL + originalURL
		}
		s.logger.Debug("  Baixando imagem: %s", downloadURL)
		imageBytes, err = DownloadImage(downloadURL, s.timeout)
		if err != nil {
			s.logger.Warn("  Imagem não migrada (%s): %v — mantendo URL original", filename, err)
			return originalURL
		}
	}

	newURL, err := s.client.UploadImage(imageBytes, filename, category)
	if err != nil {
		s.logger.Warn("  Falha no upload da imagem (%s): %v — mantendo URL original", filename, err)
		return originalURL
	}

	s.logger.Debug("  Imagem migrada: %s → %s", filename, newURL)
	return newURL
}

// isLocalPath retorna true se o path é um arquivo local (não uma URL HTTP)
func isLocalPath(path string) bool {
	return !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://")
}

// migrateImageInPayload verifica se o payload tem image_url e migra a imagem.
// Modifica o payload in-place.
func (s *RestoreService) migrateImageInPayload(payload map[string]interface{}) {
	if url, ok := payload["image_url"].(string); ok && url != "" {
		payload["image_url"] = s.migrateImageURL(url)
	}
}

// ---- Helpers ----

// stripServerFields remove campos gerados pelo servidor que não devem ser enviados na criação
func stripServerFields(item map[string]interface{}) map[string]interface{} {
	serverFields := []string{
		"id", "organization_id", "project_id",
		"created_at", "updated_at", "deleted_at",
	}
	result := make(map[string]interface{}, len(item))
	for k, v := range item {
		result[k] = v
	}
	for _, f := range serverFields {
		delete(result, f)
	}
	return result
}

// remapField substitui o valor de um campo de ID usando o mapa de IDs
func remapField(payload map[string]interface{}, field string, idMap map[string]string) {
	if oldID, ok := payload[field].(string); ok && oldID != "" {
		if newID, ok := idMap[oldID]; ok {
			payload[field] = newID
		}
	}
}

// extractNewID extrai o ID da entidade recém-criada da resposta da API
func extractNewID(resp map[string]interface{}) string {
	if resp == nil {
		return ""
	}

	// Padrão: {"data": {"id": "..."}}
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			return id
		}
	}

	// Padrão: {"id": "..."}
	if id, ok := resp["id"].(string); ok {
		return id
	}

	return ""
}
