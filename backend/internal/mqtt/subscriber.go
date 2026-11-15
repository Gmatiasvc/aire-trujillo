package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/cache"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/store"
)

const TelemetryTopicFilter = "aire-trujillo/v1/nodes/+/telemetry"

// Subscriber ingests MQTT telemetry payloads into PostgreSQL + PostGIS.
type Subscriber struct {
	brokerURL string
	repo      store.Repository
	cache     *cache.MemoryCache
	client    paho.Client
}

// NewSubscriber builds an MQTT subscriber wired to repository and cache invalidation.
func NewSubscriber(brokerURL string, repo store.Repository, memCache *cache.MemoryCache) *Subscriber {
	return &Subscriber{
		brokerURL: brokerURL,
		repo:      repo,
		cache:     memCache,
	}
}

// Start connects asynchronously to Mosquitto with automatic reconnect and QoS 1 subscription.
func (s *Subscriber) Start(ctx context.Context) {
	opts := paho.NewClientOptions().
		AddBroker(s.brokerURL).
		SetClientID(fmt.Sprintf("aire-trujillo-backend-%d", time.Now().UnixNano())).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetMaxReconnectInterval(60 * time.Second).
		SetCleanSession(true)

	opts.SetOnConnectHandler(func(c paho.Client) {
		log.Printf("[MQTT] connected to broker %s, subscribing to %s", s.brokerURL, TelemetryTopicFilter)
		tok := c.Subscribe(TelemetryTopicFilter, 1, func(_ paho.Client, msg paho.Message) {
			msgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := s.HandleMessage(msgCtx, msg.Topic(), msg.Payload()); err != nil {
				log.Printf("[MQTT] message rejected on %s: %v", msg.Topic(), err)
			}
		})
		go func() {
			if tok.WaitTimeout(10*time.Second) && tok.Error() != nil {
				log.Printf("[MQTT] subscription error on %s: %v", TelemetryTopicFilter, tok.Error())
			}
		}()
	})

	opts.SetConnectionLostHandler(func(_ paho.Client, err error) {
		log.Printf("[MQTT] connection lost: %v", err)
	})

	s.client = paho.NewClient(opts)
	go func() {
		tok := s.client.Connect()
		if tok.Wait() && tok.Error() != nil {
			log.Printf("[MQTT] initial connect error: %v", tok.Error())
		}
	}()
}

// Close disconnects cleanly from the MQTT broker.
func (s *Subscriber) Close() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
	}
}

// HandleMessage validates and persists a single telemetry message from topic.
func (s *Subscriber) HandleMessage(ctx context.Context, topic string, rawPayload []byte) error {
	topicNodeCode := extractNodeCodeFromTopic(topic)

	var payload domain.MQTTTelemetryPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	nodeCode := strings.TrimSpace(payload.NodeCode)
	if nodeCode == "" {
		nodeCode = topicNodeCode
	}
	if nodeCode == "" {
		return errors.New("missing node_code in topic and payload")
	}
	if topicNodeCode != "" && payload.NodeCode != "" && topicNodeCode != payload.NodeCode {
		return fmt.Errorf("topic node_code %q mismatches payload node_code %q", topicNodeCode, payload.NodeCode)
	}
	if payload.RecordedAt.IsZero() {
		return errors.New("recorded_at is required")
	}
	if !payload.Metrics.HasAny() {
		return errors.New("metrics cannot be empty")
	}
	if err := payload.Metrics.Validate(); err != nil {
		return fmt.Errorf("invalid metrics: %w", err)
	}

	node, err := s.repo.GetNodeByCode(ctx, nodeCode)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			log.Printf("[MQTT] warning: unknown node_code %q, discarding telemetry", nodeCode)
			return nil
		}
		return fmt.Errorf("lookup node %q: %w", nodeCode, err)
	}
	if node.Status != domain.NodeStatusActive {
		log.Printf("[MQTT] warning: inactive node %q, discarding telemetry", nodeCode)
		return nil
	}

	reading := domain.TelemetryReading{
		NodeID:     node.ID,
		RecordedAt: payload.RecordedAt.UTC(),
		Metrics:    payload.Metrics,
		Source:     domain.SourceMQTT,
	}
	if err := s.repo.UpsertReading(ctx, reading); err != nil {
		return fmt.Errorf("persist reading for %q: %w", nodeCode, err)
	}

	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	return nil
}

func extractNodeCodeFromTopic(topic string) string {
	parts := strings.Split(strings.TrimSpace(topic), "/")
	// Expected: aire-trujillo/v1/nodes/{node_code}/telemetry
	if len(parts) == 5 && parts[0] == "aire-trujillo" && parts[1] == "v1" && parts[2] == "nodes" && parts[4] == "telemetry" {
		return parts[3]
	}
	return ""
}
