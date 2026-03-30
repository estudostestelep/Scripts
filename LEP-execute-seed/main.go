package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	// ====== CARREGA CONFIGURAÇÃO ======
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[✗] Erro ao carregar config: %v\n", err)
		os.Exit(1)
	}

	// ====== EXIBE CONFIGURAÇÃO ======
	fmt.Println("\n========== 🌱 LEP Database Seeder v2.0 ==========")
	fmt.Printf("[ℹ] URL Backend: %s\n", config.Server.URL)
	fmt.Printf("[ℹ] Organização: %s\n", config.Auth.OrganizationName)
	fmt.Printf("[ℹ] Log Level: %s\n", config.Logging.Level)
	fmt.Println("================================================\n")

	// ====== CRIAR LOGGER ======
	isVerbose := config.Logging.Level == "debug" || config.Logging.Level == "verbose"
	logger := NewLogger(isVerbose)

	// ====== DETERMINAR ARQUIVOS DE SEED A EXECUTAR ======
	seedFiles := determineSeedFiles(config.Seed.File, logger)
	if len(seedFiles) == 0 {
		logger.Error("Nenhum arquivo de seed encontrado para executar")
		os.Exit(1)
	}

	fmt.Printf("[ℹ] Arquivos de seed: %v\n\n", seedFiles)

	// ====== CRIAR CLIENTE DE API (COMPARTILHADO) ======
	client := NewAPIClientV2(config.Server.URL, logger, config)

	// ====== ESTADO ACUMULADO ======
	totalCreated := 0
	totalSkipped := 0
	totalFailed := 0
	allErrors := []SeedError{}

	// ====== EXECUTAR CADA ARQUIVO DE SEED ======
	for _, seedFile := range seedFiles {
		fmt.Printf("\n\n╔══════════════════════════════════════════════════════════════╗\n")
		fmt.Printf("║ Processando: %s\n", seedFile)
		fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

		// ====== CARREGAR DADOS DE SEED ======
		logger.Info(fmt.Sprintf("Carregando %s...", seedFile))
		seedData, err := LoadSeedData(seedFile)
		if err != nil {
			logger.Error(fmt.Sprintf("Erro ao carregar seed: %v", err))
			totalFailed++
			continue
		}

		totalItems := len(seedData.Menus) + len(seedData.Categories) + len(seedData.Subcategories) + len(seedData.Environments) + len(seedData.Tables) + len(seedData.Products)
		logger.Info(fmt.Sprintf("Arquivo carregado com %d items", totalItems))

		// ====== CRIAR SERVIÇO DE SEED ======
		service := &SeedServiceV2{
			client:   client,
			logger:   logger,
			config:   config,
			seedData: seedData,
			state: &SeedState{
				created: 0,
				skipped: 0,
				failed:  0,
				errors:  []SeedError{},
			},
		}

		// ====== EXECUTAR SEED ======
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		startTime := time.Now()
		err = service.Execute(ctx)
		duration := time.Since(startTime)
		cancel()

		// ====== ACUMULAR RESULTADOS ======
		totalCreated += service.state.created
		totalSkipped += service.state.skipped
		totalFailed += service.state.failed
		allErrors = append(allErrors, service.state.errors...)

		// ====== EXIBIR RESUMO PARCIAL ======
		fmt.Println("\n========== 🎉 RESUMO - " + seedFile + " ==========")
		fmt.Printf("[✓] Criados: %d\n", service.state.created)
		fmt.Printf("[⏭] Já existiam: %d\n", service.state.skipped)
		fmt.Printf("[✗] Erros: %d\n", service.state.failed)
		fmt.Printf("[⏱] Tempo: %s\n", duration)
		fmt.Println("==========================================\n")

		// Exibir erros deste arquivo se houver
		if len(service.state.errors) > 0 {
			fmt.Println("[✗] Erros detectados:")
			for _, e := range service.state.errors {
				fmt.Printf("  - [%s] %s: %s\n", e.Type, e.Item, e.Message)
			}
			fmt.Println()
		}
	}

	// ====== EXIBIR RESUMO TOTAL ======
	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║               RESUMO TOTAL DA EXECUÇÃO                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Printf("[✓] Total Criados: %d\n", totalCreated)
	fmt.Printf("[⏭] Total Já Existiam: %d\n", totalSkipped)
	fmt.Printf("[✗] Total Erros: %d\n", totalFailed)
	fmt.Println()

	// ====== EXIBIR TODOS OS ERROS SE HOUVER ======
	if len(allErrors) > 0 {
		fmt.Println("[✗] Erros detectados no total:")
		for _, e := range allErrors {
			fmt.Printf("  - [%s] %s: %s\n", e.Type, e.Item, e.Message)
		}
		fmt.Println()
	}

	// ====== SAIR COM STATUS CORRETO ======
	if totalFailed > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

