# Script para executar o seed apos instalar Go
# Execute em um terminal normal (nao precisa de admin)

Write-Host "=== Execucao do Seed LEP ===" -ForegroundColor Cyan

# Navegar para o diretorio
Set-Location "C:\Users\pablo\OneDrive\Área de Trabalho\Trabalho Be Growth\Projetos\LEP\LEP-Script\LEP-execute-seed"

# Passo 1: Verificar Go
Write-Host "`n[1/3] Verificando instalacao do Go..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✓ $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go nao encontrado! Execute reinstall-go.ps1 primeiro" -ForegroundColor Red
    exit 1
}

# Passo 2: go mod tidy
Write-Host "`n[2/3] Atualizando dependencias (go mod tidy)..." -ForegroundColor Yellow
try {
    go mod tidy
    Write-Host "✓ Dependencias atualizadas" -ForegroundColor Green
} catch {
    Write-Host "✗ Erro ao executar go mod tidy: $_" -ForegroundColor Red
    exit 1
}

# Passo 3: Executar seed
Write-Host "`n[3/3] Executando seed..." -ForegroundColor Yellow
Write-Host "Aguarde, isso pode levar alguns minutos..." -ForegroundColor Cyan
Write-Host ""

try {
    go run main.go
    Write-Host "`n✓ Seed executado!" -ForegroundColor Green
} catch {
    Write-Host "`n✗ Erro ao executar seed: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`n=== Execucao concluida ===" -ForegroundColor Cyan
