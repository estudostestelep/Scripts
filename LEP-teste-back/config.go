package main

// TestUser representa credenciais para testes
type TestUser struct {
	Email    string
	Password string
}

// Headers representa headers multi-tenant
type Headers struct {
	OrgID  string
	ProjID string
}

// Config representa configuração de testes
type Config struct {
	// Backend URL
	BackendURL string

	// Test User (for login)
	TestUser TestUser

	// Organization Name
	TestOrgName string

	// Project Name
	TestProjName string

	// Multi-tenant headers (will be populated after login)
	Headers Headers

	// Verbose logging
	Verbose bool
}

// GetDefaultConfig retorna configuração padrão
func GetDefaultConfig() Config {
	return Config{
		// URL do backend - ALTERAR AQUI conforme necessário
		// Local: "http://localhost:8080"
		// Online: "https://lep-system-516622888070.us-central1.run.app"
		BackendURL: "http://localhost:8080",
		//BackendURL: "https://lep-system-516622888070.us-central1.run.app",

		TestUser: TestUser{
			Email:    "pablo@lep.com",
			Password: "senha123",
		},
		TestOrgName:  "Test Organization",
		TestProjName: "Test Project",
		Headers: Headers{
			OrgID:  "",
			ProjID: "",
		},
		Verbose: false,
	}
}
