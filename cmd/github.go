package cmd

import (
	"fmt"
	"os"
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

		rrCnt := 0
		nCnt := 0

		{
			repos, err := c.SearchPullRequestsReviewRequested()
			if err != nil {
				return err
			}

			for _, repo := range repos {
				rrCnt += len(repo.PullRequests)
			}
			fmt.Fprintln(v, ":eyes: Review Requested | color=red")
			for _, repo := range repos {
				fmt.Fprintf(v, "%s/%s | size=12\n", repo.Owner, repo.Name)
				for _, pr := range repo.PullRequests {
					fmt.Fprintf(v, "%s | href=%s\n", pr.Title, pr.URL)
				}
			}
		}

		{
			fmt.Fprintln(v, "---")
			fmt.Fprintln(v, ":seedling: My Pull Requests | color=green")

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

		{
			fmt.Fprintln(v, "---")
			fmt.Fprintln(v, ":bell: Notifications | color=blue")

			repos, err := c.FetchNotifications()
			if err != nil {
				return err
			}

			for _, repo := range repos {
				nCnt += len(repo.Notifications)

				fmt.Fprintf(v, "%s/%s | size=12\n", repo.Owner, repo.Name)
				for _, n := range repo.Notifications {
					fmt.Fprintf(v, "(%s) %s | href=%s\n", n.Reason, n.Title, n.URL)
					p, err := os.Executable()
					if err != nil {
						return err
					}
					fmt.Fprintf(v, "--Mark as read | shell=%s param1=github param2=read-notification param3=%s refresh=true\n", p, n.ID)
				}
			}
		}

		fmt.Printf("GitHub (%d/%d)\n", rrCnt, nCnt)
		fmt.Println("---")
		fmt.Println(v.String())
		return nil
	},
}

var readNotificationCmd = &cobra.Command{
	Use:  "read-notification",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := github.NewClient(githubToken)
		if err != nil {
			return err
		}

		if err := c.MarkNotificationAsRead(args[0]); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(githubCmd)
	githubCmd.AddCommand(readNotificationCmd)
}
