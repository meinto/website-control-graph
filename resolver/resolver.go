package resolver

//go:generate go run github.com/99designs/gqlgen

import (
	"context"

	"github.com/meinto/website-control-graph/chrome"
	"github.com/meinto/website-control-graph/graph/generated"
	"github.com/meinto/website-control-graph/model"
)

type Resolver struct{}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Control(ctx context.Context, actions []*model.Action, mapping []*model.WebsiteElement) ([]*model.Output, error) {
	return chrome.Run(actions, mapping)
}
