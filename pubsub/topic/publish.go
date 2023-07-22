package topic

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"log"
)

func publish(projectID, topicID, msg string) (string, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("pubsub: NewClient: %w", err)
	}
	defer func(client *pubsub.Client) {
		_ = client.Close()
	}(client)

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
