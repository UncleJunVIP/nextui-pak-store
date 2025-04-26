package models

import "qlova.tech/sum"

type ScreenName struct {
	MainMenu,
	Installed,
	Updates,
	Browse,
	PakList,
	PakInfo,
	DownloadPak,
	UpdatePak,
	UninstallPak sum.Int[ScreenName]
}

var ScreenNames = sum.Int[ScreenName]{}.Sum()

type Screen interface {
	Name() sum.Int[ScreenName]
	Draw() (value ScreenReturn, exitCode int, e error)
}

type ScreenReturn interface {
	Value() interface{}
}

type WrappedString struct {
	Contents string
}

func NewWrappedString(s string) WrappedString {
	return WrappedString{Contents: s}
}

func (s WrappedString) Value() interface{} {
	return s.Contents
}
