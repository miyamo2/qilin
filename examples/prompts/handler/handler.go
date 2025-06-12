package handler

import (
	"fmt"
	"strings"

	"github.com/miyamo2/qilin"
)

// GreetingPromptHandler handles greeting prompts with name parameter
func GreetingPromptHandler(c qilin.PromptContext) error {
	args := c.Arguments()
	name, ok := args["name"]
	if !ok || name == "" {
		name = "World"
	}

	c.System("You are a friendly assistant that provides warm greetings to users.")
	c.User(fmt.Sprintf("Please greet %s warmly and make them feel welcome.", name))

	return nil
}

// EmailPromptHandler generates email templates with customizable parameters
func EmailPromptHandler(c qilin.PromptContext) error {
	args := c.Arguments()
	
	recipient, _ := args["recipient"].(string)
	if recipient == "" {
		recipient = "there"
	}

	subject, _ := args["subject"].(string)
	if subject == "" {
		subject = "Important Update"
	}

	tone, _ := args["tone"].(string)
	if tone == "" {
		tone = "professional"
	}

	c.System("You are an expert email writer who crafts clear, effective emails.")
	c.User(fmt.Sprintf("Write an email with the following details:\n- Recipient: %s\n- Subject: %s\n- Tone: %s\n\nMake it engaging and appropriate for the specified tone.", recipient, subject, tone))

	return nil
}

// CodeReviewPromptHandler generates code review templates
func CodeReviewPromptHandler(c qilin.PromptContext) error {
	args := c.Arguments()

	language, _ := args["language"].(string)
	if language == "" {
		language = "Go"
	}

	focus, _ := args["focus"].(string)
	if focus == "" {
		focus = "general"
	}

	systemPrompt := "You are a senior software engineer conducting a thorough code review."
	
	var userPrompt strings.Builder
	userPrompt.WriteString(fmt.Sprintf("Review the %s code focusing on %s aspects.\n", language, focus))
	userPrompt.WriteString("Consider the following areas:\n")
	userPrompt.WriteString("- Code quality and maintainability\n")
	userPrompt.WriteString("- Performance implications\n")
	userPrompt.WriteString("- Security considerations\n")
	userPrompt.WriteString("- Best practices adherence\n")
	userPrompt.WriteString("Provide constructive feedback with specific suggestions.")

	c.System(systemPrompt)
	c.User(userPrompt.String())

	return nil
}

// DocumentationPromptHandler generates documentation templates
func DocumentationPromptHandler(c qilin.PromptContext) error {
	args := c.Arguments()

	docType, _ := args["type"].(string)
	if docType == "" {
		docType = "API"
	}

	audience, _ := args["audience"].(string)
	if audience == "" {
		audience = "developers"
	}

	c.System("You are a technical writer who specializes in creating clear, comprehensive documentation.")
	c.User(fmt.Sprintf("Create %s documentation targeted at %s. Make it clear, well-structured, and include relevant examples where appropriate.", docType, audience))

	return nil
}