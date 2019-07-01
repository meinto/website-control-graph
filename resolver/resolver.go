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

type queryResolver struct{ *Resolver }

func (r *queryResolver) Control(ctx context.Context, timeout *int, actions []*model.Action, mapping []*model.OutputMap) (*model.Output, error) {
	c := chrome.New(20)
	if timeout != nil {
		t := time.Duration(*timeout)
		c = chrome.New(t)
	}
	return c.Run(actions, mapping)
}
