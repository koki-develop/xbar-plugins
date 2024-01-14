package github

import (
	"strings"

	graphql "github.com/cli/shurcooL-graphql"
)

type PullRequest struct {
	Title string
	URL   string

	Repository *Repository
}

func (c *Client) SearchPullRequestsReviewRequested() ([]*Repository, error) {
	var q struct {
		Search struct {
			Edges []struct {
				Node struct {
					PullRequest struct {
						Title      string
						URL        string
						Repository struct {
							Name  string
							Owner struct {
								Login string
							}
						}
					} `graphql:"... on PullRequest"`
				}
			}
		} `graphql:"search(query: $query, type: ISSUE, first: 100)"`
	}
	filter := strings.Join([]string{
		"is:pr",
		"is:open",
		"review-requested:@me",
		"-reviewed-by:@me",
		"-author:app/renovate",
		"-author:app/dependabot",
	}, " ")
	if err := c.gql.Query("query", &q, map[string]any{"query": graphql.String(filter)}); err != nil {
		return nil, err
	}

	reposmap := map[*Repository][]*PullRequest{}
	for _, edge := range q.Search.Edges {
		pr := &PullRequest{
			Title: edge.Node.PullRequest.Title,
			URL:   edge.Node.PullRequest.URL,
			Repository: &Repository{
				Name:  edge.Node.PullRequest.Repository.Name,
				Owner: edge.Node.PullRequest.Repository.Owner.Login,
			},
		}
		reposmap[pr.Repository] = append(reposmap[pr.Repository], pr)
	}

	repos := make([]*Repository, 0, len(reposmap))
	for repo, prs := range reposmap {
		repo.PullRequests = prs
		repos = append(repos, repo)
	}

	return repos, nil
}
