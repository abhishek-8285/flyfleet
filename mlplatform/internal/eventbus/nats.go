package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/mlplatform/internal/metrics"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
)

type NATSBus struct {
	conn    *nats.Conn
	js      jetstream.JetStream
	stream  string
	metrics *metrics.Metrics
}

func NewNATS(url, stream string, m *metrics.Metrics) (*NATSBus, error) {
	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Warn().Err(err).Msg("nats disconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	// Ensure the stream exists.
	_, err = js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:      stream,
		Subjects:  []string{stream + ".>"},
		Storage:   jetstream.FileStorage,
		Replicas:  1,
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}

	return &NATSBus{conn: nc, js: js, stream: stream, metrics: m}, nil
}

func (b *NATSBus) Publish(ctx context.Context, subject string, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		b.metrics.BusPublishErrors.WithLabelValues(subject).Inc()
		return err
	}
	_, err = b.js.Publish(ctx, b.stream+"."+subject, data)
	if err != nil {
		b.metrics.BusPublishErrors.WithLabelValues(subject).Inc()
		return err
	}
	return nil
}

func (b *NATSBus) Subscribe(ctx context.Context, subject string, group string) (<-chan *Event, error) {
	ch := make(chan *Event, 256)

	consumer, err := b.js.CreateOrUpdateConsumer(ctx, b.stream, jetstream.ConsumerConfig{
		Durable:       group,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: b.stream + "." + subject,
	})
	if err != nil {
		return nil, err
	}

	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		var evt Event
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			log.Error().Err(err).Msg("unmarshal event")
			msg.Nak()
			return
		}
		ch <- &evt
		msg.Ack()
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		cc.Stop()
		close(ch)
	}()

	return ch, nil
}

func (b *NATSBus) Close() error {
	b.conn.Close()
	return nil
}
