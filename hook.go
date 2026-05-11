package gari

type HookType int

const (
	BeforeInsert HookType = iota + 1
	AfterInsert
	BeforeUpdate
	AfterUpdate
	BeforeDelete
	AfterDelete
)

type HookFunc func(r *Record)

//func TimeStampUpdater
