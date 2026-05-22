package openai

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"time"

	"github.com/ijry/pro-api/internal/protocol/ir"
)

// EncodeChat 把 IR ChatResponse → OpenAI ChatResponseDTO。
func EncodeChat(resp *ir.ChatResponse) ChatResponseDTO {
	if resp == nil {
		return ChatResponseDTO{Object: "chat.completion", Created: time.Now().Unix()}
	}
	out := ChatResponseDTO{
		ID:                resp.ID,
		Object:            "chat.completion",
		Created:           time.Now().Unix(),
		Model:             resp.Model,
		SystemFingerprint: resp.SystemFingerprint,
		Usage: UsageDTO{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
			PromptTokensDetails: &PromptTokensDetailsDTO{
				CachedTokens: resp.Usage.CachedTokens,
			},
			CompletionTokensDetails: &CompletionTokensDetailsDTO{
				ReasoningTokens: resp.Usage.ReasoningTokens,
			},
		},
	}
	for _, c := range resp.Choices {
		out.Choices = append(out.Choices, ChoiceDTO{
			Index:        c.Index,
			FinishReason: c.FinishReason,
			Message:      encodeMessage(c.Message),
		})
	}
	return out
}

// encodeMessage 把 IR Message → wire MessageDTO。
//
//   - 单个 text part → content 编为 string
//   - 多 parts 或非 text → content 编为数组
func encodeMessage(m ir.Message) MessageDTO {
	out := MessageDTO{Role: m.Role, Name: m.Name, ToolCallID: m.ToolCallID}
	if len(m.Content) == 1 && m.Content[0].Type == ir.ContentText {
		out.Content, _ = json.Marshal(m.Content[0].Text)
	} else if len(m.Content) > 0 {
		parts := make([]ContentPartDTO, 0, len(m.Content))
		for _, p := range m.Content {
			d := ContentPartDTO{Type: p.Type, Text: p.Text}
			if p.ImageURL.URL != "" {
				d.ImageURL = &ImageURLDTO{URL: p.ImageURL.URL, Detail: p.ImageURL.Detail}
			}
			parts = append(parts, d)
		}
		out.Content, _ = json.Marshal(parts)
	}
	for _, tc := range m.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, ToolCallDTO{
			ID:   tc.ID,
			Type: tc.Type,
			Function: FunctionCallDTO{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return out
}

// EncodeEmbed 把 IR EmbedResponse → wire EmbedResponseDTO。
//
//   - encodingFormat="base64" 时,把 []float32 编码为 little-endian base64 字符串
//   - 否则直出 []float32(序列化为 JSON 数字数组)
func EncodeEmbed(resp *ir.EmbedResponse, encodingFormat string) EmbedResponseDTO {
	out := EmbedResponseDTO{
		Object: "list",
		Model:  resp.Model,
		Usage: EmbedUsageDTO{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}
	for _, d := range resp.Data {
		item := EmbedDataDTO{Object: "embedding", Index: d.Index}
		if encodingFormat == "base64" {
			if d.EmbeddingB64 != "" {
				item.Embedding = d.EmbeddingB64
			} else {
				item.Embedding = encodeFloatsBase64(d.Embedding)
			}
		} else {
			if len(d.Embedding) == 0 && d.EmbeddingB64 != "" {
				item.Embedding = d.EmbeddingB64
			} else {
				item.Embedding = d.Embedding
			}
		}
		out.Data = append(out.Data, item)
	}
	return out
}

// encodeFloatsBase64 把 []float32 编码为 little-endian base64 字符串。
func encodeFloatsBase64(fs []float32) string {
	buf := make([]byte, 4*len(fs))
	for i, v := range fs {
		binary.LittleEndian.PutUint32(buf[4*i:], math.Float32bits(v))
	}
	return base64.StdEncoding.EncodeToString(buf)
}

// EncodeChatAsCompletion 把 IR ChatResponse 渲染为 legacy /v1/completions 响应。
func EncodeChatAsCompletion(resp *ir.ChatResponse) CompletionResponseDTO {
	out := CompletionResponseDTO{
		ID:      resp.ID,
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Usage: UsageDTO{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
	for _, c := range resp.Choices {
		text := ""
		for _, p := range c.Message.Content {
			if p.Type == ir.ContentText {
				text += p.Text
			}
		}
		out.Choices = append(out.Choices, CompletionChoiceDTO{
			Index:        c.Index,
			Text:         text,
			FinishReason: c.FinishReason,
			Logprobs:     nil,
		})
	}
	return out
}
