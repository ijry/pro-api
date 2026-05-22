package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/pkg/apierr"
)

// StreamWriter 把 IR ChatChunk 写成 OpenAI SSE 格式(data: <json>\n\n)。
//
// flush 由 handler 传入(通常是 http.Flusher.Flush)。本包不直接 import gin。
type StreamWriter struct {
	w     io.Writer
	flush func()
}

// NewStreamWriter 构造一个 SSE 写入器。flush 可为 nil(测试场景)。
func NewStreamWriter(w io.Writer, flush func()) *StreamWriter {
	return &StreamWriter{w: w, flush: flush}
}

// WriteChunk 把一个 IR chunk 写为一行 SSE。
func (s *StreamWriter) WriteChunk(chunk *ir.ChatChunk) error {
	dto := ChatChunkDTO{
		ID:      chunk.ID,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   chunk.Model,
		Choices: []ChoiceChunkDTO{{
			Index:        0,
			FinishReason: chunk.FinishReason,
			Delta: DeltaDTO{
				Role:    chunk.Delta.Role,
				Content: chunk.Delta.Content,
			},
		}},
	}
	for _, tc := range chunk.Delta.ToolCalls {
		dto.Choices[0].Delta.ToolCalls = append(dto.Choices[0].Delta.ToolCalls, ToolCallDTO{
			Index: tc.Index,
			ID:    tc.ID,
			Type:  "function",
			Function: FunctionCallDTO{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	if chunk.Usage != nil {
		dto.Usage = &UsageDTO{
			PromptTokens:     chunk.Usage.PromptTokens,
			CompletionTokens: chunk.Usage.CompletionTokens,
			TotalTokens:      chunk.Usage.TotalTokens,
		}
	}
	b, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", b); err != nil {
		return err
	}
	if s.flush != nil {
		s.flush()
	}
	return nil
}

// WriteDone 写流末标志 "data: [DONE]\n\n"。
func (s *StreamWriter) WriteDone() error {
	if _, err := io.WriteString(s.w, "data: [DONE]\n\n"); err != nil {
		return err
	}
	if s.flush != nil {
		s.flush()
	}
	return nil
}

// WriteError 在流中途出错时,以一个 OpenAI 格式 error chunk + [DONE] 关闭流。
func (s *StreamWriter) WriteError(e *apierr.Error, errType, errCode string) error {
	body := map[string]any{
		"error": map[string]any{
			"message": e.Message,
			"type":    errType,
			"code":    errCode,
		},
	}
	b, _ := json.Marshal(body)
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", b); err != nil {
		return err
	}
	if s.flush != nil {
		s.flush()
	}
	return s.WriteDone()
}
