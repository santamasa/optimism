import 'git.just'

_EXTRALDFLAGS := if os() == "macos" { "-ldflags=-extldflags=-Wl,-ld_classic" } else { "" }

TARGETOS := env('GOOS', `go env GOOS`)
TARGETARCH := env('GOARCH', `go env GOARCH`)

GORACE := "0"

_GORACE_FLAG := if GORACE == "1" { "-race " } else { "" }

[private]
go_build BIN PKG *FLAGS:
    env GO111MODULE=on GOOS={{TARGETOS}} GOARCH={{TARGETARCH}} CGO_ENABLED=0 go build -v {{_GORACE_FLAG}} {{FLAGS}} -o {{BIN}} {{PKG}}

[private]
go_test SELECTOR *FLAGS:
    go test -v {{_GORACE_FLAG}} {{FLAGS}} {{SELECTOR}}

[private]
go_fuzz FUZZ TIME='10s' PKG='': (go_test PKG _EXTRALDFLAGS "-fuzztime" TIME "-fuzz" FUZZ "-run" "NOTAREALTEST")
