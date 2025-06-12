package main

import (
	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/examples/prompts/handler"
)

func main() {
	q := qilin.New("prompts-example")

	// Simple greeting prompt with optional name parameter
	q.Prompt("greeting", handler.GreetingPromptHandler,
		qilin.PromptWithDescription("A friendly greeting prompt"),
		qilin.PromptWithArguments(
			qilin.PromptArgument{
				Name:        "name",
				Description: "The name of the person to greet",
				Required:    false,
			},
		))

	// Email template prompt with multiple parameters
	q.Prompt("email", handler.EmailPromptHandler,
		qilin.PromptWithDescription("Generate professional email templates"),
		qilin.PromptWithArguments(
			qilin.PromptArgument{
				Name:        "recipient",
				Description: "The recipient of the email",
				Required:    false,
			},
			qilin.PromptArgument{
				Name:        "subject",
				Description: "The email subject line",
				Required:    false,
			},
			qilin.PromptArgument{
				Name:        "tone",
				Description: "The tone of the email (professional, casual, formal)",
				Required:    false,
			},
		))

	// Code review prompt with language and focus parameters
	q.Prompt("code_review", handler.CodeReviewPromptHandler,
		qilin.PromptWithDescription("Generate code review templates"),
		qilin.PromptWithArguments(
			qilin.PromptArgument{
				Name:        "language",
				Description: "Programming language to review",
				Required:    false,
			},
			qilin.PromptArgument{
				Name:        "focus",
				Description: "Review focus area (performance, security, maintainability)",
				Required:    false,
			},
		))

	// Documentation prompt
	q.Prompt("documentation", handler.DocumentationPromptHandler,
		qilin.PromptWithDescription("Generate documentation templates"),
		qilin.PromptWithArguments(
			qilin.PromptArgument{
				Name:        "type",
				Description: "Type of documentation (API, user guide, tutorial)",
				Required:    false,
			},
			qilin.PromptArgument{
				Name:        "audience",
				Description: "Target audience (developers, end-users, administrators)",
				Required:    false,
			},
		))

	q.Start()
}