module github.com/miyamo2/qilin/tests/e2e

go 1.24

toolchain go1.24.0

require (
	github.com/google/uuid v1.6.0
	github.com/miyamo2/qilin v0.0.0
	github.com/stretchr/testify v1.10.0
	go.lsp.dev/jsonrpc2 v0.10.0
)

replace github.com/miyamo2/qilin => ../../

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.3.4 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	golang.org/x/exp/event v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/exp/jsonrpc2 v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
