package env

import (
	"context"
	"io"
	"os"
)

type contextKey string

var (
	ContextKeyStdout    = contextKey("stdout")
	ContextKeyStderr    = contextKey("stderr")
	ContextKeyArgs      = contextKey("args")
	ContextKeyLookupEnv = contextKey("lookupenv")
	ContextKeyGetwd     = contextKey("getwd")
)

func GetContextStdout(ctx context.Context) io.Writer {
	stdout, ok := ctx.Value(ContextKeyStdout).(io.Writer)
	if !ok {
		return os.Stdout
	}
	return stdout
}

func GetContextStderr(ctx context.Context) io.Writer {
	stderr, ok := ctx.Value(ContextKeyStderr).(io.Writer)
	if !ok {
		return os.Stderr
	}
	return stderr
}

func GetContextArgs(ctx context.Context) []string {
	args, ok := ctx.Value(ContextKeyArgs).([]string)
	if !ok {
		return os.Args
	}
	return args
}

func GetContextLookupEnv(ctx context.Context) func(string) (string, bool) {
	lookupEnv, ok := ctx.Value(ContextKeyLookupEnv).(func(string) (string, bool))
	if !ok {
		return os.LookupEnv
	}
	return lookupEnv
}

func GetContextGetwd(ctx context.Context) func() (string, error) {
	getwd, ok := ctx.Value(ContextKeyGetwd).(func() (string, error))
	if !ok {
		return os.Getwd
	}
	return getwd
}
