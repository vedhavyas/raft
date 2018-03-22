package raft

import (
	"reflect"
	"testing"
)

func TestServer_Append(t *testing.T) {
	tests := []struct {
		fTerm         int
		lTerm         int
		fstate        state
		estate        state
		fVotedFor     string
		lPrevLogIndex int
		lPrevLogTerm  int
		flogs         []log
		rlogs         []log
		entries       []log
		lcommitIndex  int
		fcommitIndex  int
		flastApplied  int
		lastApplied   int
		result        bool
	}{

		// stale term
		{
			fTerm:  2,
			lTerm:  1,
			result: false,
		},

		// missing logs
		{
			fTerm:         0,
			lTerm:         1,
			fstate:        candidate,
			estate:        follower,
			fVotedFor:     "some identifier",
			lPrevLogIndex: 1,
			result:        false,
		},

		// log with different term
		{
			fTerm:         1,
			lTerm:         1,
			fstate:        candidate,
			estate:        follower,
			fVotedFor:     "some identifier",
			lPrevLogIndex: 1,
			lPrevLogTerm:  1,
			flastApplied:  -1,
			lastApplied:   0,
			flogs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    0,
					command: "test2",
				},
			},
			rlogs: []log{
				{
					term:    0,
					command: "test",
				},
			},
			result: true,
		},

		{
			fTerm:         1,
			lTerm:         1,
			fstate:        candidate,
			estate:        follower,
			fVotedFor:     "some identifier",
			lPrevLogIndex: 1,
			lPrevLogTerm:  1,
			flastApplied:  -1,
			lastApplied:   0,
			entries: []log{
				{
					term:    1,
					command: "test2",
				},

				{
					term:    1,
					command: "test3",
				},
			},
			flogs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    0,
					command: "test2",
				},
			},
			rlogs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    1,
					command: "test2",
				},

				{
					term:    1,
					command: "test3",
				},
			},
			result: true,
		},

		{
			fTerm:         1,
			lTerm:         1,
			fstate:        candidate,
			estate:        follower,
			fVotedFor:     "some identifier",
			lPrevLogIndex: 1,
			lPrevLogTerm:  1,
			lcommitIndex:  4,
			fcommitIndex:  2,
			flastApplied:  1,
			lastApplied:   2,
			entries: []log{
				{
					term:    1,
					command: "test2",
				},

				{
					term:    1,
					command: "test3",
				},
			},
			flogs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    0,
					command: "test2",
				},
			},
			rlogs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    1,
					command: "test2",
				},

				{
					term:    1,
					command: "test3",
				},
			},
			result: true,
		},

		{
			fTerm:         1,
			lTerm:         1,
			fstate:        candidate,
			estate:        follower,
			fVotedFor:     "some identifier",
			lPrevLogIndex: 1,
			lPrevLogTerm:  1,
			lcommitIndex:  4,
			fcommitIndex:  2,
			lastApplied:   2,
			flastApplied:  1,
			entries: []log{
				{
					term:    1,
					command: "test2",
				},

				{
					term:    1,
					command: "test3",
				},
			},
			flogs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    0,
					command: "test2",
				},
			},
			rlogs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    1,
					command: "test2",
				},

				{
					term:    1,
					command: "test3",
				},
			},
			result: true,
		},
	}

	f := new(Server)
	req := new(AppendRequest)
	res := new(AppendResult)

	for i, c := range tests {
		f.currentTerm = c.fTerm
		f.state = c.fstate
		f.votedFor = c.fVotedFor
		f.logs = c.flogs
		f.lastApplied = c.flastApplied
		req.Term = c.lTerm
		req.PrevLogIndex = c.lPrevLogIndex
		req.PrevLogTerm = c.lPrevLogTerm
		req.Entries = c.entries
		req.LeaderCommit = c.lcommitIndex
		f.Append(req, res)
		if res.Success != c.result {
			t.Fatalf("expected %t but got %t", c.result, res.Success)
		}

		if f.state != follower {
			t.Fatalf("expected follower state but got %v", f.state)
		}

		if res.Term != f.currentTerm {
			t.Fatalf("expected %d term but got %d", f.currentTerm, res.Term)
		}

		if f.votedFor != "" {
			t.Fatalf("expected empty voterdFor")
		}

		if !reflect.DeepEqual(f.logs, c.rlogs) {
			t.Fatalf("expected %v logs but got %v", c.rlogs, f.logs)
		}

		if c.fcommitIndex != f.commitIndex {
			t.Fatalf("expected commit index %d but got %d", c.fcommitIndex, f.commitIndex)
		}

		if c.lastApplied != f.lastApplied {
			t.Fatalf("%d: expected last applied %d but got %d", i, c.lastApplied, f.lastApplied)
		}

	}
}

func TestServer_RequestVote(t *testing.T) {
	tests := []struct {
		fVoted     string
		fTerm      int
		cTerm      int
		cLastTerm  int
		cLastIndex int
		logs       []log
		result     bool
	}{
		// already voted
		{
			fVoted: "someone else",
			fTerm:  2,
			cTerm:  3,
			result: false,
		},

		// canditate with lower term
		{
			fTerm:  4,
			cTerm:  3,
			result: false,
		},

		// no logs
		{
			fTerm:  2,
			cTerm:  3,
			result: true,
		},

		// follower has more updated logs
		{
			fTerm:     2,
			cTerm:     3,
			cLastTerm: 1,
			logs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    1,
					command: "test2",
				},

				{
					term:    2,
					command: "test3",
				},
			},
			result: false,
		},

		// candidate has more updated logs
		{
			fTerm:     1,
			cTerm:     3,
			cLastTerm: 2,
			logs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    1,
					command: "test2",
				},
			},
			result: true,
		},

		// same term but follower has longer logs
		{
			fTerm:      1,
			cTerm:      3,
			cLastTerm:  2,
			cLastIndex: 1,
			logs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    1,
					command: "test2",
				},

				{
					term:    2,
					command: "test3",
				},
			},
			result: false,
		},

		// same term but follower has less logs
		{
			fTerm:      1,
			cTerm:      3,
			cLastTerm:  2,
			cLastIndex: 3,
			logs: []log{
				{
					term:    0,
					command: "test",
				},

				{
					term:    2,
					command: "test2",
				},
			},
			result: true,
		},
	}

	f := new(Server)
	req := new(RequestVoteReq)
	res := new(RequestVoteResp)
	for _, c := range tests {
		f.currentTerm = c.fTerm
		f.votedFor = c.fVoted
		f.logs = c.logs
		req.Term = c.cTerm
		req.LastLogTerm = c.cLastTerm
		req.LastLogIndex = c.cLastIndex
		f.RequestVote(req, res)
		if res.VoteGranted != c.result {
			t.Fatalf("expected %t but got %t", c.result, res.VoteGranted)
		}

		if res.Term != c.fTerm {
			t.Fatalf("expected %d term but got %d term", c.fTerm, res.Term)
		}
	}
}
