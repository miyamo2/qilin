package qilin_test

import (
	"fmt"
	"github.com/miyamo2/qilin"
)

func ExamplePromptRole_String() {
	fmt.Println(qilin.PromptRoleUser.String())
	fmt.Println(qilin.PromptRoleAssistant.String())

	var role qilin.PromptRole = -1
	fmt.Println(role.String())

	// Output: user
	// assistant
	// unknown
}

func ExamplePromptRole_Validate() {
	fmt.Println(qilin.PromptRoleUser.Validate())
	fmt.Println(qilin.PromptRoleAssistant.Validate())

	var role qilin.PromptRole = -1
	fmt.Println(role.Validate())

	// Output: <nil>
	// <nil>
	// invalid prompt role, must be one of: user, assistant
}

func ExamplePromptRoleFromString() {
	role := qilin.PromptRoleFromString("user")
	fmt.Println(role.String())

	role = qilin.PromptRoleFromString("assistant")
	fmt.Println(role.String())

	role = qilin.PromptRoleFromString("other")
	fmt.Println(role.String())

	// Output: user
	// assistant
	// unknown
}
