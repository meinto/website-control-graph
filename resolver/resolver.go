package resolver

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"fmt"

	"github.com/meinto/gqlgen-starter/graph/generated"
	"github.com/meinto/gqlgen-starter/model"
)

type Resolver struct{}

func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

func (r *queryResolver) Hello(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hi %s", name), nil
}

func (r *mutationResolver) Foo(ctx context.Context) (*model.Foo, error) {
	return &model.Foo{"bar"}, nil
}
