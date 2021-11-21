package rmq

import (
	"time"

	"github.com/seventv/ImageProcessor/src/global"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RmqInstance struct {
	rmq   *amqp.Connection
	chRmq *amqp.Channel
}

func New(ctx global.Context) global.Rmq {
	rmq, err := amqp.Dial(ctx.Config().Rmq.ServerURL)
	if err != nil {
		logrus.Fatal("failed to connect to rmq: ", err)
	}

	chRmq, err := rmq.Channel()
	if err != nil {
		logrus.Fatal("failed to connect to rmq: ", err)
	}

	_, err = chRmq.QueueDeclare(
		ctx.Config().Rmq.JobQueueName, // queue name
		true,                          // durable
		false,                         // auto delete
		false,                         // exclusive
		false,                         // no wait
		nil,                           // arguments
	)
	if err != nil {
		logrus.Fatal("failed to connect to rmq: ", err)
	}

	_, err = chRmq.QueueDeclare(
		ctx.Config().Rmq.ResultQueueName, // queue name
		true,                             // durable
		false,                            // auto delete
		false,                            // exclusive
		false,                            // no wait
		nil,                              // arguments
	)
	if err != nil {
		logrus.Fatal("failed to connect to rmq: ", err)
	}

	_, err = chRmq.QueueDeclare(
		ctx.Config().Rmq.UpdateQueueName, // queue name
		true,                             // durable
		false,                            // auto delete
		false,                            // exclusive
		false,                            // no wait
		nil,                              // arguments
	)
	if err != nil {
		logrus.Fatal("failed to connect to rmq: ", err)
	}

	return &RmqInstance{
		rmq:   rmq,
		chRmq: chRmq,
	}
}

func (r *RmqInstance) Subscribe(queue string) (<-chan amqp.Delivery, error) {
	return r.chRmq.Consume(
		queue, // queue name
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,   // arguments
	)
}

func (r *RmqInstance) Publish(queue string, contentType string, deliveryMode uint8, msg []byte) error {
	return r.chRmq.Publish(
		"",    // exchange
		queue, // queue name
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  contentType,
			DeliveryMode: deliveryMode,
			Timestamp:    time.Now(),
			Body:         msg,
			Priority:     0,
		}, // message to publish
	)
}

func (r *RmqInstance) Shutdown() {
	_ = r.rmq.Close()
}
