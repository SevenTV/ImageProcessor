package task

import (
	"context"
	"runtime"
	"time"

	"github.com/seventv/EmoteProcessor/src/global"
	"github.com/seventv/EmoteProcessor/src/job"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func Listen(ctx global.Context) {
	msgCh, err := ctx.Instances().Rmq.Subscribe(ctx.Config().Rmq.JobQueueName)
	if err != nil {
		logrus.Fatal("failed to listen to jobs: ", err)
	}

	maxProcs := runtime.GOMAXPROCS(0)
	workers := make(chan *taskWorker, maxProcs)
	for i := 0; i < maxProcs; i++ {
		workers <- &taskWorker{
			cb: workers,
		}
	}

	for msg := range msgCh {
		worker := <-workers
		go worker.process(ctx, msg)
	}
}

type taskWorker struct {
	cb chan *taskWorker
}

type rmqResult struct {
	JobID   string `json:"job_id"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (w *taskWorker) process(ctx global.Context, msg amqp.Delivery) {
	ctx.AddTask(1)
	defer func() {
		ctx.DoneTask()
		w.cb <- w
	}()

	job := job.Job{}

	err := json.Unmarshal(msg.Body, &job)
	if err != nil {
		logrus.Warn("bad job message: ", err)
		return
	}

	lCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(ctx.Config().MaxTaskDuration))
	defer cancel()

	task := New(lCtx, job)

	task.Start(ctx)

	logrus.Info("starting new task: ", job.ID)

	for event := range task.Events() {
		event.JobID = job.ID
		event, _ := json.Marshal(event)
		if err := ctx.Instances().Rmq.Publish(ctx.Config().Rmq.UpdateQueueName, "application/json", amqp.Transient, event); err != nil {
			logrus.Warn("failed to send update: ", err)
		}
	}
	<-task.Done()
	if err := task.Failed(); err != nil {
		if err := msg.Reject(false); err != nil {
			logrus.Warn("failed to ack: ", err)
		}
		logrus.Errorf("task failed %s: %s", job.ID, err.Error())
	} else {
		if err := msg.Ack(false); err != nil {
			logrus.Warn("failed to ack: ", err)
		}
	}

	errStr := ""
	if task.Failed() != nil {
		errStr = task.Failed().Error()
	}

	resp, _ := json.Marshal(rmqResult{
		JobID:   job.ID,
		Success: task.Failed() == nil,
		Error:   errStr,
	})

	if err := ctx.Instances().Rmq.Publish(ctx.Config().Rmq.ResultQueueName, "application/json", amqp.Persistent, resp); err != nil {
		logrus.Error("failed to ack: ", err)
	}

	logrus.Info("finished task: ", job.ID)
}
