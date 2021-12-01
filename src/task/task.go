package task

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	Aws "github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
	"github.com/seventv/ImageProcessor/src/aws"
	"github.com/seventv/ImageProcessor/src/containers"
	"github.com/seventv/ImageProcessor/src/global"
	"github.com/seventv/ImageProcessor/src/image"
	"github.com/seventv/ImageProcessor/src/job"
	"github.com/seventv/ImageProcessor/src/utils"
	"github.com/sirupsen/logrus"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var ErrUnknownJobProvider = fmt.Errorf("unknown job provider")

type Task struct {
	id uuid.UUID

	job job.Job

	mtx       sync.Mutex
	started   bool
	stopped   bool
	completed bool
	failed    error

	dir string

	files []job.File

	events chan TaskEvent

	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, job job.Job) *Task {
	ctx, cancel := context.WithCancel(ctx)
	id, _ := uuid.NewRandom()
	return &Task{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		job:    job,
		events: make(chan TaskEvent, 20),
	}
}

func (t *Task) ID() uuid.UUID {
	return t.id
}

func (t *Task) Start(ctx global.Context) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if t.started || t.stopped || t.completed {
		return
	}

	t.started = true

	go t.start(ctx)
}

func (t *Task) start(ctx global.Context) {
	defer close(t.events)
	defer func() {
		if err := recover(); err != nil && t.failed == nil {
			logrus.Error("panic to cleanup: ", err)
			t.completed = true
			t.failed = fmt.Errorf("%v", err)
			t.cancel()
			t.events <- TaskEvent{
				JobID:     t.job.ID,
				Type:      Failed,
				Timestamp: time.Now(),
			}
		}
		if err := t.cleanup(); err != nil {
			logrus.Error("failed to cleanup: ", err)
		}
	}()

	t.events <- TaskEvent{
		JobID:     t.job.ID,
		Type:      Started,
		Timestamp: time.Now(),
	}

	var (
		err  error
		data []byte
	)

	switch t.job.RawProvider {
	case job.AwsProvider:
		providerDetails := job.RawProviderDetailsAws{}
		if err = json.Unmarshal(t.job.RawProviderDetails, &providerDetails); err != nil {
			goto completed
		}

		buf := Aws.NewWriteAtBuffer([]byte{})
		if err = ctx.Instances().AwsS3.DownloadFile(t.ctx, providerDetails.Bucket, providerDetails.Key, buf); err != nil {
			goto completed
		}

		data = buf.Bytes()
	case job.LocalProvider:
		providerDetails := job.RawProviderDetailsLocal{}
		if err = json.Unmarshal(t.job.RawProviderDetails, &providerDetails); err != nil {
			goto completed
		}

		if data, err = os.ReadFile(providerDetails.Path); err != nil {
			goto completed
		}
	default:
		err = ErrUnknownJobProvider
		goto completed
	}

	t.events <- TaskEvent{
		JobID:     t.job.ID,
		Type:      Downloaed,
		Timestamp: time.Now(),
	}

	if t.ctx.Err() != nil {
		err = t.ctx.Err()
		goto completed
	}

	{
		var imgType image.ImageType

		// we now have to figure out what we have??
		if imgType, err = containers.ToType(data); err != nil {
			goto completed
		}

		dir := path.Join(ctx.Config().WorkingDir, t.id.String())
		if err = os.MkdirAll(dir, 0700); err != nil {
			goto completed
		}

		t.dir = dir

		fileName := path.Join(dir, fmt.Sprintf("raw.%s", imgType))
		if err = os.WriteFile(fileName, data, 0600); err != nil {
			goto completed
		}

		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      StageOne,
			Timestamp: time.Now(),
		}

		aspectRatioXy := [2]int{}
		if len(t.job.AspectRatioXY) == 2 && t.job.AspectRatioXY[0]*t.job.AspectRatioXY[1] != 0 {
			aspectRatioXy[0] = t.job.AspectRatioXY[0]
			aspectRatioXy[1] = t.job.AspectRatioXY[1]
		} else {
			aspectRatioXy[0] = 3
			aspectRatioXy[1] = 1
		}

		var img *image.Image
		if img, err = containers.ProcessStage1(t.ctx, ctx.Config(), fileName, imgType, aspectRatioXy); err != nil {
			goto completed
		}

		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      StageOneComplete,
			Timestamp: time.Now(),
		}

		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      StageTwo,
			Timestamp: time.Now(),
		}

		if err = containers.ProcessStage2(t.ctx, ctx.Config(), img, t.job.Sizes); err != nil {
			goto completed
		}

		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      StageTwoComplete,
			Timestamp: time.Now(),
		}

		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      StageThree,
			Timestamp: time.Now(),
		}

		if t.files, err = containers.ProcessStage3(t.ctx, ctx.Config(), img, t.job.Sizes, t.job.Settings); err != nil {
			goto completed
		}

		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      StageThreeComplete,
			Timestamp: time.Now(),
		}

		files := []string{}
		if err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".webp" {
				files = append(files, path)
			} else if filepath.Ext(path) == ".png" {
				files = append(files, path)
			} else if filepath.Ext(path) == ".gif" {
				files = append(files, path)
			} else if filepath.Ext(path) == ".avif" {
				files = append(files, path)
			}

			return nil
		}); err != nil {
			goto completed
		}

		switch t.job.ResultConsumer {
		case job.AwsConsumer:
			providerDetails := job.ResultConsumerDetailsAws{}
			if err = json.Unmarshal(t.job.ResultConsumerDetails, &providerDetails); err != nil {
				goto completed
			}
			errCh := make(chan error)
			wg := sync.WaitGroup{}
			wg.Add(len(files))
			for _, v := range files {
				go func(v string) {
					defer wg.Done()
					f, err := os.Open(v)
					if err != nil {
						errCh <- err
						return
					}
					defer f.Close()
					errCh <- ctx.Instances().AwsS3.UploadFile(
						t.ctx,
						providerDetails.Bucket,
						path.Join(providerDetails.KeyFolder, path.Base(v)),
						f,
						utils.StringPointer(mime.TypeByExtension(path.Ext(v))),
						aws.AclPublicRead,
						aws.DefaultCacheControl,
					)
				}(v)
			}
			go func() {
				wg.Wait()
				close(errCh)
			}()

			for e := range errCh {
				err = multierror.Append(err, e).ErrorOrNil()
			}
			if err != nil {
				goto completed
			}
		case job.LocalConsumer:
			providerDetails := job.ResultConsumerDetailsLocal{}
			if err = json.Unmarshal(t.job.ResultConsumerDetails, &providerDetails); err != nil {
				goto completed
			}

			err = os.MkdirAll(providerDetails.PathFolder, 0700)
			if err != nil {
				goto completed
			}

			var f []byte
			for _, v := range files {
				f, err = os.ReadFile(v)
				if err != nil {
					goto completed
				}

				err = os.WriteFile(path.Join(providerDetails.PathFolder, path.Base(v)), f, 0600)
				if err != nil {
					goto completed
				}
			}
		}
	}