// determineSeedFiles retorna lista de arquivos de seed a executar
// Se o arquivo na config for específico (com -file), executa apenas ele
// Caso contrário, executa ambos os arquivos padrão: seed-fattoria.json e seed-data.json
func determineSeedFiles(configFile string, logger *Logger) []string {
	// Se foi passado -file na CLI, retorna apenas esse arquivo
	if configFile != "seed-fattoria.json" {
		if _, err := os.Stat(configFile); err == nil {
			return []string{configFile}
		}
		logger.Error(fmt.Sprintf("Arquivo especificado não encontrado: %s", configFile))
		return []string{}
	}

	// Senão, executa os dois arquivos padrão se existirem
	defaultFiles := []string{"seed-fattoria.json", "seed-data.json"}
	var availableFiles []string

	for _, file := range defaultFiles {
		if _, err := os.Stat(file); err == nil {
			availableFiles = append(availableFiles, file)
		} else {
			logger.Error(fmt.Sprintf("Arquivo não encontrado: %s (será ignorado)", file))
		}
	}

	return availableFiles
}

// SeedServiceV2 gerencia a execução do seed contra o backend
type SeedServiceV2 struct {
	client   *APIClientV2
	logger   *Logger
	config   *Config
	seedData *SeedData
	state    *SeedState
}

// SeedState rastreia o estado da execução
type SeedState struct {
	created int
	skipped int
	failed  int
	errors  []SeedError
}

// SeedError representa um erro durante execução
type SeedError struct {
	Type    string // auth, menu, category, etc
	Item    string // nome do item
	Message string // mensagem de erro
}

