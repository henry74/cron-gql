package cron_gql

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"log"
	"os"
	"os/exec"

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

type queryResolver struct{ *Resolver }

func (r *queryResolver) Jobs(ctx context.Context, input *JobsInput) ([]*Job, error) {
	jobs := []*Job{}
	entryIDs := []int{}
	for _, v := range r.Cron.Entries() {
		entryIDs = append(entryIDs, int(v.ID))
	}
	for _, jobID := range entryIDs {
		job := r.RunningJobs[jobID]

		if input != nil && (len(input.Tags) > 0 || input.JobID != nil) {
			if input.JobID != nil { // jobID input takes precedence over tag input
				if *input.JobID == jobID {
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
