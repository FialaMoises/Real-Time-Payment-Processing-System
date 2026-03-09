// Configuration
const API_URL = 'http://localhost:8080/api/v1';

// State
let currentUser = null;
let currentAccount = null;
let authToken = null;

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    checkAuth();
    formatCPFInput();
});

// Check if user is already logged in (from localStorage)
function checkAuth() {
    const token = localStorage.getItem('authToken');
    const user = localStorage.getItem('currentUser');
    const account = localStorage.getItem('currentAccount');

    if (token && user && account) {
        authToken = token;
        currentUser = JSON.parse(user);
        currentAccount = JSON.parse(account);
        showDashboard();
    }
}

// Tab switching
function showTab(tabName) {
    // Update tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');

    // Update tab content
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    document.getElementById(`${tabName}-form`).classList.add('active');
}

// Alert messages
function showAlert(message, type = 'info') {
    const alert = document.getElementById('alert');
    alert.textContent = message;
    alert.className = `alert ${type}`;
    alert.style.display = 'block';

    // Auto-hide after 5 seconds
    setTimeout(() => {
        alert.style.display = 'none';
    }, 5000);
}

// Format CPF input
function formatCPFInput() {
    const cpfInput = document.getElementById('reg-cpf');
    if (cpfInput) {
        cpfInput.addEventListener('input', (e) => {
            let value = e.target.value.replace(/\D/g, '');
            if (value.length > 11) value = value.slice(0, 11);

            if (value.length > 9) {
                value = value.replace(/(\d{3})(\d{3})(\d{3})(\d{2})/, '$1.$2.$3-$4');
            } else if (value.length > 6) {
                value = value.replace(/(\d{3})(\d{3})(\d{1,3})/, '$1.$2.$3');
            } else if (value.length > 3) {
                value = value.replace(/(\d{3})(\d{1,3})/, '$1.$2');
            }

            e.target.value = value;
        });
    }
}

