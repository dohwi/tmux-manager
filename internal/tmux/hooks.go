package tmux

import "os/exec"

var isAvailableFn func() bool

func OverrideExecCommand(fn func(string, ...string) *exec.Cmd) func() {
	orig := execCommand
	execCommand = fn
	return func() { execCommand = orig }
}

func OverrideCurrentSession(fn func() string) func() {
	orig := currentSessionFn
	currentSessionFn = fn
	return func() { currentSessionFn = orig }
}

func OverrideIsAvailable(fn func() bool) func() {
	orig := isAvailableFn
	isAvailableFn = fn
	return func() { isAvailableFn = orig }
}
