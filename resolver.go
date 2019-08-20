package cron_gql

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/robfig/cron/v3"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	Cron        *cron.Cron
	RunningJobs map[string]Job
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) AddJob(ctx context.Context, input AddJobInput) (*Job, error) {
	entryID, err := r.Cron.AddFunc(input.CronExp, func() { execute(input.RootDir, input.Cmd, input.Args) })
	job := Job{JobID: fmt.Sprintf("%v", entryID), CronExp: input.CronExp, RootDir: input.RootDir, Cmd: input.Cmd, Args: input.Args}
	r.RunningJobs[job.JobID] = job
	return &job, err
}

func (r *mutationResolver) RemoveJob(ctx context.Context, jobID int) (*Job, error) {
	r.Cron.Remove(cron.EntryID(jobID))
	job := r.RunningJobs[fmt.Sprintf("%v", jobID)]
	delete(r.RunningJobs, fmt.Sprintf("%v", jobID))
	return &job, nil

}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Jobs(ctx context.Context) ([]*Job, error) {
	jobs := []*Job{}
	entryIDs := []string{}
	for _, v := range r.Cron.Entries() {
		entryIDs = append(entryIDs, fmt.Sprintf("%v", v.ID))
	}
	for _, jobID := range entryIDs {
		job := r.RunningJobs[jobID]
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

func execute(pwd string, cmd string, args string) (output string, err error) {
	log.Printf("Changing directory to %s", pwd)
	err = os.Chdir(pwd)
	if err != nil {
		log.Printf("%s", err)
	}
	log.Printf("Executing command: %s %s", cmd, args)
	out, err := exec.Command(cmd, args).Output()
	if err != nil {
		log.Printf("%s", err)
	}
	output = string(out[:])
	log.Println(output)
	return output, err
}
