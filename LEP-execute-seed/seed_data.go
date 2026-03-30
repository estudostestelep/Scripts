package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type OrgData struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Website     string `json:"website"`
	Description string `json:"description"`
	Active      bool   `json:"active,omitempty"`
}

type MenuData struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	Active            bool   `json:"active"`
	Order             int    `json:"order"`
	// ✨ Novos campos para seleção inteligente
	Priority          int    `json:"priority,omitempty"`
	TimeRangeStart    string `json:"time_range_start,omitempty"`
	TimeRangeEnd      string `json:"time_range_end,omitempty"`
	ApplicableDays    string `json:"applicable_days,omitempty"`
	ApplicableDates   string `json:"applicable_dates,omitempty"`
	IsManualOverride  bool   `json:"is_manual_override,omitempty"`
}

type CategoryData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MenuIDRef   int    `json:"menu_id_ref"`
	Active      bool   `json:"active"`
	Order       int    `json:"order"`
}

type SubcategoryData struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	CategoryIDRef int    `json:"category_id_ref"`
	Active        bool   `json:"active"`
	Order         int    `json:"order"`
}

type EnvironmentData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
	Active      bool   `json:"active"`
}

type TableData struct {
	Number           int    `json:"number"`
	Capacity         int    `json:"capacity"`
	Location         string `json:"location"`
	Status           string `json:"status"`
	EnvironmentIDRef int    `json:"environment_id_ref"`
}

type ProductData struct {
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Type             string  `json:"type"`
	ImageURL         string  `json:"image_url,omitempty"`
	PriceNormal      float64 `json:"price_normal"`
	PricePromo       float64 `json:"price_promo,omitempty"`
	PriceGlass       float64 `json:"price_glass,omitempty"`
	PriceBottle      float64 `json:"price_bottle,omitempty"`
	PriceHalfBottle  float64 `json:"price_half_bottle,omitempty"`
	MenuIDRef        int     `json:"menu_id_ref"`
	CategoryIDRef    int     `json:"category_id_ref"`
	SubcategoryIDRef int     `json:"subcategory_id_ref,omitempty"`
	Active           bool    `json:"active"`
	Order            int     `json:"order"`
	PrepTimeMinutes  int     `json:"prep_time_minutes"`
	// Campos específicos para vinhos
	Vintage          string  `json:"vintage,omitempty"`
	Country          string  `json:"country,omitempty"`
	Region           string  `json:"region,omitempty"`
	Winery           string  `json:"winery,omitempty"`
	WineType         string  `json:"wine_type,omitempty"`
	Volume           int     `json:"volume,omitempty"`
	AlcoholContent   float64 `json:"alcohol_content,omitempty"`
}

type SettingsData struct {
	ReservationMinAdvanceHours int    `json:"reservation_min_advance_hours"`
	ReservationMaxAdvanceDays  int    `json:"reservation_max_advance_days"`
	NotifyReservationCreate    bool   `json:"notify_reservation_create"`
	NotifyReservationUpdate    bool   `json:"notify_reservation_update"`
	NotifyReservationCancel    bool   `json:"notify_reservation_cancel"`
	NotifyTableAvailable       bool   `json:"notify_table_available"`
	NotifyConfirmation24h      bool   `json:"notify_confirmation_24h"`
	DefaultNotificationChannel string `json:"default_notification_channel"`
	EnableSMS                  bool   `json:"enable_sms"`
	EnableEmail                bool   `json:"enable_email"`
	EnableWhatsApp             bool   `json:"enable_whatsapp"`
	Timezone                   string `json:"timezone"`
	// ✨ Novos campos para sistema de notificações aprimorado
	// Timing flexível de notificações
	ConfirmationHoursBefore    int    `json:"confirmation_hours_before,omitempty"`     // Horas antes para pedir confirmação (padrão: 24)
	ReminderHoursBefore        int    `json:"reminder_hours_before,omitempty"`         // Horas antes para lembrete (padrão: 2, 0=desativado)
	AutoCancelNoResponseHours  int    `json:"auto_cancel_no_response_hours,omitempty"` // Horas sem resposta para cancelar (0=desativado)
	// Modo de processamento de respostas
	ResponseProcessingMode     string `json:"response_processing_mode,omitempty"`      // "automatic", "ai_assisted", "manual"
}

type NotificationTemplateData struct {
	Name    string `json:"name"`
	Channel string `json:"channel"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Active  bool   `json:"active"`
}

type ThemeCustomizationData struct {
	PrimaryColor           string  `json:"primary_color"`
	SecondaryColor         string  `json:"secondary_color"`
	BackgroundColor        string  `json:"background_color"`
	CardBackgroundColor    string  `json:"card_background_color"`
	TextColor              string  `json:"text_color"`
	TextSecondaryColor     string  `json:"text_secondary_color"`
	AccentColor            string  `json:"accent_color"`
	SuccessColor           string  `json:"success_color,omitempty"`
	ErrorColor             string  `json:"error_color,omitempty"`
	WarningColor           string  `json:"warning_color,omitempty"`
	InfoColor              string  `json:"info_color,omitempty"`
	DisabledOpacity        float64 `json:"disabled_opacity,omitempty"`
	ShadowIntensity        float64 `json:"shadow_intensity,omitempty"`
	IsActive               bool    `json:"is_active"`
}

type UserData struct {
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Role        string   `json:"role"` // admin, manager, waiter, kitchen (LEGACY)
	Permissions []string `json:"permissions,omitempty"`
	Active      bool     `json:"active"`
}

// AdminData representa um usuário admin do sistema (tabela admins)
type AdminData struct {
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Permissions []string `json:"permissions,omitempty"` // ex: ["master_admin"]
	Active      bool     `json:"active"`
}

// ClientData representa um usuário cliente (tabela clients)
// Pertence a UMA organização e pode ter acesso a múltiplos projetos
type ClientData struct {
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Permissions []string `json:"permissions,omitempty"` // Permissões diretas
	Active      bool     `json:"active"`
	// OrgId e ProjIds são preenchidos automaticamente durante o seed
}

type CustomerData struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date,omitempty"` // YYYY-MM-DD
	Notes     string `json:"notes,omitempty"`
	Active    bool   `json:"active"`
}

