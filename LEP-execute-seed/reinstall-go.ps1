# Script para reinstalar Go 1.24.11
# Execute como Administrador: PowerShell -> Run as Administrator

Write-Host "=== Reinstalacao do Go ===" -ForegroundColor Cyan

# Passo 1: Remover instalacao antiga
Write-Host ""
Write-Host "[1/4] Removendo instalacao corrupta..." -ForegroundColor Yellow
if (Test-Path "C:\Go") {
    Remove-Item -Recurse -Force "C:\Go" -ErrorAction SilentlyContinue
    Write-Host "OK C:\Go removido" -ForegroundColor Green
} else {
    Write-Host "OK C:\Go ja estava removido" -ForegroundColor Green
}

# Passo 2: Baixar Go 1.24.11
Write-Host ""
Write-Host "[2/4] Baixando Go 1.24.11..." -ForegroundColor Yellow
$goVersion = "1.24.11"
$goInstaller = "go$goVersion.windows-amd64.msi"
$downloadUrl = "https://go.dev/dl/$goInstaller"
$installerPath = "$env:TEMP\$goInstaller"

try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $installerPath -UseBasicParsing
    Write-Host "OK Download concluido: $installerPath" -ForegroundColor Green
} catch {
    Write-Host "ERRO ao baixar: $_" -ForegroundColor Red
    exit 1
}

# Passo 3: Instalar Go
Write-Host ""
Write-Host "[3/4] Instalando Go 1.24.11..." -ForegroundColor Yellow
try {
    Start-Process msiexec.exe -ArgumentList "/i `"$installerPath`" /quiet /norestart" -Wait -NoNewWindow
    Write-Host "OK Instalacao concluida" -ForegroundColor Green
} catch {
    Write-Host "ERRO na instalacao: $_" -ForegroundColor Red
    exit 1
}

# Passo 4: Verificar instalacao
Write-Host ""
Write-Host "[4/4] Verificando instalacao..." -ForegroundColor Yellow
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

Start-Sleep -Seconds 2

try {
    $goVersionOutput = & "C:\Go\bin\go.exe" version 2>&1
    Write-Host "OK Go instalado com sucesso!" -ForegroundColor Green
    Write-Host "  Versao: $goVersionOutput" -ForegroundColor Cyan
} catch {
    Write-Host "ERRO ao verificar: $_" -ForegroundColor Red
    Write-Host "  Tente fechar e reabrir o terminal" -ForegroundColor Yellow
}

# Limpeza
Remove-Item $installerPath -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "=== Instalacao concluida ===" -ForegroundColor Cyan
Write-Host "Feche este terminal e abra um novo para usar o Go" -ForegroundColor Yellow
