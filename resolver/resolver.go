package resolver

//go:generate go run github.com/99designs/gqlgen

import (
	"context"

	"github.com/meinto/gqlgen-starter/chrome"
	"github.com/meinto/gqlgen-starter/graph/generated"
	"github.com/meinto/gqlgen-starter/model"
)

type Resolver struct{}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Control(ctx context.Context, actions []*model.Action, mapping []*model.WebsiteElement) ([]*model.Output, error) {
	return chrome.Run(actions, mapping)
}
