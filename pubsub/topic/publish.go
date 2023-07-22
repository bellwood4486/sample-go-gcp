package topic

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"io"
)

func publish(w io.Writer, projectID, topicID, msg string) error {
	// projectID := "my-project-id"
	// topicID := "my-topic"
	// msg := "Hello World"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub: NewClient: %w", err)
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
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	_, _ = fmt.Fprintf(w, "Published a message; msg ID: %v\n", id)

	return nil
}
