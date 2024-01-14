package github

import "github.com/cli/go-gh/v2/pkg/api"

type Client struct {
	gql *api.GraphQLClient
}

func NewClient(tkn string) (*Client, error) {
	gql, err := api.NewGraphQLClient(api.ClientOptions{AuthToken: tkn})
	if err != nil {
		return nil, err
	}

	return &Client{
		gql: gql,
	}, nil
}
