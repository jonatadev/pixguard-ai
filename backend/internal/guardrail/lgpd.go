package guardrail

import (
	"regexp"
	"strings"

	"github.com/pixguard-ai/pixguard-ai/internal/domain"
)

var (
	reCPFFormatted  = regexp.MustCompile(`\b\d{3}\.\d{3}\.\d{3}-\d{2}\b`)
	reCPFContextual = regexp.MustCompile(`(?i)\b(cpf(?:\s+informado|\s+chave)?[:\s]+)(\d{11})\b`)
	reCNPJ          = regexp.MustCompile(`\b\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}\b`)
	rePixEVP        = regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`)
	reEmail         = regexp.MustCompile(`\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}\b`)
	rePhone         = regexp.MustCompile(`(?:\+55\s?)?(?:\(\d{2}\)\s?9?\d{4}-?\d{4}|\b\d{2}\s9\d{4}-?\d{4}\b)`)
)

type injectionRule struct {
	name    string
	pattern *regexp.Regexp
}

var injectionRules = []injectionRule{
	{name: "pt_ignore_instructions", pattern: regexp.MustCompile(`(?i)ignore\s+(?:todas\s+)?as\s+instru[cç][oõ]es`)},
	{name: "pt_disregard_rules", pattern: regexp.MustCompile(`(?i)desconsidere\s+(?:todas\s+)?as\s+regras`)},
	{name: "en_ignore_instructions", pattern: regexp.MustCompile(`(?i)ignore\s+(?:all\s+)?previous\s+instructions`)},
	{name: "pt_force_refund", pattern: regexp.MustCompile(`(?i)aprove\s+o\s+(?:estorno|med|reembolso)\s+imediatamente`)},
	{name: "system_prompt_probe", pattern: regexp.MustCompile(`(?i)(?:system\s+prompt|reveal\s+(?:your|the)\s+prompt)`)},
	{name: "role_hijack", pattern: regexp.MustCompile(`(?i)(?:voc[eê]\s+agora\s+[eé]|you\s+are\s+now)`)},
	{name: "jailbreak_keyword", pattern: regexp.MustCompile(`(?i)(?:jailbreak|bypass\s+guardrail)`)},
}

// Sanitize masks Brazilian LGPD PII before LLM calls and flags prompt-injection attempts.
func Sanitize(input string) (string, domain.GuardrailReport) {
	sanitized := input
	categories := make(map[string]int)
	findings := make([]domain.GuardrailFinding, 0)

	applyMask := func(category, rule string, re *regexp.Regexp, token string) {
		matches := re.FindAllStringIndex(sanitized, -1)
		if len(matches) == 0 {
			return
		}
		sanitized = re.ReplaceAllString(sanitized, token)
		categories[category] += len(matches)
		findings = append(findings, domain.GuardrailFinding{
			Category: category,
			Rule:     rule,
			Count:    len(matches),
		})
	}

	// Order matters: mask EVP UUIDs, formatted/contextual CPFs and CNPJs before phone numbers.
	applyMask("PIX_EVP", "uuid_evp_key", rePixEVP, "[PIX_EVP_REDACTED]")
	applyMask("CPF", "formatted_cpf", reCPFFormatted, "[CPF_REDACTED]")

	cpfCtxMatches := reCPFContextual.FindAllStringSubmatchIndex(sanitized, -1)
	if len(cpfCtxMatches) > 0 {
		sanitized = reCPFContextual.ReplaceAllString(sanitized, "${1}[CPF_REDACTED]")
		categories["CPF"] += len(cpfCtxMatches)
		findings = append(findings, domain.GuardrailFinding{
			Category: "CPF",
			Rule:     "contextual_11_digit_cpf",
			Count:    len(cpfCtxMatches),
		})
	}

	applyMask("CNPJ", "formatted_cnpj", reCNPJ, "[CNPJ_REDACTED]")
	applyMask("EMAIL", "email_address", reEmail, "[EMAIL_REDACTED]")
	applyMask("PHONE", "brazilian_phone", rePhone, "[PHONE_REDACTED]")

	injectionMatches := make([]string, 0)
	for _, rule := range injectionRules {
		if rule.pattern.MatchString(input) {
			injectionMatches = append(injectionMatches, rule.name)
			findings = append(findings, domain.GuardrailFinding{
				Category: "PROMPT_INJECTION",
				Rule:     rule.name,
				Count:    1,
			})
		}
	}

	totalPII := 0
	for _, count := range categories {
		totalPII += count
	}

	// Normalize extra spaces left only if input was empty/whitespace.
	sanitized = strings.TrimSpace(sanitized)
	if input == "" {
		sanitized = ""
	}

	detected := len(injectionMatches) > 0
	return sanitized, domain.GuardrailReport{
		PIIRedactionsCount:      totalPII,
		PIICategories:           categories,
		PromptInjectionDetected: detected,
		InjectionMatches:        injectionMatches,
		Findings:                findings,
		SafeForLLM:              !detected,
	}
}
