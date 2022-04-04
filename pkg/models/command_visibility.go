package models

type CommandVisibility = int

const (
	CommandVisNormal = iota
	CommandVisLow
	CommandVisSecret
	CommandVisHidden
)
