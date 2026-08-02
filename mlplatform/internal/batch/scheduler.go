package batch

import (
	"context"

	"go.temporal.io/sdk/client"
)

type Scheduler struct {
	client    client.Client
	taskQueue string
}

func NewScheduler(hostPort, namespace, taskQueue string) (*Scheduler, error) {
	c, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}
	return &Scheduler{client: c, taskQueue: taskQueue}, nil
}

// ScheduleRecurring sets up cron-like schedules for the batch workflows.
func (s *Scheduler) ScheduleRecurring(ctx context.Context) error {
	type scheduleEntry struct {
		id       string
		workflow interface{}
		cron     string
	}

	schedules := []scheduleEntry{
		{"daily-features", DailyFeatureGenerationWorkflow, "0 2 * * *"},      // 2 AM
		{"nightly-train", NightlyTrainingWorkflow, "0 3 * * *"},              // 3 AM
		{"hourly-agg", HourlyAggregationWorkflow, "0 * * * *"},              // every hour
		{"weekly-drift", WeeklyDriftDetectionWorkflow, "0 6 * * 1"},          // Mon 6 AM
	}

	sc := s.client.ScheduleClient()

	for _, sched := range schedules {
		_, err := sc.Create(ctx, client.ScheduleOptions{
			ID: sched.id,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{sched.cron},
			},
			Action: &client.ScheduleWorkflowAction{
				Workflow:  sched.workflow,
				TaskQueue: s.taskQueue,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) Close() {
	s.client.Close()
}
