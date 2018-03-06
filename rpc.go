package raft

// AppendRequest is sent by the leader to all its followers
// either as a heart beat or to commit logs
type AppendRequest struct {
	Term         int
	ID           int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []string
	LeaderCommit int
}

// AppendResult is sent by the followers to the leader in response to
// AppendRequest RPC
type AppendResult struct {
	Term    int
	Success bool
}

// minOf returns minimum of x , y
func minOf(x, y int) int {
	if x > y {
		return y
	}

	return x
}

// Append appends the entries sent by the server to commit
func (s *Server) Append(req *AppendRequest, res *AppendResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var failed bool

	defer func() {
		res.Term = s.currentTerm
		res.Success = !failed
	}()
	// check if this is from a stale leader
	if req.Term < s.currentTerm {
		failed = true
		return
	}

	// check if this server is stale
	if req.Term > s.currentTerm {
		s.state = follower
		s.currentTerm = req.Term
		s.votedFor = ""
	}

	// check for log entries
	if len(s.logs) < req.PrevLogIndex+1 {
		failed = true
		return
	}

	// check if the log term is same
	l := s.logs[req.PrevLogIndex]
	if l.term != req.PrevLogTerm {
		s.logs = s.logs[:req.PrevLogIndex]
		return
	}

	// lets append the new log to the logs
	for _, e := range req.Entries {
		s.logs = append(s.logs, log{term: req.Term, command: e})
	}

	// TODO: should i apply the new logs to state machine here ?
	// TODO check if this can be a problem at the start ?
	// maybe initialise commit index with -1
	if req.LeaderCommit > s.commitIndex {
		s.commitIndex = minOf(req.LeaderCommit, len(s.logs)-1)
	}

}

type RequestVoteReq struct {
	CandidateID                     string
	Term, LastLogIndex, LastLogTerm int
}

type RequestVoteResp struct {
	Term        int
	VoteGranted bool
}

func (s *Server) RequestVote(req *RequestVoteReq, res RequestVoteResp) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var failed bool

	defer func() {
		res.Term = s.currentTerm
		res.VoteGranted = !failed
		if !failed {
			s.votedFor = req.CandidateID
		}
	}()

	// if already voted, reject request
	if s.votedFor != "" {
		failed = true
		return
	}

	if s.currentTerm > req.Term {
		failed = true
		return
	}

	// no logs yet, then grant vote
	if len(s.logs) < 1 {
		failed = false
		return
	}

	li := len(s.logs) - 1
	le := s.logs[li]

	// check if the last entry's term is greater
	if le.term > req.LastLogTerm {
		failed = true
		return
	}

	// seems like we don't have up-to date logs
	// can restore remaining during heart beats
	if le.term < req.LastLogTerm {
		failed = false
		return
	}

	// see if the log index is up-to date
	if li > req.LastLogIndex {
		failed = true
		return
	}

	// seems like candidate has more up-to logs
	failed = false
}
