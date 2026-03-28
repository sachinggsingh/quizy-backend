package config

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

var (
	geminiClient *genai.Client
	geminiModel  string
)

// InitGemini initializes the Gemini client.
func InitGemini(ctx context.Context, apiKey, model string) error {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return fmt.Errorf("error creating Gemini client: %w", err)
	}
	geminiClient = client
	geminiModel = model
	return nil
}

// GenerateContent generates content using the initialized Gemini client.
func GenerateContent(ctx context.Context, prompt string) (string, error) {
	if geminiClient == nil {
		return "", fmt.Errorf("gemini client not initialized")
	}

	result, err := geminiClient.Models.GenerateContent(ctx, geminiModel, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}

	return result.Text(), nil
}
func BuildPrompt(title, category, difficulty, description string, numQuestions, points int) string {
	return fmt.Sprintf(`
	Generate a quiz with the following details:

Title: %s
Category: %s
Difficulty: %s
Description: %s
Number of Questions: %d
Total Points: %d

Return ONLY valid JSON:

{
  "title": "%s",
  "category": "%s",
  "description": "%s",
  "difficulty": "%s",
  "points": %d,
  "questions": [
    {
      "text": "question",
      "options": ["A", "B", "C", "D"],
      "answer": 0
    }
  ]
}

Rules:
- answer must be index (0-3)
- no explanation
- no extra text
`, title, category, difficulty, description, numQuestions, points, title, category, description, difficulty, points)
}

