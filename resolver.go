package cron_gql

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/robfig/cron/v3"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// Resolver is...
type Resolver struct {
	Cron        *cron.Cron
	RunningJobs map[int]Job
}

// Mutation is...
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query is...
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
	job := Job{}
	if _, ok := r.RunningJobs[jobID]; ok {
		log.Println(ok)
		job = r.RunningJobs[jobID]
		log.Println(job)
		job.humanizeTime()
		delete(r.RunningJobs, jobID)
	}
	return &job, nil
}

func (r *mutationResolver) RunJob(ctx context.Context, jobID int) (*Job, error) {
	job := Job{}
	entry := r.Cron.Entry(cron.EntryID(jobID))
	if entry.Valid() {
		go entry.Job.Run()
		job = r.RunningJobs[jobID]
		lastRunTime, nextRunTime, forcedRunTime := int(entry.Prev.Unix()), int(entry.Next.Unix()), int(time.Now().Unix())
		job.LastScheduledTime = &lastRunTime
		job.NextScheduledTime = &nextRunTime
		job.LastForcedTime = &forcedRunTime
		job.humanizeTime()

		r.RunningJobs[jobID] = job

	}
	return &job, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Jobs(ctx context.Context, input *JobsInput) ([]*Job, error) {
	jobs := []*Job{}

	for _, entry := range r.Cron.Entries() {
		job, lastRunTime, nextRunTime := r.RunningJobs[int(entry.ID)], int(entry.Prev.Unix()), int(entry.Next.Unix())
		job.LastScheduledTime = &lastRunTime
		job.NextScheduledTime = &nextRunTime
		job.humanizeTime()

		if input != nil && (len(input.Tags) > 0 || input.JobID != nil) {
			if input.JobID != nil { // jobID input takes precedence over tag input
				if *input.JobID == int(entry.ID) {
					jobs = append(jobs, &job)
					break
				}
			} else if job.matchTags(input.Tags) {
				jobs = append(jobs, &job)
			}
		} else { // return all jobs
			jobs = append(jobs, &job)
		}
	}
	return jobs, nil
}

func (r *Job) matchTags(tagsToCheck []*string) bool {
	result := false
	tagMap := make(map[string]bool)
	for _, jobTag := range r.Tags {
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

func (r *Job) humanizeTime() {
	if r.LastScheduledTime != nil {
		lastScheduledRun := humanize.Time(time.Unix(int64(*r.LastScheduledTime), 0))
		r.LastScheduledRun = &lastScheduledRun
	}
	if r.NextScheduledTime != nil {
		nextScheduledRun := humanize.Time(time.Unix(int64(*r.NextScheduledTime), 0))
		r.NextScheduledRun = &nextScheduledRun
	}
	if r.LastForcedTime != nil {
		lastForcedRun := humanize.Time(time.Unix(int64(*r.LastForcedTime), 0))
		r.LastForcedRun = &lastForcedRun
	}
	return
}

func execute(job AddJobInput) (string, error) {
	log.Printf("Changing directory to '%s'", job.RootDir)
	err := os.Chdir(job.RootDir)
	if err != nil {
		log.Printf("%s", err)
	}
	var args []string
	for _, v := range job.Args {
		args = append(args, *v)
	}
	log.Printf("Executing command '%s' with arguments '%s'", job.Cmd, args)

	out, err := exec.Command(job.Cmd, args...).CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	output := string(out[:])
	log.Println(output)
	return output, err
}
