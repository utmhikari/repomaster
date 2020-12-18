package model

type RepoType string

const (
	RepoTypeUnknown RepoType = "unknown"
	RepoTypeGit     RepoType = "git"
	RepoTypeSvn     RepoType = "svn"
)

type RepoCommit struct{
	Hash string
	Ref string
	Message string
	Author string
	Email string
}

type RepoStatus string

const (
	RepoStatusUnknown RepoStatus = "unknown"
	RepoStatusError RepoStatus = "error"
	RepoStatusUpdating RepoStatus = "updating"
	RepoStatusActive RepoStatus = "active"
)

// Repo the info of a spefific repo
type Repo struct{
	// ID unique id
	ID 	   uint64

	// URL is the url of the repo of remote
	URL    string

	// Type is the type of repo
	Type   RepoType

	// Status is the current status of repo
	Status RepoStatus

	// Root is the local root dir of repo
	Root string

	// Desc is the description of repo
	Desc string

	// Commit is the current commmit info of repo
	Commit RepoCommit
}

// SetStatus set status of repo instance
func (r *Repo) SetStatus(status RepoStatus){
	switch status{
	case RepoStatusError:
		r.Status = RepoStatusError
		break
	case RepoStatusActive:
		r.Status = RepoStatusActive
		break
	case RepoStatusUpdating:
		r.Status = RepoStatusUpdating
		break
	default:
		r.Status = RepoStatusUnknown
		break
	}
}

// SetStatusError
func (r *Repo) SetStatusError(errMsg string){
	r.SetStatus(RepoStatusError)
	r.Desc = errMsg
}

// SetType set type of repo
func (r *Repo) SetType(repoType RepoType){
	switch repoType{
	case RepoTypeGit:
		r.Type = RepoTypeGit
		break
	case RepoTypeSvn:
		r.Type = RepoTypeSvn
		break
	default:
		r.Type = RepoTypeUnknown
		break
	}
}