// Execute executa o seed completo
func (s *SeedServiceV2) Execute(ctx context.Context) error {
	// PASSO 1: Criar/Obter Organização e Fazer Login
	fmt.Println("========== Passo 1: Criando Organização ==========")
	orgID, projID, _, err := s.createOrganization()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Erro ao criar organização: %v", err))
		s.state.failed++
		s.state.errors = append(s.state.errors, SeedError{
			Type:    "org",
			Item:    s.config.Auth.OrganizationName,
			Message: err.Error(),
		})
		return err
	}

	s.logger.Info(fmt.Sprintf("Organização OK (ID: %s)", orgID))
	s.state.created++

	// PASSO 2: Fazer Login com Master Admin (pablo@lep.com)
	// IMPORTANTE: Sempre usar master_admin para executar seed, pois tem acesso total
	fmt.Println("\n========== Passo 2: Fazendo Login com Master Admin ==========")

	// Sempre logar com master_admin (fallback_email) para ter permissões totais
	masterAdminEmail := s.config.Auth.FallbackEmail
	masterAdminPassword := s.config.Auth.FallbackPassword

	s.logger.Info(fmt.Sprintf("Usando master_admin: %s", masterAdminEmail))

	// NOVO: Tentar login com /admin/login primeiro (novo sistema)
	_, err = s.client.AdminLogin(masterAdminEmail, masterAdminPassword)
	if err != nil {
		s.logger.Info(fmt.Sprintf("Admin login falhou, tentando login legado: %v", err))
		// Fallback: Login com sistema legado buscando a organização específica
		orgID, projID, err = s.client.LoginAndGetIDsForOrg(masterAdminEmail, masterAdminPassword, s.config.Auth.OrganizationName)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Erro ao fazer login com master_admin: %v", err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "auth",
				Item:    masterAdminEmail,
				Message: err.Error(),
			})
			return err
		}
		s.logger.Info(fmt.Sprintf("Autenticado via legado como %s", masterAdminEmail))
		s.client.SetHeaders(s.client.token, orgID, projID)
	} else {
		s.logger.Info(fmt.Sprintf("Autenticado via /admin/login como %s (admin)", masterAdminEmail))
		// Admin tem acesso a todas as orgs/projetos, mas precisamos do orgID/projID para headers
		// Buscar via login legado para obter os IDs
		orgID, projID, err = s.client.LoginAndGetIDsForOrg(masterAdminEmail, masterAdminPassword, s.config.Auth.OrganizationName)
		if err != nil {
			s.logger.Info(fmt.Sprintf("Não foi possível obter IDs via legado, continuando sem headers: %v", err))
		} else {
			s.client.SetHeaders(s.client.token, orgID, projID)
		}
	}

	// PASSO 2.5: Garantir Assinatura Enterprise
	fmt.Println("\n========== Passo 2.5: Garantindo Plano Enterprise ==========")
	err = s.client.EnsureEnterpriseSubscription()
	if err != nil {
		// Não é erro fatal - pode ser que o endpoint não exista ou usuário não tenha permissão
		s.logger.Info(fmt.Sprintf("Aviso: Não foi possível garantir assinatura Enterprise: %v", err))
		s.logger.Info("Continuando execução - usuário pode ser master_admin ou endpoint não implementado")
	} else {
		s.logger.Info("Plano Enterprise verificado/ativado com sucesso")
		s.state.created++
	}

	// PASSO 3: Criar Menus
	fmt.Println("\n========== Passo 3: Criando Menus ==========")
	menuIDs := make(map[int]string) // idx -> UUID
	for idx, menu := range s.seedData.Menus {
		// Verificar se menu já existe
		existingID, err := s.client.GetMenuByName(menu.Name)
		if err == nil && existingID != uuid.Nil {
			menuIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Menu %s já existe", menu.Name))
			s.state.skipped++
			continue
		}

		// Criar novo menu
		id, err := s.client.CreateMenu(menu.Name, menu.Order)
		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetMenuByName(menu.Name)
				if getErr == nil && existingID != uuid.Nil {
					menuIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Menu %s já existe (recuperado)", menu.Name))
					s.state.skipped++
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar menu %s: %v", menu.Name, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "menu",
				Item:    menu.Name,
				Message: err.Error(),
			})
		} else {
			menuIDs[idx] = id.String()
			s.logger.Info(fmt.Sprintf("Menu criado: %s", menu.Name))
			s.state.created++
		}
	}

	// PASSO 4: Criar Categorias
	fmt.Println("\n========== Passo 4: Criando Categorias ==========")
	categoryIDs := make(map[int]string) // idx -> UUID
	for idx, cat := range s.seedData.Categories {
		menuID, ok := menuIDs[cat.MenuIDRef]
		if !ok {
			s.logger.Error(fmt.Sprintf("Menu não encontrado para categoria %s", cat.Name))
			s.state.failed++
			continue
		}

		// Verificar se categoria já existe
		existingID, err := s.client.GetCategoryByName(cat.Name)
		if err == nil && existingID != uuid.Nil {
			categoryIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Categoria %s já existe", cat.Name))
			s.state.skipped++
			continue
		}

		// Se não existe, criar nova
		id, err := s.client.CreateCategory(menuID, cat.Name, cat.Order)
		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetCategoryByName(cat.Name)
				if getErr == nil && existingID != uuid.Nil {
					categoryIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Categoria %s já existe (recuperado)", cat.Name))
					s.state.skipped++
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar categoria %s: %v", cat.Name, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "category",
				Item:    cat.Name,
				Message: err.Error(),
			})
		} else {
			categoryIDs[idx] = id.String()
			s.logger.Info(fmt.Sprintf("Categoria criada: %s", cat.Name))
			s.state.created++
		}
	}

	// PASSO 5: Criar Subcategorias
	fmt.Println("\n========== Passo 5: Criando Subcategorias ==========")
	subcategoryIDs := make(map[int]string) // idx -> UUID
	for idx, subcat := range s.seedData.Subcategories {
		catID, ok := categoryIDs[subcat.CategoryIDRef]
		if !ok {
			s.logger.Error(fmt.Sprintf("Categoria não encontrada para subcategoria %s", subcat.Name))
			s.state.failed++
			continue
		}

		// Verificar se subcategoria já existe
		existingID, err := s.client.GetSubcategoryByName(subcat.Name)
		if err == nil && existingID != uuid.Nil {
			subcategoryIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Subcategoria %s já existe", subcat.Name))
			s.state.skipped++
			// Ainda precisamos vincular à categoria
			err = s.client.AddCategoryToSubcategory(existingID.String(), catID)
			if err != nil {
				s.logger.Debug(fmt.Sprintf("Relação subcategoria-categoria já existe ou erro: %v", err))
			}
			continue
		}

		id, err := s.client.CreateSubcategory(catID, subcat.Name)
		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetSubcategoryByName(subcat.Name)
				if getErr == nil && existingID != uuid.Nil {
					subcategoryIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Subcategoria %s já existe (recuperado)", subcat.Name))
					s.state.skipped++
					// Vincular à categoria
					s.client.AddCategoryToSubcategory(existingID.String(), catID)
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar subcategoria %s: %v", subcat.Name, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "subcategory",
				Item:    subcat.Name,
				Message: err.Error(),
			})
		} else {
			subcategoryIDs[idx] = id.String()
			s.logger.Info(fmt.Sprintf("Subcategoria criada: %s", subcat.Name))
			s.state.created++

			// Vincular subcategoria à categoria (relacionamento N:M)
			err = s.client.AddCategoryToSubcategory(id.String(), catID)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Erro ao vincular subcategoria %s à categoria: %v", subcat.Name, err))
				s.state.failed++
			} else {
				s.logger.Info(fmt.Sprintf("Subcategoria %s vinculada à categoria", subcat.Name))
			}
		}
	}

	// PASSO 6: Criar Ambientes
	fmt.Println("\n========== Passo 6: Criando Ambientes ==========")
	envIDs := make(map[int]string) // idx -> UUID
	for idx, env := range s.seedData.Environments {
		// Verificar se ambiente já existe
		existingID, err := s.client.GetEnvironmentByName(env.Name)
		if err == nil && existingID != uuid.Nil {
			envIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Ambiente %s já existe", env.Name))
			s.state.skipped++
			continue
		}

		id, err := s.client.CreateEnvironment(env.Name, env.Capacity)
		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetEnvironmentByName(env.Name)
				if getErr == nil && existingID != uuid.Nil {
					envIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Ambiente %s já existe (recuperado)", env.Name))
					s.state.skipped++
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar ambiente %s: %v", env.Name, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "environment",
				Item:    env.Name,
				Message: err.Error(),
			})
		} else {
			envIDs[idx] = id.String()
			s.logger.Info(fmt.Sprintf("Ambiente criado: %s", env.Name))
			s.state.created++
		}
	}

	// PASSO 7: Criar Mesas
	fmt.Println("\n========== Passo 7: Criando Mesas ==========")
	tableIDs := make(map[int]string) // idx -> UUID
	for idx, tbl := range s.seedData.Tables {
		// Verificar se mesa já existe
		existingID, err := s.client.GetTableByNumber(tbl.Number)
		if err == nil && existingID != uuid.Nil {
			tableIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Mesa %d já existe", tbl.Number))
			s.state.skipped++
			continue
		}

		var envID *string
		if tbl.EnvironmentIDRef >= 0 && tbl.EnvironmentIDRef < len(s.seedData.Environments) {
			if id, ok := envIDs[tbl.EnvironmentIDRef]; ok {
				envID = &id
			}
		}

		id, err := s.client.CreateTable(tbl.Number, tbl.Capacity, envID, "livre")
		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetTableByNumber(tbl.Number)
				if getErr == nil && existingID != uuid.Nil {
					tableIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Mesa %d já existe (recuperado)", tbl.Number))
					s.state.skipped++
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar mesa %d: %v", tbl.Number, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "table",
				Item:    fmt.Sprintf("mesa_%d", tbl.Number),
				Message: err.Error(),
			})
		} else {
			tableIDs[idx] = id.String()
			s.logger.Info(fmt.Sprintf("Mesa criada: %d", tbl.Number))
			s.state.created++
		}
	}

	// PASSO 8: Criar Produtos
	fmt.Println("\n========== Passo 8: Criando Produtos ==========")
	productIDs := make(map[int]string) // idx -> UUID (para ProductTags)
	for idx, prod := range s.seedData.Products {
		// Verificar se produto já existe
		existingID, err := s.client.GetProductByName(prod.Name)
		if err == nil && existingID != uuid.Nil {
			productIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Produto %s já existe", prod.Name))
			s.state.skipped++
			continue
		}

		var menuID, catID, subcatID *string

		// Obter menu_id
		if prod.MenuIDRef >= 0 && prod.MenuIDRef < len(s.seedData.Menus) {
			if id, ok := menuIDs[prod.MenuIDRef]; ok {
				menuID = &id
			}
		}

		if prod.CategoryIDRef >= 0 && prod.CategoryIDRef < len(s.seedData.Categories) {
			if id, ok := categoryIDs[prod.CategoryIDRef]; ok {
				catID = &id
			}
		}

		if prod.SubcategoryIDRef >= 0 && prod.SubcategoryIDRef < len(s.seedData.Subcategories) {
			if id, ok := subcategoryIDs[prod.SubcategoryIDRef]; ok {
				subcatID = &id
			}
		}

		// Preparar dados de vinho se aplicável
		var wineData *WineData
		if prod.Type == "vinho" && (prod.Vintage != "" || prod.Country != "" || prod.Region != "") {
			wineData = &WineData{
				Vintage:        prod.Vintage,
				Country:        prod.Country,
				Region:         prod.Region,
				Winery:         prod.Winery,
				WineType:       prod.WineType,
				Volume:         prod.Volume,
				AlcoholContent: prod.AlcoholContent,
				PriceBottle:    prod.PriceBottle,
				PriceGlass:     prod.PriceGlass,
			}
		}

		id, err := s.client.CreateProduct(
			prod.Name,
			prod.Type,
			prod.PriceNormal,
			prod.PrepTimeMinutes,
			menuID,
			catID,
			subcatID,
			wineData,
			prod.ImageURL,
		)

		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetProductByName(prod.Name)
				if getErr == nil && existingID != uuid.Nil {
					productIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Produto %s já existe (recuperado)", prod.Name))
					s.state.skipped++
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar produto %s: %v", prod.Name, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "product",
				Item:    prod.Name,
				Message: err.Error(),
			})
		} else {
			productIDs[idx] = id.String()
			if wineData != nil {
				s.logger.Info(fmt.Sprintf("Produto criado: %s (%s) - %s %s", prod.Name, prod.Type, prod.Country, prod.Vintage))
			} else {
				s.logger.Info(fmt.Sprintf("Produto criado: %s (%s)", prod.Name, prod.Type))
			}
			s.state.created++
		}
	}

	// PASSO 9: Criar Admins (NOVO SISTEMA)
	fmt.Println("\n========== Passo 9: Criando Admins ==========")
	adminIDs := make(map[int]string) // idx -> UUID
	if len(s.seedData.Admins) > 0 {
		for idx, admin := range s.seedData.Admins {
			// Verificar se admin já existe
			existingID, err := s.client.GetAdminByEmail(admin.Email)
			if err == nil && existingID != uuid.Nil {
				adminIDs[idx] = existingID.String()
				s.logger.Info(fmt.Sprintf("Admin %s já existe", admin.Email))
				s.state.skipped++
				continue
			}

			id, err := s.client.CreateAdmin(admin.Name, admin.Email, admin.Password, admin.Permissions)
			if err != nil {
				if err.Error() == "already_exists" {
					existingID, getErr := s.client.GetAdminByEmail(admin.Email)
					if getErr == nil && existingID != uuid.Nil {
						adminIDs[idx] = existingID.String()
						s.logger.Info(fmt.Sprintf("Admin %s já existe (recuperado)", admin.Email))
						s.state.skipped++
						continue
					}
				}
				s.logger.Error(fmt.Sprintf("Erro ao criar admin %s: %v", admin.Email, err))
				s.state.failed++
				s.state.errors = append(s.state.errors, SeedError{
					Type:    "admin",
					Item:    admin.Email,
					Message: err.Error(),
				})
			} else {
				adminIDs[idx] = id.String()
				s.logger.Info(fmt.Sprintf("Admin criado: %s", admin.Email))
				s.state.created++
			}
		}
	} else {
		s.logger.Info("Nenhum Admin definido no seed")
	}

	// PASSO 9.5: Criar Clients (NOVO SISTEMA)
	fmt.Println("\n========== Passo 9.5: Criando Clients ==========")
	clientIDs := make(map[int]string) // idx -> UUID
	if len(s.seedData.Clients) > 0 {
		for idx, client := range s.seedData.Clients {
			// Verificar se client já existe
			existingID, err := s.client.GetClientUserByEmail(client.Email)
			if err == nil && existingID != uuid.Nil {
				clientIDs[idx] = existingID.String()
				s.logger.Info(fmt.Sprintf("Client %s já existe", client.Email))
				s.state.skipped++
				continue
			}

			// Criar client com orgID e projID do contexto atual
			id, err := s.client.CreateClientUser(client.Name, client.Email, client.Password, orgID, []string{projID}, client.Permissions)
			if err != nil {
				if err.Error() == "already_exists" {
					existingID, getErr := s.client.GetClientUserByEmail(client.Email)
					if getErr == nil && existingID != uuid.Nil {
						clientIDs[idx] = existingID.String()
						s.logger.Info(fmt.Sprintf("Client %s já existe (recuperado)", client.Email))
						s.state.skipped++
						continue
					}
				}
				s.logger.Error(fmt.Sprintf("Erro ao criar client %s: %v", client.Email, err))
				s.state.failed++
				s.state.errors = append(s.state.errors, SeedError{
					Type:    "client",
					Item:    client.Email,
					Message: err.Error(),
				})
			} else {
				clientIDs[idx] = id.String()
				s.logger.Info(fmt.Sprintf("Client criado: %s (Org: %s)", client.Email, orgID))
				s.state.created++
			}
		}
	} else {
		s.logger.Info("Nenhum Client definido no seed")
	}

	// PASSO 10: Criar Usuários (LEGACY - manter para compatibilidade)
	fmt.Println("\n========== Passo 10: Criando Usuários (Legacy) ==========")
	userIDs := make(map[int]string) // idx -> UUID
	if len(s.seedData.Users) > 0 {
		for idx, user := range s.seedData.Users {
			// Verificar se usuário já existe
			existingID, err := s.client.GetUserByEmail(user.Email)
			if err == nil && existingID != uuid.Nil {
				userIDs[idx] = existingID.String()
				s.logger.Info(fmt.Sprintf("Usuário %s já existe", user.Email))
				s.state.skipped++
				continue
			}

			id, err := s.client.CreateUser(user.Name, user.Email, user.Password, user.Role, user.Permissions)
			if err != nil {
				// Se já existe, tentar buscar o ID existente
				if err.Error() == "already_exists" {
					existingID, getErr := s.client.GetUserByEmail(user.Email)
					if getErr == nil && existingID != uuid.Nil {
						userIDs[idx] = existingID.String()
						s.logger.Info(fmt.Sprintf("Usuário %s já existe (recuperado)", user.Email))
						s.state.skipped++
						continue
					}
				}
				s.logger.Error(fmt.Sprintf("Erro ao criar usuário %s: %v", user.Email, err))
				s.state.failed++
				s.state.errors = append(s.state.errors, SeedError{
					Type:    "user",
					Item:    user.Email,
					Message: err.Error(),
				})
			} else {
				userIDs[idx] = id.String()
				s.logger.Info(fmt.Sprintf("Usuário criado: %s (%s)", user.Email, user.Role))
				s.state.created++
			}
		}
	} else {
		s.logger.Info("Nenhum Usuário (legacy) definido no seed")
	}

	// PASSO 11: Criar Customers (clientes do restaurante)
	fmt.Println("\n========== Passo 11: Criando Customers ==========")
	customerIDs := make(map[int]string) // idx -> UUID
	for idx, cust := range s.seedData.Customers {
		// Verificar se cliente já existe
		existingID, err := s.client.GetCustomerByEmail(cust.Email)
		if err == nil && existingID != uuid.Nil {
			customerIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Cliente %s já existe", cust.Email))
			s.state.skipped++
			continue
		}

		id, err := s.client.CreateCustomer(cust.Name, cust.Email, cust.Phone, cust.BirthDate, cust.Notes)
		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetCustomerByEmail(cust.Email)
				if getErr == nil && existingID != uuid.Nil {
					customerIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Cliente %s já existe (recuperado)", cust.Email))
					s.state.skipped++
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar cliente %s: %v", cust.Email, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "customer",
				Item:    cust.Email,
				Message: err.Error(),
			})
		} else {
			customerIDs[idx] = id.String()
			s.logger.Info(fmt.Sprintf("Cliente criado: %s", cust.Email))
			s.state.created++
		}
	}

	// PASSO 12: Criar Tags
	fmt.Println("\n========== Passo 12: Criando Tags ==========")
	tagIDs := make(map[int]string) // idx -> UUID
	for idx, tag := range s.seedData.Tags {
		// Verificar se tag já existe
		existingID, err := s.client.GetTagByName(tag.Name)
		if err == nil && existingID != uuid.Nil {
			tagIDs[idx] = existingID.String()
			s.logger.Info(fmt.Sprintf("Tag %s já existe", tag.Name))
			s.state.skipped++
			continue
		}

		id, err := s.client.CreateTag(tag.Name, tag.Color, tag.Description, tag.EntityType)
		if err != nil {
			// Se já existe, tentar buscar o ID existente
			if err.Error() == "already_exists" {
				existingID, getErr := s.client.GetTagByName(tag.Name)
				if getErr == nil && existingID != uuid.Nil {
					tagIDs[idx] = existingID.String()
					s.logger.Info(fmt.Sprintf("Tag %s já existe (recuperado)", tag.Name))
					s.state.skipped++
					continue
				}
			}
			s.logger.Error(fmt.Sprintf("Erro ao criar tag %s: %v", tag.Name, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "tag",
				Item:    tag.Name,
				Message: err.Error(),
			})
		} else {
			tagIDs[idx] = id.String()
			s.logger.Info(fmt.Sprintf("Tag criada: %s", tag.Name))
			s.state.created++
		}
	}

	// PASSO 13: Criar Reservas
	fmt.Println("\n========== Passo 13: Criando Reservas ==========")
	for _, res := range s.seedData.Reservations {
		// Obter IDs dos clientes e mesas
		custID, ok := customerIDs[res.CustomerIDRef]
		if !ok {
			s.logger.Error(fmt.Sprintf("Cliente não encontrado para reserva"))
			s.state.failed++
			continue
		}

		tblID, ok := tableIDs[res.TableIDRef]
		if !ok {
			s.logger.Error(fmt.Sprintf("Mesa não encontrada para reserva"))
			s.state.failed++
			continue
		}

		// Verificar se reserva já existe (pela confirmation_key)
		existingID, err := s.client.GetReservationByConfirmationKey(res.ConfirmationKey)
		if err == nil && existingID != uuid.Nil {
			s.logger.Info(fmt.Sprintf("Reserva %s já existe", res.ConfirmationKey))
			s.state.skipped++
			continue
		}

		_, err = s.client.CreateReservation(
			custID,
			tblID,
			res.DateTime,
			res.PartySize,
			res.Notes,
			res.Status,
			res.ConfirmationKey,
		)

		if err != nil {
			s.logger.Error(fmt.Sprintf("Erro ao criar reserva %s: %v", res.ConfirmationKey, err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "reservation",
				Item:    res.ConfirmationKey,
				Message: err.Error(),
			})
		} else {
			s.logger.Info(fmt.Sprintf("Reserva criada: %s (%d pessoas)", res.ConfirmationKey, res.PartySize))
			s.state.created++
		}
	}

	// PASSO 14: Criar Product Tags (relacionamento N:M)
	fmt.Println("\n========== Passo 14: Criando Product Tags ==========")
	if len(s.seedData.ProductTags) > 0 {
		for _, pt := range s.seedData.ProductTags {
			prodID, ok := productIDs[pt.ProductIDRef]
			if !ok {
				s.logger.Error(fmt.Sprintf("Produto não encontrado para tag"))
				s.state.failed++
				continue
			}

			tagID, ok := tagIDs[pt.TagIDRef]
			if !ok {
				s.logger.Error(fmt.Sprintf("Tag não encontrada para produto"))
				s.state.failed++
				continue
			}

			err := s.client.AddTagToProduct(prodID, tagID)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Erro ao vincular tag ao produto: %v", err))
				s.state.failed++
			} else {
				s.logger.Info(fmt.Sprintf("Tag vinculada ao produto"))
				s.state.created++
			}
		}
	} else {
		s.logger.Info("Nenhum ProductTag definido no seed")
	}

	// PASSO 15: Criar Settings
	fmt.Println("\n========== Passo 15: Criando Settings ==========")
	if s.seedData.Settings.Timezone != "" || s.seedData.Settings.ReservationMinAdvanceHours > 0 {
		err := s.client.CreateSettings(&s.seedData.Settings)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Erro ao criar settings: %v", err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "settings",
				Item:    "project_settings",
				Message: err.Error(),
			})
		} else {
			s.logger.Info("Settings criado/atualizado com sucesso")
			s.state.created++
		}
	} else {
		s.logger.Info("Nenhum Settings definido no seed")
	}

	// PASSO 16: Criar Notification Templates
	fmt.Println("\n========== Passo 16: Criando Notification Templates ==========")
	if len(s.seedData.NotificationTemplates) > 0 {
		for _, tmpl := range s.seedData.NotificationTemplates {
			// Verificar se template já existe
			existingID, err := s.client.GetNotificationTemplateByName(tmpl.Name)
			if err == nil && existingID != uuid.Nil {
				s.logger.Info(fmt.Sprintf("Template %s já existe", tmpl.Name))
				s.state.skipped++
				continue
			}

			_, err = s.client.CreateNotificationTemplate(&tmpl)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Erro ao criar template %s: %v", tmpl.Name, err))
				s.state.failed++
				s.state.errors = append(s.state.errors, SeedError{
					Type:    "notification_template",
					Item:    tmpl.Name,
					Message: err.Error(),
				})
			} else {
				s.logger.Info(fmt.Sprintf("Template criado: %s (%s)", tmpl.Name, tmpl.Channel))
				s.state.created++
			}
		}
	} else {
		s.logger.Info("Nenhum NotificationTemplate definido no seed")
	}

	// PASSO 17: Criar Theme Customization
	fmt.Println("\n========== Passo 17: Criando Theme Customization ==========")
	if s.seedData.ThemeCustomization.PrimaryColor != "" {
		err := s.client.CreateThemeCustomization(&s.seedData.ThemeCustomization)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Erro ao criar theme customization: %v", err))
			s.state.failed++
			s.state.errors = append(s.state.errors, SeedError{
				Type:    "theme",
				Item:    "theme_customization",
				Message: err.Error(),
			})
		} else {
			s.logger.Info("Theme Customization criado/atualizado com sucesso")
			s.state.created++
		}
	} else {
		s.logger.Info("Nenhum ThemeCustomization definido no seed")
	}

	// PASSO 18: Importar Staff Sales Records (CSV)
	fmt.Println("\n========== Passo 18: Importando Staff Sales Records ==========")
	if len(s.seedData.StaffSalesRecordsCSV) > 0 {
		for _, csvData := range s.seedData.StaffSalesRecordsCSV {
			var csvContent []byte
			var loadErr error

			if csvData.FilePathRel != "" {
				csvContent, loadErr = os.ReadFile(csvData.FilePathRel)
				if loadErr != nil {
					s.logger.Error(fmt.Sprintf("Erro ao ler CSV %s: %v", csvData.FilePathRel, loadErr))
					s.state.failed++
					s.state.errors = append(s.state.errors, SeedError{
						Type:    "staff_sales_csv",
						Item:    csvData.FileName,
						Message: loadErr.Error(),
					})
					continue
				}
			} else if csvData.FileContent != "" {
				csvContent = []byte(csvData.FileContent)
			} else {
				s.logger.Error(fmt.Sprintf("CSV %s sem conteúdo ou caminho", csvData.FileName))
				s.state.failed++
				continue
			}

			result, err := s.client.ImportStaffSalesCSV(csvData.FileName, csvContent)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Erro ao importar CSV %s: %v", csvData.FileName, err))
				s.state.failed++
				s.state.errors = append(s.state.errors, SeedError{
					Type:    "staff_sales_csv",
					Item:    csvData.FileName,
					Message: err.Error(),
				})
			} else {
				if len(result.Errors) > 0 {
					s.logger.Info(fmt.Sprintf("CSV importado com avisos: %s (%d registros, %d erros)",
						csvData.FileName, result.RecordsCount, len(result.Errors)))
					for _, e := range result.Errors {
						s.logger.Debug(fmt.Sprintf("  - %s", e))
					}
				} else {
					s.logger.Info(fmt.Sprintf("CSV importado: %s (%d registros)", csvData.FileName, result.RecordsCount))
				}
				s.state.created += result.RecordsCount
			}
		}
	} else {
		s.logger.Info("Nenhum Staff Sales CSV definido no seed")
	}

	return nil
}

