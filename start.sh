#!/bin/bash

# Banco Fiala - Startup Script
# Este script inicia todo o sistema automaticamente

echo "🏦 Banco Fiala - Sistema de Pagamentos em Tempo Real"
echo "===================================================="
echo ""

# Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Por favor, inicie o Docker Desktop."
    exit 1
fi

echo "✅ Docker está rodando"
echo ""

# Parar containers antigos (se existirem)
echo "🧹 Limpando containers antigos..."
docker-compose down > /dev/null 2>&1

# Iniciar infraestrutura
echo "🚀 Iniciando infraestrutura (PostgreSQL, Redis, RabbitMQ, Frontend)..."
docker-compose up -d

if [ $? -ne 0 ]; then
    echo "❌ Erro ao iniciar containers Docker"
    exit 1
fi

echo ""
echo "⏳ Aguardando serviços iniciarem (15 segundos)..."
sleep 15

# Verificar se containers estão rodando
echo ""
echo "📊 Status dos containers:"
docker-compose ps

echo ""
echo "✅ Infraestrutura pronta!"
echo ""
echo "=========================================="
echo "🌐 Serviços Disponíveis:"
echo "=========================================="
echo "  Frontend (Banco Fiala):  http://localhost:3000"
echo "  API Backend:             http://localhost:8080"
echo "  RabbitMQ Management:     http://localhost:15672"
echo "                           (user: admin, pass: admin)"
echo ""
echo "=========================================="
echo "🚀 Para iniciar a API, execute:"
echo "=========================================="
echo "  go run cmd/api-server/main.go"
echo ""
echo "=========================================="
echo "📝 Para testar:"
echo "=========================================="
echo "  1. Abra http://localhost:3000 no navegador"
echo "  2. Crie uma conta nova"
echo "  3. Faça login"
echo "  4. Deposite dinheiro"
echo "  5. Teste transferências!"
echo ""
echo "=========================================="
echo "🛑 Para parar tudo:"
echo "=========================================="
echo "  docker-compose down"
echo ""
echo "Boa sorte! 🚀"
