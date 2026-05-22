package ir

import (
	"fmt"

	"github.com/ijry/pro-api/pkg/apierr"
)

// Validate 校验 ChatRequest 的最低必备字段。
//
//   - model 不可空 → CodeMissingParam
//   - messages 至少 1 条 → CodeMissingParam
//   - 每条 message role 必须为合法枚举 → CodeInvalidParam
//   - tool message 必须有 tool_call_id → CodeInvalidParam
//   - assistant 的 tool_calls 内每条必须有 id+name
//   - content part type 不能为空
//   - max_tokens < 0 → CodeInvalidParam
//   - temperature ∉ [0, 2] → CodeInvalidParam
//   - top_p ∉ [0, 1] → CodeInvalidParam
//   - stop 最多 4 条
func Validate(r *ChatRequest) error {
	if r == nil {
		return apierr.New(apierr.CodeMissingParam, "request is nil")
	}
	if r.Model == "" {
		return apierr.New(apierr.CodeMissingParam, "model is required")
	}
	if len(r.Messages) == 0 {
		return apierr.New(apierr.CodeMissingParam, "messages must not be empty")
	}
	for i, m := range r.Messages {
		switch m.Role {
		case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		default:
			return apierr.New(apierr.CodeInvalidParam,
				fmt.Sprintf("messages[%d].role invalid: %s", i, m.Role))
		}
		if m.Role == RoleTool && m.ToolCallID == "" {
			return apierr.New(apierr.CodeInvalidParam,
				fmt.Sprintf("messages[%d] tool message requires tool_call_id", i))
		}
		if m.Role == RoleAssistant {
			for j, tc := range m.ToolCalls {
				if tc.ID == "" || tc.Function.Name == "" {
					return apierr.New(apierr.CodeInvalidParam,
						fmt.Sprintf("messages[%d].tool_calls[%d] missing id/name", i, j))
				}
			}
		}
		for j, p := range m.Content {
			if p.Type == "" {
				return apierr.New(apierr.CodeInvalidParam,
					fmt.Sprintf("messages[%d].content[%d].type empty", i, j))
			}
		}
	}
	if r.MaxTokens < 0 {
		return apierr.New(apierr.CodeInvalidParam, "max_tokens must be >= 0")
	}
	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return apierr.New(apierr.CodeInvalidParam, "temperature must be in [0,2]")
	}
	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return apierr.New(apierr.CodeInvalidParam, "top_p must be in [0,1]")
	}
	if len(r.Stop) > 4 {
		return apierr.New(apierr.CodeInvalidParam, "at most 4 stop sequences")
	}
	return nil
}

// ValidateEmbed 校验 EmbedRequest。
func ValidateEmbed(r *EmbedRequest) error {
	if r == nil {
		return apierr.New(apierr.CodeMissingParam, "request is nil")
	}
	if r.Model == "" {
		return apierr.New(apierr.CodeMissingParam, "model is required")
	}
	if len(r.Input) == 0 {
		return apierr.New(apierr.CodeMissingParam, "input must not be empty")
	}
	for i, s := range r.Input {
		if s == "" {
			return apierr.New(apierr.CodeInvalidParam,
				fmt.Sprintf("input[%d] must not be empty", i))
		}
	}
	return nil
}
