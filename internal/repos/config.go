package repos

import (
	"context"
	"github.com/marioidival/job-processing-system/internal/db"
	"github.com/marioidival/job-processing-system/pkg/database"
)

const defaultPollingInterval = 1000

type ConfigRepo struct {
	dbc *database.Client
	q   *db.Queries
}

func NewConfigRepo(dbc *database.Client) ConfigRepo {
	return ConfigRepo{
		dbc: dbc,
		q:   db.New(dbc),
	}
}

func (repo ConfigRepo) GetPollingInterval(ctx context.Context) int32 {
	interval, err := repo.q.GetPollingInterval(ctx)
	if err != nil {
		return defaultPollingInterval
	}
	return interval
}
