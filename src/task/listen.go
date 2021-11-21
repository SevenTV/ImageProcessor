package task

import (
	"context"
	"runtime"
	"time"

	"github.com/seventv/ImageProcessor/src/global"
	"github.com/seventv/ImageProcessor/src/job"
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

type RmqResult struct {
	JobID   string     `json:"job_id"`
	Success bool       `json:"success"`
	Files   []job.File `json:"files"`
	Error   string     `json:"error"`
}

func (w *taskWorker) process(ctx global.Context, msg amqp.Delivery) {
	ctx.AddTask(1)
	defer func() {
		ctx.DoneTask()
		w.cb <- w
	}()

	j := job.Job{}

	err := json.Unmarshal(msg.Body, &j)
	if err != nil {
		logrus.Warn("bad job message: ", err)
		return
	}

	if len(j.AspectRatioXY) == 0 {
		j.AspectRatioXY = []int{3, 1}
	}

	if j.Settings == 0 {
		j.Settings = job.AllSettings
	}

	if len(j.Sizes) == 0 {
		j.Sizes = map[string]job.ImageSize{
			"4x": {
				Width:  384,
				Height: 128,
			},
			"3x": {
				Width:  288,
				Height: 96,
			},
			"2x": {
				Width:  192,
				Height: 64,
			},
			"1x": {
				Width:  96,
				Height: 32,
			},
		}
	}

	lCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(ctx.Config().MaxTaskDuration))
	defer cancel()

	task := New(lCtx, j)

	task.Start(ctx)

	logrus.Info("starting new task: ", j.ID)

	for event := range task.Events() {
		event.JobID = j.ID
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
		logrus.Errorf("task failed %s: %s", j.ID, err.Error())
	} else {
		if err := msg.Ack(false); err != nil {
			logrus.Warn("failed to ack: ", err)
		}
	}

	errStr := ""
	if task.Failed() != nil {
		errStr = task.Failed().Error()
	}

	resp, _ := json.Marshal(RmqResult{
		JobID:   j.ID,
		Success: task.Failed() == nil,
		Error:   errStr,
		Files:   task.Files(),
	})

	if err := ctx.Instances().Rmq.Publish(ctx.Config().Rmq.ResultQueueName, "application/json", amqp.Persistent, resp); err != nil {
		logrus.Error("failed to ack: ", err)
	}

	logrus.Info("finished task: ", j.ID)
}
