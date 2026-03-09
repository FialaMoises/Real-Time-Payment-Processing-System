@echo off
REM Banco Fiala - Startup Script para Windows
REM Este script inicia todo o sistema automaticamente

echo ========================================
echo 🏦 Banco Fiala
echo Sistema de Pagamentos em Tempo Real
echo ========================================
echo.

REM Verificar se Docker está rodando
docker info >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker não está rodando.
    echo Por favor, inicie o Docker Desktop.
    pause
    exit /b 1
)

echo ✅ Docker está rodando
echo.

REM Parar containers antigos (se existirem)
echo 🧹 Limpando containers antigos...
docker-compose down >nul 2>&1

REM Iniciar infraestrutura
echo 🚀 Iniciando infraestrutura...
echo    - PostgreSQL (porta 5432)
echo    - Redis (porta 6379)
echo    - RabbitMQ (porta 5672)
echo    - Frontend (porta 3000)
echo.
docker-compose up -d

if errorlevel 1 (
    echo ❌ Erro ao iniciar containers Docker
    pause
    exit /b 1
)

echo.
echo ⏳ Aguardando serviços iniciarem (15 segundos)...
timeout /t 15 /nobreak >nul

REM Verificar status
echo.
echo 📊 Status dos containers:
docker-compose ps

echo.
echo ========================================
echo ✅ Infraestrutura pronta!
echo ========================================
echo.
echo 🌐 Serviços Disponíveis:
echo ========================================
echo   Frontend (Banco Fiala):  http://localhost:3000
echo   API Backend:             http://localhost:8080
echo   RabbitMQ Management:     http://localhost:15672
echo                            (user: admin, pass: admin)
echo.
echo ========================================
echo 🚀 Para iniciar a API, execute:
echo ========================================
echo   go run cmd/api-server/main.go
echo.
echo ========================================
echo 📝 Como testar:
echo ========================================
echo   1. Abra http://localhost:3000 no navegador
echo   2. Crie uma conta nova
echo   3. Faça login
echo   4. Deposite dinheiro
echo   5. Teste transferências!
echo.
echo ========================================
echo 🛑 Para parar tudo:
echo ========================================
echo   docker-compose down
echo.
echo Pressione qualquer tecla para abrir o frontend...
pause >nul

start http://localhost:3000

echo.
echo Boa sorte! 🚀
echo.
pause
