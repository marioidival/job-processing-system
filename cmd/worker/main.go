package main

import (
	"context"
	"flag"
	"github.com/marioidival/job-processing-system/internal/repos"
	"github.com/marioidival/job-processing-system/pkg/database"
	"github.com/peterbourgon/ff"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

func run() error {
	fs := flag.NewFlagSet("api", flag.ExitOnError)

	var databaseURL string

	fs.StringVar(&databaseURL, "database-url", "", "e.g., postgres://username:password@localhost:5432/database_name")

	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix())
	if err != nil {
		return err
	}
	ctx := context.Background()
	dbc, err := database.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer dbc.Close()

	jobRepo := repos.NewJobRepo(dbc)
	configRepo := repos.NewConfigRepo(dbc)

	for {

		for i := 1; i < 5; i++ {
			go func(jobRepo *repos.Repo) {
				jobs, errRepo := jobRepo.GetPendingJobs(context.Background())
				if errRepo != nil {
					return
				}
				for _, pendingJob := range jobs {
					switch strings.ToLower(pendingJob.Action) {
					case "sum":
						var result int32
						for _, v := range pendingJob.Data {
							result += v
						}
						errExc := jobRepo.UpdateJob(context.Background(), pendingJob.ID, "PROCESSED", result)
						if errExc != nil {
							_ = jobRepo.UpdateJob(context.Background(), pendingJob.ID, "ERROR", 0)
						}
					case "sub":
						var result int32
						for _, v := range pendingJob.Data {
							result -= v
						}
						errExc := jobRepo.UpdateJob(context.Background(), pendingJob.ID, "PROCESSED", result)
						if errExc != nil {
							_ = jobRepo.UpdateJob(context.Background(), pendingJob.ID, "ERROR", 0)
						}
					default:
						log.Println("this action is not allowed", pendingJob.Action)
						_ = jobRepo.UpdateJob(context.Background(), pendingJob.ID, "ERROR", 0)
					}
				}
			}(&jobRepo)
		}
		interval := configRepo.GetPollingInterval(context.Background())
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

	return nil
}
