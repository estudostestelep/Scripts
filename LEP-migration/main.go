package main

import (
	"fmt"
	"os"
)

func main() {
	// 1. Carregar configuração
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[✗] Erro ao carregar configuração: %v\n", err)
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "[✗] Configuração inválida: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nUso:")
		fmt.Fprintln(os.Stderr, "  Dump:            go run . -mode dump -source-url <url> -source-email <email> -source-password <senha>")
		fmt.Fprintln(os.Stderr, "  Restore:         go run . -mode restore -dest-url <url> -dest-email <email> -dest-password <senha> -file <dump.json>")
		fmt.Fprintln(os.Stderr, "  Patch passwords: go run . -mode patch-passwords -source-db <postgres://...> -dest-db <postgres://...>")
		os.Exit(1)
	}

	isVerbose := config.Logging.Level == "debug" || config.Logging.Level == "verbose"
	logger := NewLogger(isVerbose)

	printHeader()
	config.Print()

	// 2. Executar modo selecionado
	switch config.Mode {
	case "dump":
		if err := runDump(config, logger); err != nil {
			logger.Error("Dump falhou: %v", err)
			os.Exit(1)
		}

	case "restore":
		if err := runRestore(config, logger); err != nil {
			logger.Error("Restore falhou: %v", err)
			os.Exit(1)
		}

	case "patch-passwords":
		if err := runPatchPasswords(config, logger); err != nil {
			logger.Error("Patch de senhas falhou: %v", err)
			os.Exit(1)
		}
	}
}

func runPatchPasswords(config *Config, logger *Logger) error {
	service := NewPatchPasswordsService(logger)
	return service.Run(config.SourceDB.URL, config.DestDB.URL)
}

func runDump(config *Config, logger *Logger) error {
	client := NewMigrationClient(
		config.Source.URL,
		config.Server.Timeout,
		logger,
		config.Logging.ShowPayloads,
	)

	service := NewDumpService(client, logger)
	return service.Run(
		config.Source.Email,
		config.Source.Password,
		config.Source.URL,
		config.Dump.OutputFile,
	)
}

func runRestore(config *Config, logger *Logger) error {
	// Carregar arquivo de dump
	logger.Info("Carregando arquivo de dump: %s", config.Dump.InputFile)
	dump, err := LoadDumpFile(config.Dump.InputFile)
	if err != nil {
		return err
	}
	logger.Success("Dump carregado: %d organizações (gerado em %s)", len(dump.Organizations), dump.Metadata.DumpedAt)
	logger.Info("Origem original: %s", dump.Metadata.SourceURL)

	client := NewMigrationClient(
		config.Dest.URL,
		config.Server.Timeout,
		logger,
		config.Logging.ShowPayloads,
	)

	service := NewRestoreService(client, logger, dump.Metadata.SourceURL, config.Server.Timeout, config.Dest.Email, config.Dest.Password)
	return service.Run(dump, config.Dest.Password)
}

func printHeader() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║          LEP Migration Tool v1.0.0               ║")
	fmt.Println("║   Migração de dados entre instâncias do LEP      ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
}
