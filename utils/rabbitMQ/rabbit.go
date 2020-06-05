package rabbit

import (
	env "github.com/Gimulator/hub/utils/environment"
	"github.com/streadway/amqp"
)

type Rabbit struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbit() (*Rabbit, error) {
	r := &Rabbit{}

	conn, err := amqp.Dial(env.RabbitURI())
	if err != nil {
		return nil, err
	}
	r.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	r.ch = ch

	return r, nil
}

func (r *Rabbit) Send(body []byte) error {
	queue, err := r.ch.QueueDeclare(
		env.RabbitQueue(), // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return err
	}

	if err := r.ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return err
	}

	return nil
}
