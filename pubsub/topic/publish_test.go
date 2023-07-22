package topic

import (
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
)

type config struct {
	ProjectID string `env:"GCP_PROJECT_ID"`
	TopicID   string `env:"GCP_TOPIC_ID"`
}

var cfg config

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}

	os.Exit(m.Run())
}

func Test_publish(t *testing.T) {
	type args struct {
		projectID string
		topicID   string
		msg       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "publish",
			args: args{
				projectID: cfg.ProjectID,
				topicID:   cfg.TopicID,
				msg:       "Hello World",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := publish(tt.args.projectID, tt.args.topicID, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
