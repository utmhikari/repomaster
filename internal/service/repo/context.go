package repo

import "sync"

// Type repo type
type Type string

const (
	TypeUnknown Type = "unknown"
	TypeGit     Type = "git"
	TypeSvn     Type = "svn"
)

// IsValidType is type value valid
func IsValidType(t string) bool {
	return t == string(TypeGit) ||
		t == string(TypeSvn)
}

// Status repo status
type Status string

const (
	StatusUnknown  Status = "unknown"
	StatusError    Status = "error"
	StatusUpdating Status = "updating"
	StatusActive   Status = "active"
)

// IsValidStatus is status value valid
func IsValidStatus(s string) bool {
	return s == string(StatusActive) ||
		s == string(StatusError) ||
		s == string(StatusUpdating)
}

// Commit repo head commit info
type Commit struct {
	Hash    string `json:"hash"`
	Ref     string `json:"ref"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Email   string `json:"email"`
}

// Repo the info of a spefific repo
type Repo struct {
	// URL is the url of the repo of remote
	URL string `json:"url"`

	// Type is the type of repo
	Type Type `json:"type"`

	// Status is the current status of repo
	Status Status `json:"status"`

	// Desc is the description of repo
	Desc string `json:"desc"`

	// Commit is the current commmit info of repo
	Commit Commit `json:"commit"`
}

// IsActive is in active status
func (r *Repo) IsActive() bool {
	return r.Status == StatusActive
}

// IsStatusNormal is status in normal state
func (r *Repo) IsStatusNormal() bool {
	return r.Status == StatusUpdating || r.Status == StatusActive
}

// SetStatusError set to error status with description
func (r *Repo) SetStatusError(errMsg string) {
	r.Status = StatusError
	r.Desc = errMsg
}

// mu global runtime mutex
var mu sync.RWMutex

// context the repo context in repomaster runtime
type context struct {
	// root the local root of repo
	root string
	// mu mutex to protect repo instance
	mu sync.RWMutex
	// v the repo instance
	v Repo
}

// SetRepoStatus set status of repo instance with lock
func (c *context) SetRepoStatus(status Status) {
	c.mu.Lock()
	switch status {
	case StatusError:
		c.v.Status = StatusError
		break
	case StatusActive:
		c.v.Status = StatusActive
		break
	case StatusUpdating:
		c.v.Status = StatusUpdating
		break
	default:
		c.v.Status = StatusUnknown
		break
	}
	c.mu.Unlock()
}

// SetRepoStatusError set status of repo as error with lock
func (c *context) SetRepoStatusError(errMsg string) {
	c.mu.Lock()
	c.v.SetStatusError(errMsg)
	c.mu.Unlock()
}

// SetRepoType set type of repo with lock
func (c *context) SetRepoType(repoType Type) {
	c.mu.Lock()
	switch repoType {
	case TypeGit:
		c.v.Type = TypeGit
		break
	case TypeSvn:
		c.v.Type = TypeSvn
		break
	default:
		c.v.Type = TypeUnknown
		break
	}
	c.mu.Unlock()
}

// IsRepoStatusNormal is repo at normal status
func (c *context) IsRepoStatusNormal() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.IsStatusNormal()
}
