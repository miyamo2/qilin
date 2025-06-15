package qilin

import "errors"

const (
	ErrorMessageFailedToHandleTool = "failed to handle Tool (name: %s): %w"
)

var (
	ErrQilinLockingConflicts = errors.New(
		"qilin is already running or there is a configuration process conflict",
	)

	// ErrInvalidPromptRole occurs when an invalid prompt role is provided.
	ErrInvalidPromptRole = errors.New(
		"invalid prompt role, must be one of: user, assistant")
)
