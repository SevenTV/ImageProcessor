package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/bugsnag/panicwrap"

	"github.com/seventv/EmoteProcessor/src/aws"
	"github.com/seventv/EmoteProcessor/src/configure"
	"github.com/seventv/EmoteProcessor/src/global"
	"github.com/seventv/EmoteProcessor/src/rmq"
	"github.com/seventv/EmoteProcessor/src/task"
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
		logrus.Info("7TV Emote Processor")
		logrus.Infof("Version: %s", Version)
		logrus.Infof("build.Time: %s", Time)
		logrus.Infof("build.User: %s", User)
	}

	logrus.Debug("MaxProcs: ", runtime.GOMAXPROCS(0))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	c, cancel := context.WithCancel(context.Background())

	ctx := global.New(c, config)

	ctx.Instances().Rmq = rmq.New(ctx)
	if ctx.Config().Aws.Region != "" {
		ctx.Instances().AwsS3 = aws.NewS3(ctx)
	}

	go task.Listen(ctx)

	logrus.Info("running")

	done := make(chan struct{})
	go func() {
		<-sig
		cancel()
		go func() {
			select {
			case <-time.After(time.Minute):
			case <-sig:
			}
			logrus.Fatal("force shutdown")
		}()

		logrus.Info("shutting down")

		ctx.Instances().Rmq.Shutdown()

		ctx.Wait()

		close(done)
	}()

	<-done

	logrus.Info("shutdown")
	os.Exit(0)
}
