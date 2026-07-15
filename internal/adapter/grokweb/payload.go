package grokweb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

func buildSSOCookie(token string) string {
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "sso=")
	return "sso=" + token + "; sso-rw=" + token
}

func buildHeaders(cred adapter.Credential) http.Header {
	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	h.Set("Content-Type", "application/json")
	h.Set("Origin", "https://grok.com")
	h.Set("Referer", "https://grok.com/")
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-origin")
	h.Set("User-Agent", defaultUserAgent)
	h.Set("Cookie", buildSSOCookie(cred.APIKey))
	h.Set("x-statsig-id", randomID(16))
	h.Set("x-xai-request-id", randomID(16))
	return h
}

func randomID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return strings.Repeat("0", n*2)
	}
	return hex.EncodeToString(b)
}

func buildPayload(req *ir.ChatRequest) (map[string]any, error) {
	spec, ok := lookupModel(req.Model)
	if !ok {
		return nil, fmt.Errorf("grok-web: unsupported model %q", req.Model)
	}
	body := map[string]any{
		"deviceEnvInfo": map[string]any{
			"darkModeEnabled": false,
			"devicePixelRatio": 2,
			"screenWidth": 2056,
			"screenHeight": 1329,
			"viewportWidth": 2056,
			"viewportHeight": 1083,
		},
		"disableMemory": false,
		"disableSearch": false,
		"disableSelfHarmShortCircuit": false,
		"disableTextFollowUps": false,
		"enableImageGeneration": false,
		"enableImageStreaming": false,
		"enableSideBySide": true,
		"fileAttachments": []string{},
		"forceConcise": false,
		"forceSideBySide": false,
		"imageAttachments": []string{},
		"imageGenerationCount": 0,
		"isAsyncChat": false,
		"isReasoning": false,
		"message": flattenMessages(req.Messages),
		"modelMode": spec.ModelMode,
		"modelName": spec.ModelName,
		"responseMetadata": map[string]any{
			"requestModelDetails": map[string]any{"modelId": spec.ModelName},
		},
		"returnImageBytes": false,
		"returnRawGrokInXaiRequest": false,
		"sendFinalMetadata": true,
		"temporary": true,
		"toolOverrides": map[string]any{},
	}
	if req.Temperature != nil || req.TopP != nil {
		override := map[string]any{}
		if req.Temperature != nil {
			override["temperature"] = *req.Temperature
		}
		if req.TopP != nil {
			override["topP"] = *req.TopP
		}
		body["responseMetadata"].(map[string]any)["modelConfigOverride"] = override
	}
	return body, nil
}

func flattenMessages(messages []ir.Message) string {
	var parts []string
	lastUser := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == ir.RoleUser {
			lastUser = i
			break
		}
	}
	for i, msg := range messages {
		text := flattenContent(msg.Content)
		if strings.TrimSpace(text) == "" {
			continue
		}
		if i == lastUser {
			parts = append(parts, text)
			continue
		}
		role := msg.Role
		if role == "" {
			role = ir.RoleUser
		}
		parts = append(parts, role+": "+text)
	}
	return strings.Join(parts, "\n\n")
}

func flattenContent(parts []ir.ContentPart) string {
	var out []string
	for _, p := range parts {
		switch p.Type {
		case ir.ContentText:
			if strings.TrimSpace(p.Text) != "" {
				out = append(out, p.Text)
			}
		case ir.ContentImageURL:
			if p.ImageURL.URL != "" {
				out = append(out, "[image: "+p.ImageURL.URL+"]")
			}
		}
	}
	return strings.Join(out, "\n")
}
