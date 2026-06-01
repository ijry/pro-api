package account

import (
	"bytes"
	"context"

	"github.com/ijry/pro-api/pkg/apierr"
)

// FormatParser 是各格式的解析器接口(由 importer/ 子包提供具体实现)。
type FormatParser interface {
	Format() string
	Match(payload []byte) bool
	Parse(ctx context.Context, payload []byte) ([]*Account, error)
}

// importerImpl 默认 Importer 实现。
type importerImpl struct {
	parsers []FormatParser
	oauth   OAuthFlow // 部分 parser(raw_refresh_token)需用 OAuth 换 AT
}

// NewImporter 构造一个无 parser 的 Importer;调用方需自行注册 parser。
// 一般场景请使用 NewDefaultImporter,它注册了 9 个默认 parser。
func NewImporter(oauth OAuthFlow) Importer {
	return &importerImpl{oauth: oauth}
}

// RegisterParser 追加一个 parser;Detect 按注册顺序匹配,先 JSON 复杂格式,后 raw 类。
func (i *importerImpl) RegisterParser(p FormatParser) {
	i.parsers = append(i.parsers, p)
}

func (i *importerImpl) Detect(payload []byte) (string, bool) {
	trimmed := bytes.TrimSpace(payload)
	for _, p := range i.parsers {
		if p.Match(trimmed) {
			return p.Format(), true
		}
	}
	return "", false
}

func (i *importerImpl) Parse(payload []byte, format string) ([]*Account, error) {
	trimmed := bytes.TrimSpace(payload)
	for _, p := range i.parsers {
		if p.Format() == format {
			return p.Parse(context.Background(), trimmed)
		}
	}
	return nil, apierr.New(apierr.CodeAccountImportFormat, "unknown format: "+format)
}
