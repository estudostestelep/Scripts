package main

// MigrationDump representa o dump completo de todos os dados do sistema
type MigrationDump struct {
	Metadata      DumpMetadata       `json:"metadata"`
	Organizations []OrganizationDump `json:"organizations"`
}

// DumpMetadata contém informações sobre o dump
type DumpMetadata struct {
	DumpedAt    string `json:"dumped_at"`
	SourceURL   string `json:"source_url"`
	ToolVersion string `json:"tool_version"`
}

// OrganizationDump representa uma organização com todos os seus projetos
type OrganizationDump struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Email    string        `json:"email"`
	Projects []ProjectDump `json:"projects"`
}

// ProjectDump representa um projeto com todos os seus dados
type ProjectDump struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Data ProjectData `json:"data"`
}

// ProjectData contém todos os dados de um projeto agrupados por entidade.
// Cada entidade é armazenada como map[string]interface{} para preservar
// o contrato exato da API sem necessidade de tipagem rígida.
type ProjectData struct {
	Menus             []map[string]interface{} `json:"menus"`
	Categories        []map[string]interface{} `json:"categories"`
	Subcategories     []map[string]interface{} `json:"subcategories"`
	Tags              []map[string]interface{} `json:"tags"`
	Products          []map[string]interface{} `json:"products"`
	Environments      []map[string]interface{} `json:"environments"`
	Tables            []map[string]interface{} `json:"tables"`
	Customers         []map[string]interface{} `json:"customers"`
	Reservations      []map[string]interface{} `json:"reservations"`
	Orders            []map[string]interface{} `json:"orders"`
	Waitlist          []map[string]interface{} `json:"waitlist"`
	Settings          map[string]interface{}   `json:"settings"`
	DisplaySettings   map[string]interface{}   `json:"display_settings"`
	Theme             map[string]interface{}   `json:"theme"`
	Roles             []map[string]interface{} `json:"roles"`
	Campaigns         []map[string]interface{} `json:"campaigns"`
	StaffAvailability []map[string]interface{} `json:"staff_availability"`
	StaffSchedules    []map[string]interface{} `json:"staff_schedules"`
	StaffAttendance   []map[string]interface{} `json:"staff_attendance"`
	StaffCommissions  []map[string]interface{} `json:"staff_commissions"`
	StaffStock        []map[string]interface{} `json:"staff_stock"`

	// Relações N:N entre subcategorias e categorias
	// Cada entrada: {"subcategory_id": "...", "category_ids": ["...", ...]}
	SubcategoryCategories []map[string]interface{} `json:"subcategory_categories"`
}

// IDMap mantém o mapeamento de IDs antigos (source) para novos (destino)
// durante o processo de restore, permitindo resolver referências entre entidades.
type IDMap struct {
	Menus         map[string]string
	Categories    map[string]string
	Subcategories map[string]string
	Tags          map[string]string
	Products      map[string]string
	Environments  map[string]string
	Tables        map[string]string
	Customers     map[string]string
	Orders        map[string]string
}

// NewIDMap cria um IDMap vazio
func NewIDMap() *IDMap {
	return &IDMap{
		Menus:         make(map[string]string),
		Categories:    make(map[string]string),
		Subcategories: make(map[string]string),
		Tags:          make(map[string]string),
		Products:      make(map[string]string),
		Environments:  make(map[string]string),
		Tables:        make(map[string]string),
		Customers:     make(map[string]string),
		Orders:        make(map[string]string),
	}
}

// RestoreStats acumula estatísticas do processo de restore
type RestoreStats struct {
	Created int
	Skipped int
	Failed  int
}

func (s *RestoreStats) Add(other RestoreStats) {
	s.Created += other.Created
	s.Skipped += other.Skipped
	s.Failed += other.Failed
}