type ReservationData struct {
	CustomerIDRef   int    `json:"customer_id_ref"`
	TableIDRef      int    `json:"table_id_ref"`
	DateTime        string `json:"datetime"` // ISO8601
	PartySize       int    `json:"party_size"`
	Notes           string `json:"notes,omitempty"`
	Status          string `json:"status"` // confirmed, cancelled, completed, no_show
	ConfirmationKey string `json:"confirmation_key,omitempty"`
}

type OrderItemData struct {
	ProductIDRef int     `json:"product_id_ref"`
	Quantity     int     `json:"quantity"`
	Notes        string  `json:"notes,omitempty"`
	Price        float64 `json:"price,omitempty"`
}

type OrderData struct {
	TableIDRef      int              `json:"table_id_ref,omitempty"`
	CustomerIDRef   int              `json:"customer_id_ref,omitempty"`
	Items           []OrderItemData  `json:"items"`
	Status          string           `json:"status"` // pending, preparing, ready, delivered, cancelled
	TotalAmount     float64          `json:"total_amount"`
	Notes           string           `json:"notes,omitempty"`
	PrepTimeMinutes int              `json:"prep_time_minutes,omitempty"`
	Source          string           `json:"source"` // internal, public
}

type WaitlistData struct {
	CustomerIDRef int    `json:"customer_id_ref"`
	PartySize     int    `json:"party_size"`
	Status        string `json:"status"` // waiting, seated, left
	Notes         string `json:"notes,omitempty"`
}

type TagData struct {
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"` // hex color
	Description string `json:"description,omitempty"`
	EntityType  string `json:"entity_type,omitempty"` // product, menu, etc
	Active      bool   `json:"active"`
}

type ProductTagData struct {
	ProductIDRef int `json:"product_id_ref"`
	TagIDRef     int `json:"tag_id_ref"`
}

type NotificationConfigData struct {
	EventType  string   `json:"event_type"` // reservation_created, order_ready, etc
	Enabled    bool     `json:"enabled"`
	Channels   []string `json:"channels"` // sms, email, whatsapp
	TemplateID int      `json:"template_id,omitempty"`
}

type LeadData struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Message   string `json:"message,omitempty"`
	Source    string `json:"source"` // website, phone, referral
	Status    string `json:"status"` // new, contacted, converted, rejected
	Active    bool   `json:"active"`
}

// StaffSalesRecordCSVData representa um arquivo CSV para importação de vendas
type StaffSalesRecordCSVData struct {
	FileName    string `json:"file_name"`               // Nome do arquivo para referência
	FilePathRel string `json:"file_path_rel,omitempty"` // Caminho relativo ao CSV (prioridade)
	FileContent string `json:"file_content,omitempty"`  // Conteúdo CSV direto (alternativa)
}

type SeedData struct {
	// Organization & Projects
	Organization OrgData `json:"organization"`

	// Menu System
	Menus         []MenuData        `json:"menus"`
	Categories    []CategoryData    `json:"categories"`
	Subcategories []SubcategoryData `json:"subcategories"`

	// Physical Layout
	Environments []EnvironmentData `json:"environments"`
	Tables       []TableData       `json:"tables"`

	// Products & Items
	Products []ProductData `json:"products"`
	Tags     []TagData     `json:"tags,omitempty"`

	// Users & Staff (NOVO SISTEMA)
	Admins  []AdminData  `json:"admins,omitempty"`  // Usuários admin (tabela admins)
	Clients []ClientData `json:"clients,omitempty"` // Usuários cliente (tabela clients)

	// Users & Staff (LEGACY - manter para compatibilidade)
	Users []UserData `json:"users,omitempty"`

	// Customers & Transactions
	Customers    []CustomerData    `json:"customers,omitempty"`
	Reservations []ReservationData `json:"reservations,omitempty"`
	Orders       []OrderData       `json:"orders,omitempty"`
	Waitlist     []WaitlistData    `json:"waitlist,omitempty"`
	Leads        []LeadData        `json:"leads,omitempty"`

	// Relationships
	ProductTags         []ProductTagData         `json:"product_tags,omitempty"`
	NotificationConfigs []NotificationConfigData `json:"notification_configs,omitempty"`

	// Configuration
	Settings              SettingsData               `json:"settings,omitempty"`
	NotificationTemplates []NotificationTemplateData `json:"notification_templates,omitempty"`
	ThemeCustomization    ThemeCustomizationData     `json:"theme_customization,omitempty"`

	// Staff Sales Records (importação CSV)
	StaffSalesRecordsCSV []StaffSalesRecordCSVData `json:"staff_sales_records_csv,omitempty"`
}

func LoadSeedDataFromFile(filePath string) (*SeedData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de seed: %w", err)
	}

	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("erro ao parsear JSON: %w", err)
	}

	return &seedData, nil
}

func (s *SeedData) ValidateSeedData() error {
	if s.Organization.Name == "" {
		return fmt.Errorf("organização deve ter um nome")
	}
	if len(s.Menus) == 0 {
		return fmt.Errorf("deve haver pelo menos um menu")
	}

	return nil
}
