package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Notification struct {
	ID     string
	Reason string
	Title  string
	URL    string

	Repository *Repository
}

func (c *Client) FetchNotifications() ([]*Repository, error) {
	ctx := context.Background()

	var apins []*struct {
		ID      string `json:"id"`
		Reason  string `json:"reason"`
		Subject *struct {
			Title            string  `json:"title"`
			URL              *string `json:"url"`
			LatestCommentURL *string `json:"latest_comment_url"`
		}
		Repository *struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
		}

		ResourceURL string
	}
	if err := c.rest.Get("notifications", &apins); err != nil {
		return nil, err
	}

	sem := semaphore.NewWeighted(10)
	eg, ctx := errgroup.WithContext(ctx)

	for _, n := range apins {
		n := n
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		eg.Go(func() error {
			defer sem.Release(1)
			if n.Subject.URL == nil && n.Subject.LatestCommentURL == nil {
				return nil
			}

			var resource struct {
				HTMLURL string `json:"html_url"`
			}
			u := n.Subject.LatestCommentURL
			if u == nil {
				u = n.Subject.URL
			}
			if err := c.rest.Get(*u, &resource); err != nil {
				return err
			}
			n.ResourceURL = resource.HTMLURL

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	var ns []*Notification
	for _, n := range apins {
		ns = append(ns, &Notification{
			ID:     n.ID,
			Reason: n.Reason,
			Title:  n.Subject.Title,
			URL:    n.ResourceURL,
			Repository: &Repository{
				Name:  n.Repository.Name,
				Owner: n.Repository.Owner.Login,
			},
		})
	}

	return c.groupNotificationsByRepository(ns), nil
}

func (c *Client) MarkNotificationAsRead(id string) error {
	var resp map[string]any
	if err := c.rest.Patch(fmt.Sprintf("notifications/threads/%s", id), nil, &resp); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil
		}
		return err
	}
	fmt.Printf("%#v\n", resp)
	return nil
}

func (c *Client) groupNotificationsByRepository(ns []*Notification) []*Repository {
	repos := map[string]*Repository{}
	for _, n := range ns {
		key := n.Repository.Owner + "/" + n.Repository.Name
		if _, ok := repos[key]; !ok {
			repos[key] = &Repository{
				Owner: n.Repository.Owner,
				Name:  n.Repository.Name,
			}
		}
		repos[key].Notifications = append(repos[key].Notifications, n)
	}

	var rs []*Repository
	for _, repo := range repos {
		rs = append(rs, repo)
	}
	return rs
}
