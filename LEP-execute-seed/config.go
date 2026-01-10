package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config representa a configuração do seeder
type Config struct {
	Server struct {
		URL     string `yaml:"url"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"server"`

	Auth struct {
		OrganizationName string `yaml:"organization_name"`
		FallbackEmail    string `yaml:"fallback_email"`
		FallbackPassword string `yaml:"fallback_password"`
		AutoEmail        bool   `yaml:"auto_email"`
	} `yaml:"auth"`

	Seed struct {
		File        string `yaml:"file"`
		StopOnError bool   `yaml:"stop_on_error"`
		Parallel    bool   `yaml:"parallel"`
	} `yaml:"seed"`

	Logging struct {
		Level        string `yaml:"level"`
		ShowPayloads bool   `yaml:"show_payloads"`
	} `yaml:"logging"`
}

// LoadConfig carrega configuração de arquivo YAML + flags de linha de comando
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: struct {
			URL     string `yaml:"url"`
			Timeout int    `yaml:"timeout"`
		}{
			// Defaults
			URL: "http://localhost:8080",
			//URL:     "https://lep-system-516622888070.us-central1.run.app",
			Timeout: 30,
		},
		Auth: struct {
			OrganizationName string `yaml:"organization_name"`
			FallbackEmail    string `yaml:"fallback_email"`
			FallbackPassword string `yaml:"fallback_password"`
			AutoEmail        bool   `yaml:"auto_email"`
		}{
			FallbackEmail:    "pablo@lep.com",
			FallbackPassword: "senha123",
			AutoEmail:        true,
		},
		Seed: struct {
			File        string `yaml:"file"`
			StopOnError bool   `yaml:"stop_on_error"`
			Parallel    bool   `yaml:"parallel"`
		}{
			File:        "seed-fattoria.json",
			StopOnError: false,
			Parallel:    false,
		},
		Logging: struct {
			Level        string `yaml:"level"`
			ShowPayloads bool   `yaml:"show_payloads"`
		}{
			Level:        "info",
			ShowPayloads: false,
		},
	}

	// 1. Tentar carregar config.yaml se existir
	if _, err := os.Stat("config.yaml"); err == nil {
		fmt.Println("[ℹ] Carregando config.yaml...")
		data, err := os.ReadFile("config.yaml")
		if err != nil {
			return nil, fmt.Errorf("erro ao ler config.yaml: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("erro ao parsear config.yaml: %w", err)
		}
		fmt.Println("[✓] config.yaml carregado com sucesso")
	}

	// 2. Sobrescrever com flags de linha de comando
	url := flag.String("url", config.Server.URL, "URL base da API LEP")
	file := flag.String("file", config.Seed.File, "Arquivo JSON com dados de seed")
	verbose := flag.Bool("verbose", false, "Ativar modo verbose")
	org := flag.String("org", config.Auth.OrganizationName, "Nome da organização")
	timeout := flag.Int("timeout", config.Server.Timeout, "Timeout em segundos")

	flag.Parse()

	// Aplicar flags se foram definidas
	config.Server.URL = *url
	config.Seed.File = *file
	config.Auth.OrganizationName = *org
	config.Server.Timeout = *timeout

	if *verbose {
		config.Logging.Level = "debug"
		config.Logging.ShowPayloads = true
	}

	return config, nil
}

// Print exibe a configuração carregada
func (c *Config) Print() {
	fmt.Println("\n========== CONFIGURAÇÃO ==========")
	fmt.Printf("[ℹ] URL Backend: %s\n", c.Server.URL)
	fmt.Printf("[ℹ] Arquivo: %s\n", c.Seed.File)
	fmt.Printf("[ℹ] Organização: %s\n", c.Auth.OrganizationName)
	fmt.Printf("[ℹ] Log Level: %s\n", c.Logging.Level)
	if c.Logging.ShowPayloads {
		fmt.Printf("[ℹ] Mostrando payloads: SIM\n")
	}
	fmt.Println("==================================\n")
}

// GetEmailSlug retorna o slug da organização para criar email
func (c *Config) GetEmailSlug() string {
	slug := strings.ToLower(c.Auth.OrganizationName)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}

// GetAutoEmail gera email automático baseado no nome da organização
func (c *Config) GetAutoEmail() string {
	return fmt.Sprintf("%s@lep.com", c.GetEmailSlug())
}
