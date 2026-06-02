package quota

import "github.com/ijry/pro-api/internal/account"

// NewTracker 用默认 anthropic + openai 提供商构造。
func NewTracker(repo account.Repo) account.QuotaTracker {
	return account.NewQuotaTracker(repo, map[string]account.ProviderExtractor{
		"anthropic": NewAnthropic(),
		"openai":    NewOpenAI(),
	})
}
