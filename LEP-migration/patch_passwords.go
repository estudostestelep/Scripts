package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PatchPasswordsService migra hashes de senha diretamente entre dois bancos PostgreSQL.
// Necessário porque bcrypt é unidirecional — a API nunca expõe o campo password.
type PatchPasswordsService struct {
	logger *Logger
}

// NewPatchPasswordsService cria um novo PatchPasswordsService
func NewPatchPasswordsService(logger *Logger) *PatchPasswordsService {
	return &PatchPasswordsService{logger: logger}
}

// Run executa a migração de senhas do banco de origem para o destino
func (s *PatchPasswordsService) Run(sourceDBURL, destDBURL string) error {
	s.logger.Section("PATCH DE SENHAS — CONEXÃO DIRETA COM POSTGRESQL")

	// Conectar ao banco de origem
	s.logger.Info("Conectando ao banco de origem...")
	srcDB, err := sql.Open("postgres", sourceDBURL)
	if err != nil {
		return fmt.Errorf("erro ao conectar na origem: %w", err)
	}
	defer srcDB.Close()

	if err := srcDB.Ping(); err != nil {
		return fmt.Errorf("ping na origem falhou: %w", err)
	}
	s.logger.Success("Conectado ao banco de origem")

	// Conectar ao banco de destino
	s.logger.Info("Conectando ao banco de destino...")
	dstDB, err := sql.Open("postgres", destDBURL)
	if err != nil {
		return fmt.Errorf("erro ao conectar no destino: %w", err)
	}
	defer dstDB.Close()

	if err := dstDB.Ping(); err != nil {
		return fmt.Errorf("ping no destino falhou: %w", err)
	}
	s.logger.Success("Conectado ao banco de destino")

	// Migrar senhas de admins
	adminStats, err := s.patchAdmins(srcDB, dstDB)
	if err != nil {
		s.logger.Error("Falha ao migrar admins: %v", err)
	}

	// Migrar senhas de clients
	clientStats, err := s.patchClients(srcDB, dstDB)
	if err != nil {
		s.logger.Error("Falha ao migrar clients: %v", err)
	}

	s.logger.Section("RESULTADO DO PATCH DE SENHAS")
	s.logger.Info("Admins  — atualizados: %d, não encontrados: %d, erros: %d",
		adminStats.updated, adminStats.notFound, adminStats.failed)
	s.logger.Info("Clients — atualizados: %d, não encontrados: %d, erros: %d",
		clientStats.updated, clientStats.notFound, clientStats.failed)

	total := adminStats.updated + clientStats.updated
	if total > 0 {
		s.logger.Success("Patch concluído: %d hashes de senha migrados", total)
	} else {
		s.logger.Warn("Nenhuma senha foi migrada — verifique se as organizações foram restauradas antes de executar o patch")
	}

	return nil
}

type patchStats struct {
	updated  int
	notFound int
	failed   int
}

// patchAdmins migra senhas da tabela admins (match por email, que é único globalmente)
func (s *PatchPasswordsService) patchAdmins(srcDB, dstDB *sql.DB) (patchStats, error) {
	s.logger.Subsection("Admins")

	rows, err := srcDB.Query(`
		SELECT email, password
		FROM admins
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		return patchStats{}, fmt.Errorf("erro ao ler admins da origem: %w", err)
	}
	defer rows.Close()

	var stats patchStats
	for rows.Next() {
		var email, passwordHash string
		if err := rows.Scan(&email, &passwordHash); err != nil {
			s.logger.Error("  Erro ao ler linha de admin: %v", err)
			stats.failed++
			continue
		}

		result, err := dstDB.Exec(`
			UPDATE admins SET password = $1
			WHERE email = $2 AND deleted_at IS NULL
		`, passwordHash, email)
		if err != nil {
			s.logger.Error("  Admin '%s': erro ao atualizar: %v", email, err)
			stats.failed++
			continue
		}

		rows2, err := result.RowsAffected()
		if err != nil || rows2 == 0 {
			s.logger.Debug("  Admin '%s': não encontrado no destino", email)
			stats.notFound++
			continue
		}

		s.logger.Success("  Admin '%s': senha migrada", email)
		stats.updated++
	}

	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("erro ao iterar admins: %w", err)
	}

	return stats, nil
}

// patchClients migra senhas da tabela clients.
// Match por email + org_email — necessário porque org_id é diferente entre ambientes.
func (s *PatchPasswordsService) patchClients(srcDB, dstDB *sql.DB) (patchStats, error) {
	s.logger.Subsection("Clients")

	// Busca na origem: email do client + email da organização (chave de correlação entre ambientes)
	rows, err := srcDB.Query(`
		SELECT c.email, c.password, o.email AS org_email
		FROM clients c
		JOIN organizations o ON c.org_id = o.id
		WHERE c.deleted_at IS NULL
	`)
	if err != nil {
		return patchStats{}, fmt.Errorf("erro ao ler clients da origem: %w", err)
	}
	defer rows.Close()

	var stats patchStats
	for rows.Next() {
		var clientEmail, passwordHash, orgEmail string
		if err := rows.Scan(&clientEmail, &passwordHash, &orgEmail); err != nil {
			s.logger.Error("  Erro ao ler linha de client: %v", err)
			stats.failed++
			continue
		}

		// No destino: atualiza pelo email do client + email da organização (via join)
		result, err := dstDB.Exec(`
			UPDATE clients c
			SET password = $1
			FROM organizations o
			WHERE c.org_id = o.id
			  AND c.email = $2
			  AND o.email = $3
			  AND c.deleted_at IS NULL
		`, passwordHash, clientEmail, orgEmail)
		if err != nil {
			s.logger.Error("  Client '%s' (org: %s): erro ao atualizar: %v", clientEmail, orgEmail, err)
			stats.failed++
			continue
		}

		rows2, err := result.RowsAffected()
		if err != nil || rows2 == 0 {
			s.logger.Debug("  Client '%s' (org: %s): não encontrado no destino", clientEmail, orgEmail)
			stats.notFound++
			continue
		}

		s.logger.Success("  Client '%s' (org: %s): senha migrada", clientEmail, orgEmail)
		stats.updated++
	}

	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("erro ao iterar clients: %w", err)
	}

	return stats, nil
}
