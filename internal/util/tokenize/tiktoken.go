package tokenize

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

// Tiktoken 是 OpenAI 系列模型的 tokenizer 包装。
type Tiktoken struct {
	encoding string
	enc      *tiktoken.Tiktoken
}

// NewTiktoken 用 encoding 名(如 "cl100k_base" / "o200k_base")构造。
func NewTiktoken(encoding string) (*Tiktoken, error) {
	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, fmt.Errorf("tokenize: tiktoken get encoding %q: %w", encoding, err)
	}
	return &Tiktoken{encoding: encoding, enc: enc}, nil
}

// Count 返回 text 在该编码下的 token 数。
func (t *Tiktoken) Count(_, text string) (int, error) {
	if text == "" {
		return 0, nil
	}
	return len(t.enc.Encode(text, nil, nil)), nil
}

// CountMessages 实现 OpenAI 官方近似算法:
//
//	每条消息 +3 token(role 等开销)+ content token
//	每条消息有 Name 字段时 +1 token
//	最后整个会话 +3 token(reply primer)
func (t *Tiktoken) CountMessages(model string, messages []Message) (int, error) {
	total := 0
	for _, m := range messages {
		total += 3
		c, err := t.Count(model, m.Content)
		if err != nil {
			return 0, err
		}
		total += c
		if m.Name != "" {
			total += 1
			n, err := t.Count(model, m.Name)
			if err != nil {
				return 0, err
			}
			total += n
		}
		if m.Role != "" {
			n, err := t.Count(model, m.Role)
			if err != nil {
				return 0, err
			}
			total += n
		}
	}
	total += 3
	return total, nil
}

// Name 返回 tokenizer 标识。
func (t *Tiktoken) Name() string { return "tiktoken/" + t.encoding }
