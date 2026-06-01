package account

import "github.com/ijry/pro-api/internal/app"

// NewDefaultImporter 用 9 个默认 parser 装配。注册顺序 = 优先级:
// JSON 复杂格式优先(避免 raw_* 把 JSON 当字符串吞掉),原始字符串放尾。
func NewDefaultImporter(oauth OAuthFlow) Importer {
	i := &importerImpl{oauth: oauth}
	i.RegisterParser(OAuthSession{})
	i.RegisterParser(ClaudeAuthJSON{})
	i.RegisterParser(CodexAuthJSON{})
	i.RegisterParser(Sub2API{})
	i.RegisterParser(CLIProxy{})
	i.RegisterParser(AuthsBatch{})
	i.RegisterParser(RawAccessToken{})
	i.RegisterParser(RawRefreshToken{OAuth: oauth})
	i.RegisterParser(RawAPIKey{})
	return i
}

// FacadeFrom 从 app.Application.AccountSvc 提取已 wire 的 Facade。
// 未装配或类型不匹配时返回 nil。
func FacadeFrom(a *app.Application) *Facade {
	if a == nil {
		return nil
	}
	if f, ok := a.AccountSvc.(*Facade); ok {
		return f
	}
	return nil
}
