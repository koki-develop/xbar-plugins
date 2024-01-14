package cmd

import (
	"fmt"
	"strings"

	"github.com/koki-develop/xbar-plugins/internal/github"
	"github.com/spf13/cobra"
)

var (
	githubToken string
)

var githubCmd = &cobra.Command{
	Use: "github",
	RunE: func(cmd *cobra.Command, args []string) error {
		v := new(strings.Builder)

		c, err := github.NewClient(githubToken)
		if err != nil {
			return err
		}
		cnt := 0

		{
			repos, err := c.SearchPullRequestsReviewRequested()
			if err != nil {
				return err
			}

			for _, repo := range repos {
				cnt += len(repo.PullRequests)
			}
			fmt.Fprintf(v, "GitHub (%d)\n", cnt)
			fmt.Fprintln(v, "---")

			fmt.Fprintln(v, "Review Requested")
			for _, repo := range repos {
				fmt.Fprintf(v, "%s/%s | size=12\n", repo.Owner, repo.Name)
				for _, pr := range repo.PullRequests {
					fmt.Fprintf(v, "%s | href=%s\n", pr.Title, pr.URL)
				}
			}
		}

		{
			fmt.Fprintln(v, "---")
			fmt.Fprintln(v, "My Pull Requests")

			repos, err := c.SearchPullRequestsMine()
			if err != nil {
				return err
			}

			for _, repo := range repos {
				fmt.Fprintf(v, "%s/%s | size=12\n", repo.Owner, repo.Name)
				for _, pr := range repo.PullRequests {
					fmt.Fprintf(v, "%s | href=%s\n", pr.Title, pr.URL)
				}
			}
		}

		fmt.Println(v.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(githubCmd)
}
