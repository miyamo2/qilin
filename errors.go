package qilin

import "errors"

const (
	ErrorMessageFailedToHandleTool = "failed to handle Tool (name: %s): %w"
)

var (
	ErrQilinLockingConflicts = errors.New(
		"qilin is already running or there is a configuration process conflict",
	)
)
