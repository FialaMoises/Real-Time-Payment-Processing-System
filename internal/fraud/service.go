package fraud

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/pkg/logger"
)

type Service interface {
	// Check verifica uma transação contra todas as regras de fraude
	CheckTransaction(ctx context.Context, txCtx *TransactionContext) (*FraudCheckResult, error)

	// Alerts management
	GetAlertByTransactionID(ctx context.Context, txID uuid.UUID) (*FraudAlert, error)
	ListPendingAlerts(ctx context.Context, limit, offset int) ([]FraudAlert, error)
	ReviewAlert(ctx context.Context, alertID, reviewerID uuid.UUID, status AlertStatus) error

	// Scores
	GetAccountRiskHistory(ctx context.Context, accountID uuid.UUID, limit int) ([]FraudScore, error)
}

type service struct {
	repo  Repository
	rules []FraudRule
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CheckTransaction executa todas as verificações de fraude
func (s *service) CheckTransaction(ctx context.Context, txCtx *TransactionContext) (*FraudCheckResult, error) {
	// Carregar regras ativas (cache em produção)
	rules, err := s.repo.GetActiveRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load fraud rules: %w", err)
	}

	result := &FraudCheckResult{
		IsFraud:   false,
		RiskScore: 0,
		Reasons:   []string{},
		Action:    ActionFlag,
		Rules:     []FraudRule{},
	}

	// Calcular scores individuais
	score := &FraudScore{
		TransactionID: txCtx.TransactionID,
		AccountID:     txCtx.AccountID,
		Features:      make(map[string]interface{}),
	}

	// 1. Verificar amount score
	amountScore := s.calculateAmountScore(txCtx.Amount)
	score.AmountScore = amountScore
	score.Features["amount"] = txCtx.Amount

	// 2. Verificar velocity score
	velocityScore, err := s.calculateVelocityScore(ctx, txCtx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to calculate velocity score")
		velocityScore = 0
	}
	score.VelocityScore = velocityScore
	score.Features["recent_tx_count"] = len(txCtx.RecentTransactions)

	// 3. Verificar time score
	timeScore := s.calculateTimeScore(txCtx.Timestamp)
	score.TimeScore = timeScore
	score.Features["hour"] = txCtx.Timestamp.Hour()

	// 4. Pattern score (baseado em comportamento histórico)
	patternScore := s.calculatePatternScore(txCtx)
	score.PatternScore = patternScore

	// 5. ML score (placeholder para modelo futuro)
	score.MLScore = 0

	// Total score
	score.TotalScore = amountScore + velocityScore + timeScore + patternScore
	result.RiskScore = score.TotalScore

	// Salvar score para histórico/ML
	if err := s.repo.CreateScore(ctx, score); err != nil {
		logger.Error().Err(err).Msg("failed to save fraud score")
	}

	// Aplicar regras
	for _, rule := range rules {
		triggered, reason := s.evaluateRule(ctx, &rule, txCtx, score)
		if triggered {
			result.IsFraud = true
			result.Rules = append(result.Rules, rule)
			result.Reasons = append(result.Reasons, reason)

			// Aumentar score baseado na regra
			result.RiskScore += rule.RiskScore

			// Atualizar ação baseada na regra mais severa
			if rule.Action == ActionBlock {
				result.Action = ActionBlock
			} else if rule.Action == ActionReview && result.Action != ActionBlock {
				result.Action = ActionReview
			}
		}
	}

	// Criar alerta se fraud detectada
	if result.IsFraud {
		alert := &FraudAlert{
			TransactionID: txCtx.TransactionID,
			RiskScore:     result.RiskScore,
			Status:        AlertStatusPending,
			Reason:        fmt.Sprintf("%d rule(s) triggered", len(result.Rules)),
			Details: map[string]interface{}{
				"reasons":     result.Reasons,
				"rules":       extractRuleNames(result.Rules),
				"total_score": result.RiskScore,
			},
		}

		if err := s.repo.CreateAlert(ctx, alert); err != nil {
			logger.Error().Err(err).Msg("failed to create fraud alert")
		} else {
			result.AlertID = &alert.ID

			logger.Warn().
				Str("transaction_id", txCtx.TransactionID.String()).
				Int("risk_score", result.RiskScore).
				Str("action", string(result.Action)).
				Strs("reasons", result.Reasons).
				Msg("fraud detected")
		}
	}

	return result, nil
}

// calculateAmountScore calcula score baseado no valor da transação
func (s *service) calculateAmountScore(amount float64) int {
	switch {
	case amount >= 100000:
		return 90
	case amount >= 50000:
		return 70
	case amount >= 10000:
		return 40
	case amount >= 5000:
		return 20
	case amount >= 1000:
		return 10
	default:
		return 0
	}
}

