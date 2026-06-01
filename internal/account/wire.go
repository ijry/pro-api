package account

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
