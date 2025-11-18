package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// What we want to get back from the LLM.
type EnrichedJob struct {
	Summary      string   `json:"summary"`
	Skills       []string `json:"skills"`
	TailoredNote string   `json:"tailored_note"`
	RawSnippet   string   `json:"raw_snippet"`
}

// Minimal types to talk to /v1/chat/completions.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// EnrichJobWithLLM calls OpenAI once and tries to turn the
// job description into a structured EnrichedJob.
func EnrichJobWithLLM(rawText, role, company string) (EnrichedJob, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return EnrichedJob{}, fmt.Errorf("missing OPENAI_API_KEY")
	}

	// 1) Build the prompt
	prompt := fmt.Sprintf(`
You are an AI assistant for job seekers.

Summarize the job description and extract useful fields.

Return STRICT JSON only, with this exact shape:

{
  "summary": "3-6 sentence summary of the role",
  "skills": ["skill1", "skill2", "..."],
  "tailored_note": "short advice for this candidate (resume tweaks, strategy, etc.)",
  "raw_snippet": "most important 250 characters from the job description"
}

Do NOT add any extra keys or text outside the JSON.

JOB TITLE: %s
COMPANY: %s

DESCRIPTION:
%s
`, role, company, rawText)

	reqPayload := chatRequest{
		// You can switch to "gpt-4.1-mini" or another model if you prefer.
		Model: "gpt-4o-mini",
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return EnrichedJob{}, fmt.Errorf("marshal chat request: %w", err)
	}

	// 2) Build HTTP request
	httpReq, err := http.NewRequest(
		http.MethodPost,
		"https://api.openai.com/v1/chat/completions",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return EnrichedJob{}, fmt.Errorf("create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 3) Send it
	client := &http.Client{Timeout: 15 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return EnrichedJob{}, fmt.Errorf("call OpenAI: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(httpResp.Body)
		return EnrichedJob{}, fmt.Errorf("OpenAI HTTP %d: %s", httpResp.StatusCode, strings.TrimSpace(string(b)))
	}

	// 4) Decode the OpenAI chat response
	var chatResp chatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&chatResp); err != nil {
		return EnrichedJob{}, fmt.Errorf("decode chat response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return EnrichedJob{}, fmt.Errorf("no choices returned from OpenAI")
	}

	rawContent := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// 5) Try to parse the model's JSON into our EnrichedJob struct
	var ej EnrichedJob
	if err := json.Unmarshal([]byte(rawContent), &ej); err != nil {
		// Fallback: if the model didn't return valid JSON, just stuff the text into RawSnippet
		ej.RawSnippet = rawContent
	}

	return ej, nil
}
