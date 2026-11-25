package agents

import (
	"fmt"
	"time"

	"github.com/jonatadev/pixguard-ai/internal/domain"
)

const medMaxWindowDays = 80

// RunCustomerProfileTool returns deterministic customer relationship signals.
func RunCustomerProfileTool(c domain.PixCase) ToolTrace {
	start := time.Now()
	highValue := c.AmountBRL >= 5000.0
	return ToolTrace{
		Name:    "get_customer_profile",
		Summary: fmt.Sprintf("Perfil %s analisado; valor R$ %.2f (alto_valor=%v).", c.CustomerAlias, c.AmountBRL, highValue),
		Data: map[string]any{
			"customer_alias": c.CustomerAlias,
			"amount_brl":     c.AmountBRL,
			"high_value_tx":  highValue,
		},
		DurationMs: time.Since(start).Milliseconds(),
	}
}

// RunRecentTransactionsTool returns deterministic velocity/timing anomalies for the Pix transaction.
func RunRecentTransactionsTool(c domain.PixCase) ToolTrace {
	start := time.Now()
	pldftCandidate := c.AmountBRL >= 5000.0 && c.RecipientAccountAgeDays <= 15
	return ToolTrace{
		Name:    "get_recent_pix_transactions",
		Summary: fmt.Sprintf("Transacao chave %s ha %d dia(s); candidato PLD-FT=%v.", c.PixKeyType, c.DaysSinceTx, pldftCandidate),
		Data: map[string]any{
			"case_id":         c.ID,
			"pix_key_type":    c.PixKeyType,
			"days_since_tx":   c.DaysSinceTx,
			"pldft_candidate": pldftCandidate,
		},
		DurationMs: time.Since(start).Milliseconds(),
	}
}

// RunRecipientReputationTool evaluates the receiving PSP account age and mule-account indicators.
func RunRecipientReputationTool(c domain.PixCase) ToolTrace {
	start := time.Now()
	newlyOpened := c.RecipientAccountAgeDays < 15
	cautelarTrigger := newlyOpened && c.AmountBRL > 1500.0
	return ToolTrace{
		Name:    "check_recipient_reputation",
		Summary: fmt.Sprintf("PSP %s com conta de %d dias (recem_aberta=%v, gatilho_cautelar=%v).", c.RecipientBank, c.RecipientAccountAgeDays, newlyOpened, cautelarTrigger),
		Data: map[string]any{
			"recipient_bank":             c.RecipientBank,
			"recipient_account_age_days": c.RecipientAccountAgeDays,
			"newly_opened_account":       newlyOpened,
			"cautelar_trigger":           cautelarTrigger,
		},
		DurationMs: time.Since(start).Milliseconds(),
	}
}

// RunMEDWindowTool checks the statutory 80-day MED request deadline.
func RunMEDWindowTool(c domain.PixCase) ToolTrace {
	start := time.Now()
	valid := c.DaysSinceTx <= medMaxWindowDays
	remaining := medMaxWindowDays - c.DaysSinceTx
	return ToolTrace{
		Name:    "check_med_window",
		Summary: fmt.Sprintf("Janela MED 80d: days_since_tx=%d, valid=%v, remaining_days=%d.", c.DaysSinceTx, valid, remaining),
		Data: map[string]any{
			"days_since_tx":  c.DaysSinceTx,
			"max_days":       medMaxWindowDays,
			"valid":          valid,
			"remaining_days": remaining,
		},
		DurationMs: time.Since(start).Milliseconds(),
	}
}
