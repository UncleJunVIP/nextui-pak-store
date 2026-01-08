package ui

import gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"

type ScreenResult[T any] struct {
	Value    T
	ExitCode gaba.ExitCode
}

func success[T any](value T) ScreenResult[T] {
	return ScreenResult[T]{
		Value:    value,
		ExitCode: gaba.ExitCodeSuccess,
	}
}

func back[T any](value T) ScreenResult[T] {
	return ScreenResult[T]{
		Value:    value,
		ExitCode: gaba.ExitCodeBack,
	}
}

func withCode[T any](value T, code gaba.ExitCode) ScreenResult[T] {
	return ScreenResult[T]{
		Value:    value,
		ExitCode: code,
	}
}
