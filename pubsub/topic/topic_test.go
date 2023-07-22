package topic

import (
	"context"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type config struct {
	ProjectID string `env:"GCP_PROJECT_ID"`
	TopicID   string `env:"GCP_TOPIC_ID"`
}

var (
	cfg    config
	client *pubsub.Client
)

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}

	var err error
	if client, err = pubsub.NewClient(context.Background(), cfg.ProjectID); err != nil {
		log.Fatalf("failed to create pubsub client: %v", err)
	}
	defer func(client *pubsub.Client) {
		_ = client.Close()
	}(client)

	os.Exit(m.Run())
}

func Test_publish(t *testing.T) {
	type args struct {
		client  *pubsub.Client
		topicID string
		msg     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "publish",
			args: args{
				client:  client,
				topicID: cfg.TopicID,
				msg:     "Hello World",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := publish(tt.args.client, tt.args.topicID, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
