package github

type Repository struct {
	Owner string
	Name  string

	PullRequests  []*PullRequest
	Notifications []*Notification
}
