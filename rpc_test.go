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

		// TODO: New logs with different and with out different term
		// TODO: commit index
		// TODO: last applied
	}

	f := new(Server)
	req := new(AppendRequest)
	res := new(AppendResult)

	for _, c := range tests {
		f.currentTerm = c.fTerm
		f.state = c.fstate
		f.votedFor = c.fVotedFor
		f.logs = c.flogs
		req.Term = c.lTerm
		req.PrevLogIndex = c.lPrevLogIndex
		req.PrevLogTerm = c.lPrevLogTerm
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
	}
}
