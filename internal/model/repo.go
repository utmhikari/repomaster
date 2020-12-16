package model


type RepoType string

const (
	RepoTypeUnknown RepoType = "unknown"
	RepoTypeGit     RepoType = "git"
	RepoTypeSvn     RepoType = "svn"
)

type RepoCommit struct{
	Hash string
	Branch string
}

type Repo struct{
	URL    string
	Type   RepoType
	Commit RepoCommit
}

// NewRepo create a repo instance
func NewRepo() Repo{
	repo := Repo{}
	repo.Type = RepoTypeUnknown
	return repo
}
