package cmd

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/koki-develop/xbar-plugins/internal/github"
	"github.com/spf13/cobra"
)

//go:embed github.png
var githubIconBytes []byte
var githubIcon = base64.StdEncoding.EncodeToString(githubIconBytes)

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
		mpCnt := 0
		nCnt := 0

		{
			s := new(strings.Builder)

			repos, err := c.SearchPullRequestsReviewRequested()
			if err != nil {
				return err
			}

			for _, repo := range repos {
				rrCnt += len(repo.PullRequests)
			}
			for _, repo := range repos {
				fmt.Fprintf(s, "%s/%s | size=12\n", repo.Owner, repo.Name)
				for _, pr := range repo.PullRequests {
					fmt.Fprintf(s, "%s | href=%s\n", pr.Title, pr.URL)
				}
			}

			if rrCnt == 0 {
				fmt.Fprintln(s, "No review requested")
			}

			fmt.Fprintf(v, ":eyes: Review Requested (%d) | color=red\n", rrCnt)
			fmt.Fprint(v, s.String())
		}

		{
			s := new(strings.Builder)

			repos, err := c.SearchPullRequestsMine()
			if err != nil {
				return err
			}

			for _, repo := range repos {
				mpCnt += len(repo.PullRequests)

				fmt.Fprintf(s, "%s/%s | size=12\n", repo.Owner, repo.Name)
				for _, pr := range repo.PullRequests {
					fmt.Fprintf(s, "%s | href=%s\n", pr.Title, pr.URL)
				}
			}

			if mpCnt == 0 {
				fmt.Fprintln(s, "No pull requests")
			}

			fmt.Fprintln(v, "---")
			fmt.Fprintf(v, ":seedling: My Pull Requests (%d) | color=green\n", mpCnt)
			fmt.Fprint(v, s.String())
		}

		{
			s := new(strings.Builder)

			repos, err := c.FetchNotifications()
			if err != nil {
				return err
			}

			for _, repo := range repos {
				nCnt += len(repo.Notifications)

				fmt.Fprintf(s, "%s/%s | size=12\n", repo.Owner, repo.Name)
				for _, n := range repo.Notifications {
					fmt.Fprintf(s, "(%s) %s | href=%s\n", n.Reason, n.Title, n.URL)
					p, err := os.Executable()
					if err != nil {
						return err
					}
					fmt.Fprintf(s, "--Mark as read | shell=%s param1=github param2=read-notification param3=%s refresh=true\n", p, n.ID)
				}
			}

			if nCnt == 0 {
				fmt.Fprintln(s, "No notifications")
			}

			fmt.Fprintln(v, "---")
			fmt.Fprintf(v, ":bell: Notifications (%d) | color=blue\n", nCnt)
			fmt.Fprint(v, s.String())
		}

		fmt.Printf("(%d/%d/%d) | templateImage=%s\n", rrCnt, mpCnt, nCnt, githubIcon)
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
