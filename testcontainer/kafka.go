package testcontainer

import (
	"context"
	"crypto/rand"
	"log"
	"log/slog"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	Image = "confluentinc/cp-kafka:7.8.0"
)

// StartKafkaContainer starts a Kafka container and returns its bootstrap servers.
// It also returns a cleanup function to stop the container.
func StartKafkaContainer(ctx context.Context) *kafka.KafkaContainer {
	slog.Info("Starting Kafka container...")

	kafkaContainer, err := kafka.Run(ctx, Image,
		kafka.WithClusterID(rand.Text()),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("9093/tcp")),
		testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
			if hostConfig.NetworkMode == "bridge" {
				hostConfig.NetworkMode = ""
			}
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	bootstrapServers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Kafka container started", slog.String("bootstrapServers", strings.Join(bootstrapServers, ",")))
	return kafkaContainer
}
