package api

import (
	"github.com/labstack/echo/v4"
	"github.com/marioidival/job-processing-system/internal/repos"
	"net/http"

	"github.com/marioidival/job-processing-system/pkg/database"
)

type Server struct {
	jobRepo repos.Repo
}

func NewServer(dbc *database.Client) *Server {
	return &Server{
		jobRepo: repos.NewJobRepo(dbc),
	}
}

type CreateJobRequest struct {
	Data   []int32 `json:"data"`
	Action string  `json:"action"`
}

func (s *Server) GetJobs(ctx echo.Context) error {
	ctxReq := ctx.Request().Context()
	queryParams := ctx.QueryParams()
	if status := queryParams.Get("status"); status != "" {
		// Get jobs by status
		jobs, err := s.jobRepo.GetJobsByStatus(ctxReq, status)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, echo.Map{"message": "no jobs found"})
		}
		return ctx.JSON(http.StatusOK, echo.Map{"jobs": jobs})
	}
	// get all jobs
	jobs, err := s.jobRepo.GetAllJobs(ctxReq)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, echo.Map{"message": "no jobs found"})
	}
	return ctx.JSON(http.StatusOK, echo.Map{"jobs": jobs})
}

func (s *Server) SaveJobs(ctx echo.Context) error {
	request := new(CreateJobRequest)
	if err := ctx.Bind(request); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	ctxReq := ctx.Request().Context()

	err := s.jobRepo.CreateJob(ctxReq, request.Action, request.Data)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable,
			echo.Map{"message": "it's not possible to create a job with your parameters"})
	}

	return ctx.JSON(http.StatusAccepted, echo.Map{"message": "job enqueued"})
}
