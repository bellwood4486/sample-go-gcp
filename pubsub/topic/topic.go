package topic

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
)

func publish(client *pubsub.Client, topicID, msg string) (string, error) {
	ctx := context.Background()
	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("pubsub: result.Get: %w", err)
	}
	log.Printf("Published a message; msg ID: %v\n", id)

	return id, nil
}

func pullMsgs(client *pubsub.Client, subID string) error {
	sub := client.Subscription(subID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var received int32
	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		atomic.AddInt32(&received, 1)
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %w", err)
	}
	log.Printf("Received %d messages.\n", received)

	return nil
}
