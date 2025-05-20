package models

import "qlova.tech/sum"

type ScreenName struct {
	MainMenu,
	Browse,
	PakList,
	PakInfo,
	DownloadPak,
	Updates,
	ManageInstalled sum.Int[ScreenName]
}

var ScreenNames = sum.Int[ScreenName]{}.Sum()

type Screen interface {
	Name() sum.Int[ScreenName]
	Draw() (value interface{}, exitCode int, e error)
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
