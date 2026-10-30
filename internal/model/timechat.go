package model

import "time"

type TimeChannelKind string

const (
	TimeChannelDM      TimeChannelKind = "dm"
	TimeChannelGroup   TimeChannelKind = "group"
	TimeChannelPublic  TimeChannelKind = "public"
	TimeChannelPrivate TimeChannelKind = "private"
)

type TimeChannel struct {
	ID           string
	Name         string
	Title        string
	Kind         TimeChannelKind
	Unread       int64
	Mentions     int64
	Muted        bool
	LastPostAt   time.Time
	LastViewedAt time.Time
}

type TimeMessage struct {
	ID           string
	RootID       string
	ChannelID    string
	ChannelTitle string
	Author       string
	FromMe       bool
	At           time.Time
	Text         string
	Replies      int
	Files        int
	System       bool
}

type TimeChannelPosts struct {
	Channel  TimeChannel
	Since    *time.Time
	Messages []TimeMessage
	Total    int
}

type TimeSearchOutput struct {
	Terms    string
	Messages []TimeMessage
}

type TimeUpdatesInput struct {
	Since         *time.Time
	Kinds         []TimeChannelKind
	IncludeMuted  bool
	MaxPerChannel int
}

type TimeUpdatesOutput struct {
	Since    *time.Time
	Channels []TimeChannelPosts
	Skipped  []TimeChannel
}
