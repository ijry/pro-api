package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/protocol/ir"
)

// GenerateImage 调用 DALL-E 图片生成接口。
func (a *OpenAI) GenerateImage(ctx context.Context, req *ir.ImageRequest, cred adapter.Credential) (*ir.ImageResponse, error) {
	base := a.baseURL
	if cred.BaseURL != "" {
		base = cred.BaseURL
	}
	url := base + "/v1/images/generations"

	body := map[string]any{
		"model":  req.Model,
		"prompt": req.Prompt,
	}
	if req.N > 0 {
		body["n"] = req.N
	}
	if req.Size != "" {
		body["size"] = req.Size
	}
	if req.Quality != "" {
		body["quality"] = req.Quality
	}
	if req.Style != "" {
		body["style"] = req.Style
	}
	if req.ResponseFormat != "" {
		body["response_format"] = req.ResponseFormat
	}
	if req.User != "" {
		body["user"] = req.User
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAuthHeader(httpReq, cred.APIKey, "bearer")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, respBody)
	}

	var raw struct {
		Created int64 `json:"created"`
		Data    []struct {
			URL           string `json:"url"`
			B64JSON       string `json:"b64_json"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("openai: decode image response: %w", err)
	}

	irResp := &ir.ImageResponse{Created: raw.Created}
	for _, d := range raw.Data {
		irResp.Data = append(irResp.Data, ir.ImageData{
			URL:           d.URL,
			B64JSON:       d.B64JSON,
			RevisedPrompt: d.RevisedPrompt,
		})
	}
	return irResp, nil
}

// TextToSpeech 调用 OpenAI TTS 接口，返回音频二进制。
func (a *OpenAI) TextToSpeech(ctx context.Context, req *ir.SpeechRequest, cred adapter.Credential) (*ir.SpeechResponse, error) {
	base := a.baseURL
	if cred.BaseURL != "" {
		base = cred.BaseURL
	}
	url := base + "/v1/audio/speech"

	body := map[string]any{
		"model": req.Model,
		"input": req.Input,
		"voice": req.Voice,
	}
	if req.ResponseFormat != "" {
		body["response_format"] = req.ResponseFormat
	}
	if req.Speed != 0 {
		body["speed"] = req.Speed
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	setAuthHeader(httpReq, cred.APIKey, "bearer")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, b)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = audioContentType(req.ResponseFormat)
	}
	return &ir.SpeechResponse{ContentType: ct, Data: data}, nil
}

// Transcribe 调用 OpenAI Whisper 语音转文字接口。
func (a *OpenAI) Transcribe(ctx context.Context, req *ir.TranscribeRequest, cred adapter.Credential) (*ir.TranscribeResponse, error) {
	base := a.baseURL
	if cred.BaseURL != "" {
		base = cred.BaseURL
	}
	url := base + "/v1/audio/transcriptions"

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", req.Filename)
	if err != nil {
		return nil, err
	}
	if _, err = fw.Write(req.Audio); err != nil {
		return nil, err
	}
	_ = mw.WriteField("model", req.Model)
	if req.Language != "" {
		_ = mw.WriteField("language", req.Language)
	}
	if req.Prompt != "" {
		_ = mw.WriteField("prompt", req.Prompt)
	}
	if req.ResponseFormat != "" {
		_ = mw.WriteField("response_format", req.ResponseFormat)
	}
	mw.Close()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	setAuthHeader(httpReq, cred.APIKey, "bearer")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, adapter.ClassifyNetErr(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, adapter.ClassifyHTTPStatus(resp.StatusCode, respBody)
	}

	var raw struct {
		Text     string  `json:"text"`
		Language string  `json:"language"`
		Duration float64 `json:"duration"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		// plain-text response_format fallback
		return &ir.TranscribeResponse{Text: string(respBody)}, nil
	}
	return &ir.TranscribeResponse{
		Text:     raw.Text,
		Language: raw.Language,
		Duration: raw.Duration,
	}, nil
}

func audioContentType(format string) string {
	switch format {
	case "opus":
		return "audio/ogg"
	case "aac":
		return "audio/aac"
	case "flac":
		return "audio/flac"
	case "wav":
		return "audio/wav"
	case "pcm":
		return "audio/pcm"
	default:
		return "audio/mpeg"
	}
}
