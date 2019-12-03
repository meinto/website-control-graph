package resolver

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"time"

	"github.com/meinto/website-control-graph/chrome"
	"github.com/meinto/website-control-graph/graph/generated"
	"github.com/meinto/website-control-graph/model"
)

type Resolver struct{}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }

func (r *queryResolver) Control(
	ctx context.Context,
	timeout *int,
	omitEmpty *bool,
	actions []*model.Action,
	mapping []*model.OutputCollections) (*model.Output, error) {
	c := chrome.New(20, omitEmpty)
	if timeout != nil {
		t := time.Duration(*timeout)
		c = chrome.New(t, omitEmpty)
	}
	return c.Run(actions, mapping)
}

func (r *mutationResolver) Control(
	ctx context.Context,
	timeout *int,
	omitEmpty *bool,
	actions []*model.Action,
	mapping []*model.OutputCollections) (*model.Output, error) {
	c := chrome.New(20, omitEmpty)
	if timeout != nil {
		t := time.Duration(*timeout)
		c = chrome.New(t, omitEmpty)
	}
	return c.Run(actions, mapping)
}
