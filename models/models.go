package models

import (
	"gorm.io/gorm"
)

type Poll struct {
	gorm.Model
	Name             string
	Open             bool
	LimitIps         string
	PollLink         string
	AdminLink        string
	CompletePollLink string `gorm:"-:all"` //used for passing data to admin view
}

type PollOption struct {
	gorm.Model
	PollId int
	Name   string
	Votes  int
}

type VotedIp struct {
	gorm.Model
	Ip     string
	PollId int
}
