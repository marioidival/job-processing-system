package repos

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/marioidival/job-processing-system/internal/db"
	"github.com/marioidival/job-processing-system/pkg/database"
)

type Repo struct {
	dbc *database.Client
	q   *db.Queries
}

func NewJobRepo(dbc *database.Client) Repo {
	return Repo{
		dbc: dbc,
		q:   db.New(dbc),
	}
}

type Job struct {
	ID        int32     `json:"-"`
	JobUUID   uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	Data      []int32   `json:"data"`
	Action    string    `json:"action"`
	Result    int32     `json:"result,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r Repo) GetAllJobs(ctx context.Context) ([]Job, error) {
	jobsDB, err := r.q.GetAllJobs(ctx)
	if err != nil {
		return nil, err
	}
	return toJobs(jobsDB), nil
}

func (r Repo) GetJobsByStatus(ctx context.Context, status string) ([]Job, error) {
	jobsDB, err := r.q.GetJobsByStatus(ctx, db.JobStatus(strings.ToUpper(status)))
	if err != nil {
		return nil, err
	}
	return toJobs(jobsDB), nil
}

func (r Repo) CreateJob(ctx context.Context, action string, data []int32) error {
	return r.q.CreateJob(ctx, db.CreateJobParams{
		Action: action,
		Data:   data,
	})
}

func (r Repo) GetPendingJobs(ctx context.Context) ([]Job, error) {
	jobsDB, err := r.q.GetPendingJobs(ctx)
	if err != nil {
		return nil, err
	}
	return toPendingJobs(jobsDB), nil
}

func (r Repo) UpdateJob(ctx context.Context, jobID int32, status string, result int32) error {
	return r.q.UpdateJob(ctx, db.UpdateJobParams{
		Result: FromInt32(result),
		Job:    jobID,
		Status: db.JobStatus(status),
	})
}

func toJobs(jobs []db.Job) []Job {
	j := make([]Job, 0)
	for _, job := range jobs {
		j = append(j, toJob(job))
	}
	return j
}

func toJob(jobDB db.Job) Job {
	return Job{
		JobUUID:   jobDB.JobUuid,
		Status:    string(jobDB.Status),
		Data:      jobDB.Data,
		Result:    ToInt32(jobDB.Result),
		Action:    jobDB.Action,
		CreatedAt: jobDB.CreatedAt,
		UpdatedAt: jobDB.UpdatedAt,
	}
}

func toPendingJobs(jobs []db.GetPendingJobsRow) []Job {
	j := make([]Job, 0)
	for _, job := range jobs {
		j = append(j, toPendingJob(job))
	}
	return j
}

func toPendingJob(jobDB db.GetPendingJobsRow) Job {
	return Job{
		ID:     jobDB.ID,
		Data:   jobDB.Data,
		Action: jobDB.Action,
	}
}
