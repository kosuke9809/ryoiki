package jj

import "time"

type User struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

type Target struct {
	CommitID    string   `json:"commit_id"`
	Parents     []string `json:"parents"`
	ChangeID    string   `json:"change_id"`
	Description string   `json:"description"`
	Author      User     `json:"author"`
	Committer   User     `json:"committer"`
}

type Workspace struct {
	Name   string `json:"name"`
	Target Target `json:"target"`
}