// createOrganization cria organização ou faz login se existir
func (s *SeedServiceV2) createOrganization() (orgID, projID, email string, err error) {
	email = s.config.GetAutoEmail()
	password := "senha123"

	// Tentar criar
	orgID, projID, err = s.client.CreateOrganization(
		s.config.Auth.OrganizationName,
		email,
		password,
	)

	if err != nil {
		// Se falhou, tentar login com o email que tentamos criar
		// (pois a organização pode já existir com essas credenciais)
		s.logger.Info(fmt.Sprintf("Organização pode já existir, tentando login com %s", email))
		orgID, projID, err = s.client.LoginAndGetIDs(email, password)

		if err != nil {
			// Se ainda falhar, tentar com fallback
			// IMPORTANTE: Usar LoginAndGetIDsForOrg para buscar especificamente a organização "LEP Fattoria"
			s.logger.Info(fmt.Sprintf("Tentando fallback com %s para organização '%s'", s.config.Auth.FallbackEmail, s.config.Auth.OrganizationName))
			orgID, projID, err = s.client.LoginAndGetIDsForOrg(
				s.config.Auth.FallbackEmail,
				s.config.Auth.FallbackPassword,
				s.config.Auth.OrganizationName,
			)
			if err == nil {
				email = s.config.Auth.FallbackEmail
			}
		}
	}

	return orgID, projID, email, err
}

// login faz login de um usuário
func (s *SeedServiceV2) login(email string) error {
	password := "senha123"

	_, _, err := s.client.LoginAndGetIDs(email, password)
	return err
}

// LoadSeedData carrega dados do arquivo JSON
func LoadSeedData(filename string) (*SeedData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("erro ao parsear JSON: %w", err)
	}

	return &seedData, nil
}
