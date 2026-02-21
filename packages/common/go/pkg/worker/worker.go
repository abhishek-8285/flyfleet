package worker

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// Client handles job submission
type Client struct {
	client *asynq.Client
}

// NewClient creates a new Asynq client
func NewClient(redisAddr string) *Client {
	return &Client{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

// SubmitJob enqueues a new background job
func (c *Client) SubmitJob(typeName string, payload []byte, retries int, delay time.Duration) (*asynq.TaskInfo, error) {
	task := asynq.NewTask(typeName, payload)
	return c.client.Enqueue(task, asynq.MaxRetry(retries), asynq.ProcessIn(delay))
}

// Server handles job processing
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer creates a new Asynq server
func NewServer(redisAddr string, concurrency int) *Server {
	return &Server{
		server: asynq.NewServer(
			asynq.RedisClientOpt{Addr: redisAddr},
			asynq.Config{
				Concurrency: concurrency,
				Queues: map[string]int{
					"critical": 6,
					"default":  3,
					"low":      1,
				},
			},
		),
		mux: asynq.NewServeMux(),
	}
}

// RegisterHandler registers a handler for a task type
func (s *Server) RegisterHandler(typeName string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(typeName, handler)
}

// Start runs the worker server
func (s *Server) Start() error {
	return s.server.Run(s.mux)
}

// Shutdown stops the worker server
func (s *Server) Shutdown() {
	s.server.Shutdown()
}
