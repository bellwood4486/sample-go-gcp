package topic

import (
	"context"
	"fmt"
	"log"

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
