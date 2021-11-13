package global

import (
	"context"
	"io"

	"github.com/streadway/amqp"
)

type Instances struct {
	AwsS3 AwsS3
	Rmq   Rmq
}

type AwsS3 interface {
	UploadFile(ctx context.Context, bucket, key string, data io.Reader, contentType, acl, cacheControl *string) error
	DownloadFile(ctx context.Context, bucket, key string, file io.WriterAt) error
}

type Rmq interface {
	Subscribe(name string) (<-chan amqp.Delivery, error)
	Publish(queue string, contentType string, deliveryMode uint8, msg []byte) error
	Shutdown()
}
