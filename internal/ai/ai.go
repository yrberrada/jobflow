package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Simple chat request/response structs for OpenAI's /v1/chat/completions
// (tested pattern; you may tweak model later).
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// EnrichJobWithLLM takes the raw job description and returns a cleaned,
// enriched note string you can put into Notion ("Notes" column).
func EnrichJobWithLLM(rawText, role, company string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// No API key? Just return the raw text so the rest still works.
		return "⚠️ OPENAI_API_KEY not set – using raw description.\n\n" + rawText, nil
	}

	systemPrompt := `You are a concise job-application assistant.

Given a job posting, you must:

1) Write a SHORT summary in 3–4 sentences (what the role is, team, impact).
2) List 5–10 core skills or technologies in bullet points.
3) Suggest 2–3 short, tailored talking points the candidate can reuse in cover letters or outreach.

Return everything as **plain text**, no JSON, no markdown headings. Use bullets with a simple dash "-".`

	userPrompt := fmt.Sprintf(
		"Company: %s\nRole: %s\n\nJob posting:\n%s",
		company, role, rawText,
	)

	reqBody := chatRequest{
		Model: "gpt-4.1-mini", // or gpt-4o-mini, gpt-3.5-turbo, etc.
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("OpenAI HTTP %d", resp.StatusCode)
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("decode OpenAI response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return cr.Choices[0].Message.Content, nil
}
