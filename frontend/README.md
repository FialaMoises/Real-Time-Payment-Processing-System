# 🏦 Banco Fiala - Frontend

Interface web para o Sistema de Pagamentos em Tempo Real.

## 🎨 Features

- ✅ **Autenticação** - Login e registro de usuários
- ✅ **Dashboard** - Visão geral da conta e saldo
- ✅ **Transações** - Depósito, saque e transferências
- ✅ **Histórico** - Lista de todas as transações
- ✅ **Ledger** - Histórico imutável (auditoria)
- ✅ **Design Responsivo** - Funciona em desktop e mobile

## 🚀 Como Rodar

### Com Docker (Recomendado)

O frontend é servido automaticamente via Docker Compose:

```bash
# Na raiz do projeto
docker-compose up -d
```

Acesse: `http://localhost:3000`

### Sem Docker (Servidor local)

Use qualquer servidor HTTP estático:

```bash
# Python
python -m http.server 3000

# Node.js (http-server)
npx http-server -p 3000

# PHP
php -S localhost:3000
```

## 🎯 Como Usar

### 1. Criar Conta

![Registro](https://via.placeholder.com/800x400/667eea/ffffff?text=Tela+de+Registro)

- Preencha nome, email, CPF e senha
- Clique em "Criar Conta"
- Um número de conta será gerado automaticamente

### 2. Login

![Login](https://via.placeholder.com/800x400/667eea/ffffff?text=Tela+de+Login)

- Use seu email e senha
- O token JWT é salvo no localStorage

### 3. Dashboard

![Dashboard](https://via.placeholder.com/800x400/667eea/ffffff?text=Dashboard)

- Veja seu saldo em tempo real
- Acesse as funcionalidades através dos botões

### 4. Depositar

![Depositar](https://via.placeholder.com/800x400/667eea/ffffff?text=Depositar)

- Digite o valor
- Adicione uma descrição (opcional)
- Confirme o depósito

### 5. Transferir

![Transferir](https://via.placeholder.com/800x400/667eea/ffffff?text=Transferir)

- Cole o Account ID da conta destino
- Digite o valor
- Confirme a transferência

### 6. Histórico

![Histórico](https://via.placeholder.com/800x400/667eea/ffffff?text=Histórico)

- Veja todas as suas transações
- Status (Completed, Pending, Failed)
- Data e hora de cada operação

### 7. Ledger (Auditoria)

![Ledger](https://via.placeholder.com/800x400/667eea/ffffff?text=Ledger)

- Histórico **imutável** de todas as operações
- Cada entrada mostra o saldo após a operação
- Não pode ser modificado ou deletado

## 🔧 Configuração

### API URL

Por padrão, o frontend conecta em `http://localhost:8080/api/v1`

Para mudar, edite `app.js`:

```javascript
const API_URL = 'http://seu-backend:8080/api/v1';
```

## 🎨 Design

### Cores

```css
--primary: #2563eb      /* Azul */
--success: #10b981      /* Verde */
--danger: #ef4444       /* Vermelho */
--warning: #f59e0b      /* Amarelo */
```

### Fonte

```css
font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
```

## 📱 Responsividade

O frontend é totalmente responsivo e funciona em:

- ✅ Desktop (1920x1080+)
- ✅ Laptop (1366x768+)
- ✅ Tablet (768x1024)
- ✅ Mobile (320x568+)

## 🔒 Segurança

- ✅ Senhas nunca são exibidas
- ✅ Token JWT armazenado no localStorage
- ✅ Requisições autenticadas via Bearer token
- ✅ CORS habilitado no backend

## 📊 Estrutura de Arquivos

```
frontend/
├── index.html      # Estrutura HTML
├── style.css       # Estilos CSS
├── app.js          # Lógica JavaScript
└── README.md       # Esta documentação
```

## 🧪 Testando

1. **Abrir DevTools** (F12)
2. **Console** - Ver logs de requisições
3. **Network** - Verificar chamadas à API
4. **Application** - Ver localStorage (token, user, account)

## 🐛 Debug

### Ver Token JWT

```javascript
console.log(localStorage.getItem('authToken'));
```

### Ver Usuário Atual

```javascript
console.log(localStorage.getItem('currentUser'));
```

### Ver Conta Atual

```javascript
console.log(localStorage.getItem('currentAccount'));
```

### Limpar Dados

```javascript
localStorage.clear();
location.reload();
```

## 🚀 Próximas Features

- [ ] Notificações em tempo real (WebSocket)
- [ ] Exportar extrato PDF
- [ ] Gráficos de gastos
- [ ] Dark mode
- [ ] Múltiplos idiomas
- [ ] Biometria (WebAuthn)

## 📝 Licença

MIT License - Use como quiser!

---

**Made with ❤️ for Banco Fiala**
