package raft

import "sync"

// state of the each Server
type state int8

const (
	follower = iota
	candidate
	leader
)

// log is a single log item Server client holds
type log struct {
	term    int
	command string
}

// Server is a single peer participating in the raft protocol
type Server struct {
	currentTerm int
	votedFor    string
	logs        []log

	commitIndex int
	lastApplied int

	nextIndex  []int
	matchIndex []int
	state      state

	peers  []string // list of peer servers
	leader string   // leader id

	mu sync.RWMutex
}
