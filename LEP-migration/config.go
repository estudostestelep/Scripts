package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config representa a configuração completa do script de migração
type Config struct {
	Mode string `yaml:"mode"` // "dump" | "restore" | "patch-passwords"

	Source SourceDestConfig `yaml:"source"`
	Dest   SourceDestConfig `yaml:"destination"`

	Dump DumpConfig `yaml:"dump"`

	// Configuração de banco de dados para o modo patch-passwords
	SourceDB DBConfig `yaml:"source_db"`
	DestDB   DBConfig `yaml:"dest_db"`

	Server struct {
		Timeout int `yaml:"timeout"`
	} `yaml:"server"`

	Logging struct {
		Level        string `yaml:"level"`
		ShowPayloads bool   `yaml:"show_payloads"`
	} `yaml:"logging"`
}

// DBConfig contém a connection string de um banco PostgreSQL
type DBConfig struct {
	URL string `yaml:"url"` // ex: postgres://user:pass@host:5432/dbname?sslmode=disable
}

// SourceDestConfig contém credenciais e URL de uma instância do backend
type SourceDestConfig struct {
	URL      string `yaml:"url"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

// DumpConfig contém configurações de arquivos de dump
type DumpConfig struct {
	OutputFile string `yaml:"output_file"` // gerado automaticamente se vazio
	InputFile  string `yaml:"input_file"`  // arquivo a restaurar
}

// LoadConfig carrega configuração na ordem: defaults → config.yaml → flags CLI
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Defaults
	config.Mode = "dump"
	config.Source.URL = "http://localhost:8080"
	config.Dest.URL = "http://localhost:8080"
	config.Server.Timeout = 60
	config.Logging.Level = "info"
	config.Logging.ShowPayloads = false

	// Carregar config.yaml se existir
	if _, err := os.Stat("config.yaml"); err == nil {
		fmt.Println("[ℹ] Carregando config.yaml...")
		data, err := os.ReadFile("config.yaml")
		if err != nil {
			return nil, fmt.Errorf("erro ao ler config.yaml: %w", err)
		}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("erro ao parsear config.yaml: %w", err)
		}
		fmt.Println("[✓] config.yaml carregado")
	}

	// Flags CLI (sobrescrevem config.yaml)
	mode := flag.String("mode", config.Mode, "Modo de operação: dump | restore")

	sourceURL := flag.String("source-url", config.Source.URL, "URL do backend de origem (modo dump)")
	sourceEmail := flag.String("source-email", config.Source.Email, "Email para login na origem")
	sourcePassword := flag.String("source-password", config.Source.Password, "Senha para login na origem")

	destURL := flag.String("dest-url", config.Dest.URL, "URL do backend de destino (modo restore)")
	destEmail := flag.String("dest-email", config.Dest.Email, "Email para login no destino")
	destPassword := flag.String("dest-password", config.Dest.Password, "Senha para login no destino")

	dumpFile := flag.String("file", config.Dump.InputFile, "Arquivo de dump para restore (modo restore)")
	outputFile := flag.String("output", config.Dump.OutputFile, "Arquivo de saída para dump (modo dump, padrão: gerado automaticamente)")

	sourceDB := flag.String("source-db", config.SourceDB.URL, "Connection string do banco de origem (modo patch-passwords)")
	destDB := flag.String("dest-db", config.DestDB.URL, "Connection string do banco de destino (modo patch-passwords)")

	timeout := flag.Int("timeout", config.Server.Timeout, "Timeout HTTP em segundos")
	verbose := flag.Bool("verbose", false, "Ativar modo verbose (exibe payloads)")

	flag.Parse()

	config.Mode = *mode

	if *sourceURL != "" {
		config.Source.URL = *sourceURL
	}
	if *sourceEmail != "" {
		config.Source.Email = *sourceEmail
	}
	if *sourcePassword != "" {
		config.Source.Password = *sourcePassword
	}
	if *destURL != "" {
		config.Dest.URL = *destURL
	}
	if *destEmail != "" {
		config.Dest.Email = *destEmail
	}
	if *destPassword != "" {
		config.Dest.Password = *destPassword
	}
	if *dumpFile != "" {
		config.Dump.InputFile = *dumpFile
	}
	if *outputFile != "" {
		config.Dump.OutputFile = *outputFile
	}
	if *sourceDB != "" {
		config.SourceDB.URL = *sourceDB
	}
	if *destDB != "" {
		config.DestDB.URL = *destDB
	}
	config.Server.Timeout = *timeout

	if *verbose {
		config.Logging.Level = "debug"
		config.Logging.ShowPayloads = true
	}

	// Gerar nome de arquivo de saída automático se não informado
	if config.Dump.OutputFile == "" {
		ts := time.Now().Format("20060102-150405")
		config.Dump.OutputFile = fmt.Sprintf("migration-dump-%s.json", ts)
	}

	return config, nil
}

// Validate valida a configuração para o modo selecionado
func (c *Config) Validate() error {
	switch c.Mode {
	case "dump":
		if c.Source.URL == "" {
			return fmt.Errorf("source URL é obrigatório para modo dump")
		}
		if c.Source.Email == "" || c.Source.Password == "" {
			return fmt.Errorf("email e senha da origem são obrigatórios para modo dump")
		}
	case "restore":
		if c.Dest.URL == "" {
			return fmt.Errorf("destination URL é obrigatório para modo restore")
		}
		if c.Dest.Email == "" || c.Dest.Password == "" {
			return fmt.Errorf("email e senha do destino são obrigatórios para modo restore")
		}
		if c.Dump.InputFile == "" {
			return fmt.Errorf("arquivo de dump (-file) é obrigatório para modo restore")
		}
	case "patch-passwords":
		if c.SourceDB.URL == "" {
			return fmt.Errorf("connection string do banco de origem (-source-db) é obrigatória para modo patch-passwords")
		}
		if c.DestDB.URL == "" {
			return fmt.Errorf("connection string do banco de destino (-dest-db) é obrigatória para modo patch-passwords")
		}
	default:
		return fmt.Errorf("modo inválido '%s': use 'dump', 'restore' ou 'patch-passwords'", c.Mode)
	}
	return nil
}

// Print exibe a configuração ativa
func (c *Config) Print() {
	fmt.Println("\n========== CONFIGURAÇÃO ==========")
	fmt.Printf("[ℹ] Modo: %s\n", c.Mode)
	switch c.Mode {
	case "dump":
		fmt.Printf("[ℹ] Source URL: %s\n", c.Source.URL)
		fmt.Printf("[ℹ] Source Email: %s\n", c.Source.Email)
		fmt.Printf("[ℹ] Output: %s\n", c.Dump.OutputFile)
	case "restore":
		fmt.Printf("[ℹ] Dest URL: %s\n", c.Dest.URL)
		fmt.Printf("[ℹ] Dest Email: %s\n", c.Dest.Email)
		fmt.Printf("[ℹ] Input File: %s\n", c.Dump.InputFile)
	case "patch-passwords":
		fmt.Printf("[ℹ] Source DB: %s\n", maskDBURL(c.SourceDB.URL))
		fmt.Printf("[ℹ] Dest DB: %s\n", maskDBURL(c.DestDB.URL))
	}
	fmt.Printf("[ℹ] Timeout: %ds\n", c.Server.Timeout)
	fmt.Printf("[ℹ] Log Level: %s\n", c.Logging.Level)
	fmt.Println("==================================\n")
}

// maskDBURL oculta a senha na connection string para exibição segura nos logs.
// Ex: postgres://user:senha@host:5432/db → postgres://user:***@host:5432/db
func maskDBURL(url string) string {
	// Formato: postgres://user:password@host/db
	atIdx := strings.Index(url, "@")
	if atIdx < 0 {
		return url
	}
	schemeEnd := strings.Index(url, "://")
	if schemeEnd < 0 {
		return url
	}
	credPart := url[schemeEnd+3 : atIdx]
	colonIdx := strings.Index(credPart, ":")
	if colonIdx < 0 {
		return url
	}
	user := credPart[:colonIdx]
	return url[:schemeEnd+3] + user + ":***" + url[atIdx:]
}