// Register
async function register(event) {
    event.preventDefault();

    const data = {
        email: document.getElementById('reg-email').value,
        password: document.getElementById('reg-password').value,
        full_name: document.getElementById('reg-name').value,
        cpf: document.getElementById('reg-cpf').value,
        phone: document.getElementById('reg-phone').value || undefined
    };

    try {
        const response = await fetch(`${API_URL}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            showAlert(`Conta criada com sucesso! Número da conta: ${result.account_number}`, 'success');

            // Auto-login after registration
            setTimeout(() => {
                document.getElementById('login-email').value = data.email;
                document.getElementById('login-password').value = data.password;
                showTab('login');
                document.querySelectorAll('.tab-btn')[0].click();
            }, 2000);
        } else {
            showAlert(result.error || 'Erro ao criar conta', 'error');
        }
    } catch (error) {
        showAlert('Erro ao conectar com o servidor', 'error');
        console.error(error);
    }
}

// Login
async function login(event) {
    event.preventDefault();

    const data = {
        email: document.getElementById('login-email').value,
        password: document.getElementById('login-password').value
    };

    try {
        const response = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            authToken = result.access_token;
            currentUser = result.user;

            // Save to localStorage
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('currentUser', JSON.stringify(currentUser));

            // Get account info
            await getAccountInfo();

            showAlert(`Bem-vindo, ${currentUser.full_name}!`, 'success');
            setTimeout(showDashboard, 1000);
        } else {
            showAlert(result.error || 'Email ou senha incorretos', 'error');
        }
    } catch (error) {
        showAlert('Erro ao conectar com o servidor', 'error');
        console.error(error);
    }
}

// Get account info
async function getAccountInfo() {
    try {
        const response = await fetch(`${API_URL}/accounts/me`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });

        if (response.ok) {
            currentAccount = await response.json();
            localStorage.setItem('currentAccount', JSON.stringify(currentAccount));
            updateAccountDisplay();
        }
    } catch (error) {
        console.error('Error fetching account:', error);
    }
}

// Show dashboard
function showDashboard() {
    document.getElementById('auth-section').style.display = 'none';
    document.getElementById('dashboard-section').style.display = 'block';
    document.getElementById('user-info').style.display = 'flex';
    document.getElementById('user-name').textContent = currentUser.full_name;

    updateAccountDisplay();
    refreshBalance();
}

// Update account display
function updateAccountDisplay() {
    if (currentAccount) {
        document.getElementById('account-number').textContent = currentAccount.account_number;
        document.getElementById('balance').textContent = formatCurrency(currentAccount.balance);
    }
}

// Refresh balance
async function refreshBalance() {
    if (!currentAccount) return;

    try {
        const response = await fetch(`${API_URL}/accounts/${currentAccount.id}/balance`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });

        if (response.ok) {
            const data = await response.json();
            currentAccount.balance = data.balance;
            localStorage.setItem('currentAccount', JSON.stringify(currentAccount));
            updateAccountDisplay();
        }
    } catch (error) {
        console.error('Error refreshing balance:', error);
    }
}

// Logout
function logout() {
    localStorage.removeItem('authToken');
    localStorage.removeItem('currentUser');
    localStorage.removeItem('currentAccount');

    authToken = null;
    currentUser = null;
    currentAccount = null;

    document.getElementById('auth-section').style.display = 'block';
    document.getElementById('dashboard-section').style.display = 'none';
    document.getElementById('user-info').style.display = 'none';

    hideAllSections();
    showAlert('Logout realizado com sucesso', 'info');
}

// Show section
function showSection(section) {
    hideAllSections();
    document.getElementById(`${section}-section`).style.display = 'block';

    if (section === 'history') {
        loadTransactions();
    } else if (section === 'ledger') {
        loadLedger();
    }
}

// Hide all sections
function hideAllSections() {
    document.querySelectorAll('.transaction-card, #history-section, #ledger-section').forEach(el => {
        el.style.display = 'none';
    });

    // Clear forms
    document.querySelectorAll('form').forEach(form => {
        if (form.id !== 'login-form' && form.id !== 'register-form') {
            form.reset();
        }
    });
}

// Deposit
async function deposit(event) {
    event.preventDefault();

    const amount = parseFloat(document.getElementById('deposit-amount').value);
    const description = document.getElementById('deposit-description').value || 'Depósito';

    const data = {
        idempotency_key: `deposit-${Date.now()}-${Math.random()}`,
        account_id: currentAccount.id,
        amount: amount,
        description: description
    };

    try {
        const response = await fetch(`${API_URL}/transactions/deposit`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            showAlert(`Depósito de R$ ${formatCurrency(amount)} realizado com sucesso!`, 'success');
            hideAllSections();
            await refreshBalance();
        } else {
            showAlert(result.error || 'Erro ao realizar depósito', 'error');
        }
    } catch (error) {
        showAlert('Erro ao conectar com o servidor', 'error');
        console.error(error);
    }
}

// Withdrawal
async function withdrawal(event) {
    event.preventDefault();

    const amount = parseFloat(document.getElementById('withdrawal-amount').value);
    const description = document.getElementById('withdrawal-description').value || 'Saque';

    const data = {
        idempotency_key: `withdrawal-${Date.now()}-${Math.random()}`,
        account_id: currentAccount.id,
        amount: amount,
        description: description
    };

    try {
        const response = await fetch(`${API_URL}/transactions/withdrawal`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            showAlert(`Saque de R$ ${formatCurrency(amount)} realizado com sucesso!`, 'success');
            hideAllSections();
            await refreshBalance();
        } else {
            showAlert(result.error || 'Erro ao realizar saque', 'error');
        }
    } catch (error) {
        showAlert('Erro ao conectar com o servidor', 'error');
        console.error(error);
    }
}

// Transfer
async function transfer(event) {
    event.preventDefault();

    const toAccount = document.getElementById('transfer-to-account').value;
    const amount = parseFloat(document.getElementById('transfer-amount').value);
    const description = document.getElementById('transfer-description').value || 'Transferência';

    const data = {
        idempotency_key: `transfer-${Date.now()}-${Math.random()}`,
        from_account_id: currentAccount.id,
        to_account_id: toAccount,
        amount: amount,
        description: description
    };

    try {
        const response = await fetch(`${API_URL}/transactions/transfer`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            showAlert(`Transferência de R$ ${formatCurrency(amount)} realizada com sucesso!`, 'success');
            hideAllSections();
            await refreshBalance();
        } else {
            showAlert(result.error || 'Erro ao realizar transferência', 'error');
        }
    } catch (error) {
        showAlert('Erro ao conectar com o servidor', 'error');
        console.error(error);
    }
}

// Load transactions
async function loadTransactions() {
    const container = document.getElementById('transactions-list');
    container.innerHTML = '<p class="loading">Carregando...</p>';

    try {
        const response = await fetch(`${API_URL}/transactions?account_id=${currentAccount.id}&limit=50`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });

        if (response.ok) {
            const data = await response.json();
            const transactions = data.transactions || [];

            if (transactions.length === 0) {
                container.innerHTML = '<p class="loading">Nenhuma transação encontrada</p>';
                return;
            }

            container.innerHTML = transactions.map(tx => {
                const isCredit = tx.to_account_id === currentAccount.id;
                const typeLabel = getTransactionTypeLabel(tx.type);
                const amountClass = isCredit ? 'credit' : 'debit';
                const sign = isCredit ? '+' : '-';

                return `
                    <div class="transaction-item">
                        <div class="transaction-info">
                            <div class="transaction-type">${typeLabel}</div>
                            <div class="transaction-description">${tx.description || '-'}</div>
                            <div class="transaction-date">${formatDate(tx.created_at)}</div>
                        </div>
                        <div>
                            <div class="transaction-amount ${amountClass}">${sign} R$ ${formatCurrency(tx.amount)}</div>
                            <div class="transaction-status status-${tx.status.toLowerCase()}">${tx.status}</div>
                        </div>
                    </div>
                `;
            }).join('');
        } else {
            container.innerHTML = '<p class="loading">Erro ao carregar transações</p>';
        }
    } catch (error) {
        container.innerHTML = '<p class="loading">Erro ao conectar com o servidor</p>';
        console.error(error);
    }
}

// Load ledger
async function loadLedger() {
    const container = document.getElementById('ledger-list');
    container.innerHTML = '<p class="loading">Carregando...</p>';

    try {
        const response = await fetch(`${API_URL}/ledger/${currentAccount.id}?limit=50`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });

        if (response.ok) {
            const data = await response.json();
            const entries = data.entries || [];

            if (entries.length === 0) {
                container.innerHTML = '<p class="loading">Nenhuma entrada no ledger</p>';
                return;
            }

            container.innerHTML = entries.map(entry => {
                const isCredit = entry.operation_type === 'CREDIT';
                const amountClass = isCredit ? 'credit' : 'debit';
                const sign = isCredit ? '+' : '-';

                return `
                    <div class="ledger-item">
                        <div class="ledger-operation">
                            <span>${entry.operation_type}</span>
                            <span class="transaction-status status-completed">ID: ${entry.id}</span>
                        </div>
                        <div class="ledger-details">
                            <div><strong>Transaction:</strong> ${entry.transaction_id.substring(0, 8)}...</div>
                            <div><strong>Data:</strong> ${formatDate(entry.created_at)}</div>
                        </div>
                        <div class="ledger-amount ${amountClass}">
                            ${sign} R$ ${formatCurrency(Math.abs(entry.amount))}
                        </div>
                        <div style="font-size: 14px; color: var(--text-light); margin-top: 8px;">
                            Saldo após: R$ ${formatCurrency(entry.balance_after)}
                        </div>
                    </div>
                `;
            }).join('');
        } else {
            container.innerHTML = '<p class="loading">Erro ao carregar ledger</p>';
        }
    } catch (error) {
        container.innerHTML = '<p class="loading">Erro ao conectar com o servidor</p>';
        console.error(error);
    }
}

// Helper functions
function formatCurrency(value) {
    return parseFloat(value).toFixed(2).replace('.', ',');
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('pt-BR', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function getTransactionTypeLabel(type) {
    const labels = {
        'DEPOSIT': '💰 Depósito',
        'WITHDRAWAL': '💸 Saque',
        'TRANSFER': '🔄 Transferência'
    };
    return labels[type] || type;
}
