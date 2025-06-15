package qilin

import (
	"errors"
	"testing"
)

func TestPromptRole_Validate(t *testing.T) {
	type test struct {
		role PromptRole
		err  error
	}
	tests := map[string]test{
		"user": {
			role: PromptRoleUser,
			err:  nil,
		},
		"assistant": {
			role: PromptRoleAssistant,
			err:  nil,
		},
		"other": {
			role: -1,
			err:  ErrInvalidPromptRole,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tc.role.Validate(); !errors.Is(err, tc.err) {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
		})
	}
}

func TestPromptRole_String(t *testing.T) {
	type test struct {
		role PromptRole
		str  string
	}
	tests := map[string]test{
		"user": {
			role: PromptRoleUser,
			str:  "user",
		},
		"assistant": {
			role: PromptRoleAssistant,
			str:  "assistant",
		},
		"other": {
			role: -1,
			str:  "unknown",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if str := tc.role.String(); str != tc.str {
				t.Errorf("expected string %s, got %s", tc.str, str)
			}
		})
	}
}

func TestPromptRoleFromString(t *testing.T) {
	type test struct {
		str        string
		promptRole PromptRole
	}
	tests := map[string]test{
		"user": {
			str:        "user",
			promptRole: PromptRoleUser,
		},
		"assistant": {
			str:        "assistant",
			promptRole: PromptRoleAssistant,
		},
		"other": {
			str:        "other",
			promptRole: PromptRoleUnknown,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			promptRole := PromptRoleFromString(tc.str)
			if promptRole != tc.promptRole {
				t.Errorf("expected %v, got %v", tc.promptRole, promptRole)
			}
		})
	}
}
