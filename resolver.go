package cron_gql

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"fmt"
	"log"

	"github.com/robfig/cron/v3"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	Cron *cron.Cron
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) AddJob(ctx context.Context, input AddJobInput) (*Job, error) {
	entryId, err := r.Cron.AddFunc(input.CronExp, func() { log.Printf("Running this with schedule (%s)", input.CronExp) })
	result := Job{EntryID: fmt.Sprintf("%v", entryId)}
	return &result, err
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Jobs(ctx context.Context) ([]*Job, error) {
	panic("not implemented")
}
