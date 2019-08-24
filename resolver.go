package cron_gql

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/dustin/go-humanize"
	"github.com/robfig/cron/v3"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	Cron        *cron.Cron
	RunningJobs map[int]Job
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) AddJob(ctx context.Context, jobInput AddJobInput) (*Job, error) {
	entryID, err := r.Cron.AddFunc(jobInput.CronExp, func() { execute(jobInput) })
	job := Job{JobID: int(entryID), CronExp: jobInput.CronExp, RootDir: jobInput.RootDir, Cmd: jobInput.Cmd, Args: jobInput.Args, Tags: jobInput.Tags}
	r.RunningJobs[job.JobID] = job
	return &job, err
}

func (r *mutationResolver) RemoveJob(ctx context.Context, jobID int) (*Job, error) {
	r.Cron.Remove(cron.EntryID(jobID))
	job := r.RunningJobs[jobID]
	delete(r.RunningJobs, jobID)
	return &job, nil
}

func (r *mutationResolver) RunJob(ctx context.Context, jobID int) (*Job, error) {
	job := Job{}
	entry := r.Cron.Entry(cron.EntryID(jobID))
	if entry.Valid() {
		entry.Job.Run()
		job = r.RunningJobs[jobID]
		lastRun, nextRun := humanize.Time(entry.Prev), humanize.Time(entry.Next)
		job.LastRun = &lastRun
		job.NextRun = &nextRun

	}
	return &job, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Jobs(ctx context.Context, input *JobsInput) ([]*Job, error) {
	jobs := []*Job{}

	for _, entry := range r.Cron.Entries() {
		job, lastRun, nextRun := r.RunningJobs[int(entry.ID)], humanize.Time(entry.Prev), humanize.Time(entry.Next)
		job.LastRun = &lastRun
		job.NextRun = &nextRun

		if input != nil && (len(input.Tags) > 0 || input.JobID != nil) {
			if input.JobID != nil { // jobID input takes precedence over tag input
				if *input.JobID == int(entry.ID) {
					jobs = append(jobs, &job)
					break
				}
			} else if matchTags(input.Tags, job) {
				jobs = append(jobs, &job)
			}
		} else { // return all jobs
			jobs = append(jobs, &job)
		}
	}
	return jobs, nil
}

func matchTags(tagsToCheck []*string, job Job) bool {
	result := false
	tagMap := make(map[string]bool)
	for _, jobTag := range job.Tags {
		tagMap[*jobTag] = true
	}
	for _, tag := range tagsToCheck {
		if tagMap[*tag] {
			result = true
			break
		}
	}
	return result
}

func execute(job AddJobInput) (output string, err error) {
	log.Printf("Changing directory to '%s'", job.RootDir)
	err = os.Chdir(job.RootDir)
	if err != nil {
		log.Printf("%s", err)
	}
	log.Printf("Executing command '%s' with arguments '%s'", job.Cmd, *job.Args)

	out, err := exec.Command(job.Cmd, *job.Args).Output()
	if err != nil {
		log.Printf("%s", err)
	}
	output = string(out[:])
	log.Println(output)
	return output, err
}
