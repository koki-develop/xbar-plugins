package github

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Notification struct {
	Title string
	URL   string

	Repository *Repository
}

func (c *Client) FetchNotifications() ([]*Repository, error) {
	ctx := context.Background()

	var apins []*struct {
		Subject *struct {
			Title string  `json:"title"`
			URL   *string `json:"url"`
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
			if n.Subject.URL == nil {
				return nil
			}

			var resource struct {
				HTMLURL string `json:"html_url"`
			}
			if err := c.rest.Get(*n.Subject.URL, &resource); err != nil {
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
			Title: n.Subject.Title,
			URL:   n.ResourceURL,
			Repository: &Repository{
				Name:  n.Repository.Name,
				Owner: n.Repository.Owner.Login,
			},
		})
	}

	repomap := map[*Repository][]*Notification{}
	for _, n := range ns {
		r := &Repository{
			Name:  n.Repository.Name,
			Owner: n.Repository.Owner,
		}
		repomap[r] = append(repomap[r], n)
	}

	var rs []*Repository
	for r, ns := range repomap {
		r.Notifications = ns
		rs = append(rs, r)
	}

	return rs, nil
}