// calculateVelocityScore calcula score baseado na velocidade de transações
func (s *service) calculateVelocityScore(ctx context.Context, txCtx *TransactionContext) (int, error) {
	now := txCtx.Timestamp

	// Verificar últimas 5 minutos
	count5min, err := s.repo.GetTransactionCount(ctx, txCtx.AccountID, now.Add(-5*time.Minute))
	if err != nil {
		return 0, err
	}

	// Verificar última hora
	count1h, err := s.repo.GetTransactionCount(ctx, txCtx.AccountID, now.Add(-1*time.Hour))
	if err != nil {
		return 0, err
	}

	score := 0

	// Mais de 5 transações em 5 minutos
	if count5min > 5 {
		score += 60
	} else if count5min > 3 {
		score += 30
	}

	// Mais de 10 transações em 1 hora
	if count1h > 10 {
		score += 50
	} else if count1h > 5 {
		score += 20
	}

	return score, nil
}

// calculateTimeScore calcula score baseado no horário da transação
func (s *service) calculateTimeScore(timestamp time.Time) int {
	hour := timestamp.Hour()

	// Horários suspeitos: 2h - 5h da manhã
	if hour >= 2 && hour <= 5 {
		return 30
	}

	// Horários ligeiramente suspeitos: 0h - 2h e 5h - 7h
	if (hour >= 0 && hour < 2) || (hour > 5 && hour < 7) {
		return 15
	}

	return 0
}

// calculatePatternScore analisa padrões de comportamento
func (s *service) calculatePatternScore(txCtx *TransactionContext) int {
	if len(txCtx.RecentTransactions) == 0 {
		return 0
	}

	score := 0

	// Verificar transações muito próximas (menos de 10 segundos)
	for _, recentTx := range txCtx.RecentTransactions {
		diff := txCtx.Timestamp.Sub(recentTx.Timestamp)
		if diff < 10*time.Second {
			score += 50
			break
		}
	}

	// Verificar valor muito acima da média
	var totalAmount float64
	for _, recentTx := range txCtx.RecentTransactions {
		totalAmount += recentTx.Amount
	}
	if len(txCtx.RecentTransactions) > 0 {
		avgAmount := totalAmount / float64(len(txCtx.RecentTransactions))
		if txCtx.Amount > avgAmount*5 { // 5x a média
			score += 40
		} else if txCtx.Amount > avgAmount*3 { // 3x a média
			score += 20
		}
	}

	return score
}

// evaluateRule avalia uma regra específica
func (s *service) evaluateRule(ctx context.Context, rule *FraudRule, txCtx *TransactionContext, score *FraudScore) (bool, string) {
	switch rule.RuleType {
	case RuleTypeHighAmount:
		threshold, ok := rule.Condition["threshold"].(float64)
		if !ok {
			return false, ""
		}
		if txCtx.Amount >= threshold {
			return true, fmt.Sprintf("Amount %.2f exceeds threshold %.2f", txCtx.Amount, threshold)
		}

	case RuleTypeVelocity:
		count, ok1 := rule.Condition["count"].(float64)
		windowMinutes, ok2 := rule.Condition["window_minutes"].(float64)
		if !ok1 || !ok2 {
			return false, ""
		}

		since := txCtx.Timestamp.Add(-time.Duration(windowMinutes) * time.Minute)
		txCount, err := s.repo.GetTransactionCount(ctx, txCtx.AccountID, since)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get transaction count")
			return false, ""
		}

		if txCount >= int(count) {
			return true, fmt.Sprintf("%d transactions in %d minutes", txCount, int(windowMinutes))
		}

	case RuleTypeTimePattern:
		startHour, ok1 := rule.Condition["start_hour"].(float64)
		endHour, ok2 := rule.Condition["end_hour"].(float64)
		if !ok1 || !ok2 {
			return false, ""
		}

		hour := txCtx.Timestamp.Hour()
		if hour >= int(startHour) && hour <= int(endHour) {
			return true, fmt.Sprintf("Transaction at suspicious time: %02d:00", hour)
		}
	}

	return false, ""
}

func (s *service) GetAlertByTransactionID(ctx context.Context, txID uuid.UUID) (*FraudAlert, error) {
	return s.repo.GetAlertByTransactionID(ctx, txID)
}

func (s *service) ListPendingAlerts(ctx context.Context, limit, offset int) ([]FraudAlert, error) {
	return s.repo.ListAlerts(ctx, AlertStatusPending, limit, offset)
}

func (s *service) ReviewAlert(ctx context.Context, alertID, reviewerID uuid.UUID, status AlertStatus) error {
	return s.repo.UpdateAlertStatus(ctx, alertID, status, reviewerID)
}

func (s *service) GetAccountRiskHistory(ctx context.Context, accountID uuid.UUID, limit int) ([]FraudScore, error) {
	return s.repo.GetScoresByAccountID(ctx, accountID, limit)
}

// Helper functions
func extractRuleNames(rules []FraudRule) []string {
	names := make([]string, len(rules))
	for i, rule := range rules {
		names[i] = rule.Name
	}
	return names
}