completed:
	t.completed = true
	t.failed = err
	t.cancel()
	if err != nil {
		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      Failed,
			Timestamp: time.Now(),
		}
	} else {
		t.events <- TaskEvent{
			JobID:     t.job.ID,
			Type:      Completed,
			Timestamp: time.Now(),
		}
	}
}

func (t *Task) Stop() {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	t.events <- TaskEvent{
		JobID:     t.job.ID,
		Type:      Stopped,
		Timestamp: time.Now(),
	}

	t.stopped = true
	t.cancel()
}

func (t *Task) Done() <-chan struct{} {
	return t.ctx.Done()
}

func (t *Task) Events() <-chan TaskEvent {
	return t.events
}

func (t *Task) Files() []job.File {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if !t.completed || t.failed != nil {
		return nil
	}

	return t.files
}

func (t *Task) Completed() bool {
	return t.completed
}

func (t *Task) Failed() error {
	return t.failed
}

func (t *Task) Started() bool {
	return t.started
}

func (t *Task) Stopped() bool {
	return t.stopped
}

func (t *Task) cleanup() error {
	if !t.started {
		return nil
	}

	t.events <- TaskEvent{
		JobID:     t.job.ID,
		Type:      Cleaned,
		Timestamp: time.Now(),
	}

	t.cancel()
	return os.RemoveAll(t.dir)
}

func (t *Task) Job() job.Job {
	return t.job
}
