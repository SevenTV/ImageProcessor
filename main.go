package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bugsnag/panicwrap"
	"github.com/davecgh/go-spew/spew"

	"github.com/seventv/ImageProcessor/src/aws"
	"github.com/seventv/ImageProcessor/src/configure"
	"github.com/seventv/ImageProcessor/src/global"
	"github.com/seventv/ImageProcessor/src/job"
	"github.com/seventv/ImageProcessor/src/rmq"
	"github.com/seventv/ImageProcessor/src/task"
	"github.com/sirupsen/logrus"
)

var (
	Version = "development"
	Unix    = ""
	Time    = "unknown"
	User    = "unknown"
)

func init() {
	debug.SetGCPercent(2000)
	if i, err := strconv.Atoi(Unix); err == nil {
		Time = time.Unix(int64(i), 0).Format(time.RFC3339)
	}
}

func main() {
	config := configure.New()

	exitStatus, err := panicwrap.BasicWrap(func(s string) {
		logrus.Error(s)
	})
	if err != nil {
		logrus.Error("failed to setup panic handler: ", err)
		os.Exit(2)
	}

	if exitStatus >= 0 {
		os.Exit(exitStatus)
	}

	if !config.NoHeader {
		logrus.Info("7TV Image Processor")
		logrus.Infof("Version: %s", Version)
		logrus.Infof("build.Time: %s", Time)
		logrus.Infof("build.User: %s", User)
	}

	logrus.Debug("MaxProcs: ", runtime.GOMAXPROCS(0))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	c, cancel := context.WithCancel(context.Background())

	ctx := global.New(c, config)

	done := make(chan struct{})
	go func() {
		select {
		case <-sig:
		case <-ctx.Done():
		}
		cancel()
		go func() {
			select {
			case <-time.After(time.Minute):
			case <-sig:
			}
			logrus.Fatal("force shutdown")
		}()

		logrus.Info("shutting down")

		if ctx.Instances().Rmq != nil {
			ctx.Instances().Rmq.Shutdown()
		}

		ctx.Wait()

		close(done)
	}()

	if config.Input != "" && config.Output != "" {
		rawDetails, _ := json.Marshal(job.RawProviderDetailsLocal{
			Path: config.Input,
		})
		resultDetails, _ := json.Marshal(job.ResultConsumerDetailsLocal{
			PathFolder: config.Output,
		})

		ar := strings.Split(config.AspectRatio, ":")
		if len(ar) != 2 {
			logrus.Fatal("invalid aspect ratio: ", config.AspectRatio)
		}

		arXY := make([]int, 2)

		arXY[0], err = strconv.Atoi(ar[0])
		if err != nil {
			logrus.Fatal("invalid aspect ratio: ", config.AspectRatio)
		}
		arXY[1], err = strconv.Atoi(ar[1])
		if err != nil {
			logrus.Fatal("invalid aspect ratio: ", config.AspectRatio)
		}

		sizes := map[string]job.ImageSize{}

		for _, v := range config.Sizes {
			splits := strings.Split(v, ":")
			if len(splits) != 3 {
				logrus.Fatal("invalid size: ", v)
			}
			size := job.ImageSize{}
			size.Width, err = strconv.Atoi(splits[1])
			if err != nil {
				logrus.Fatal("invalid size: ", config.AspectRatio)
			}
			size.Height, err = strconv.Atoi(splits[2])
			if err != nil {
				logrus.Fatal("invalid size: ", config.AspectRatio)
			}

			sizes[splits[0]] = size
		}

		if len(sizes) == 0 {
			logrus.Fatal("no sizes specified")
		}

		job := job.Job{
			ID: "custom-task",

			AspectRatioXY: arXY,
			Settings:      job.AllSettings,
			Sizes:         sizes,

			RawProvider:           job.LocalProvider,
			RawProviderDetails:    rawDetails,
			ResultConsumer:        job.LocalConsumer,
			ResultConsumerDetails: resultDetails,
		}

		task := task.New(c, job)

		task.Start(ctx)
		for event := range task.Events() {
			spew.Dump(event)
		}
		<-task.Done()
		if task.Failed() != nil {
			logrus.Fatal(task.Failed())
		}

		spew.Dump(task.Files())

		cancel()

	} else {
		ctx.Instances().Rmq = rmq.New(ctx)
		if ctx.Config().Aws.Region != "" {
			ctx.Instances().AwsS3 = aws.NewS3(ctx)
		}

		go task.Listen(ctx)

		logrus.Info("running")
	}

	<-done

	logrus.Info("shutdown")
	os.Exit(0)
}
