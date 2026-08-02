package eventbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// KafkaBus is an alternative eventbus implementation backed by Apache Kafka.
type KafkaBus struct {
	brokers []string
	groupID string
	writers map[string]*kafka.Writer
	readers map[string]*kafka.Reader
}

func NewKafka(brokers []string, groupID string) (*KafkaBus, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("no kafka brokers configured")
	}
	return &KafkaBus{
		brokers: brokers,
		groupID: groupID,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
	}, nil
}

func (b *KafkaBus) writer(subject string) *kafka.Writer {
	if w, ok := b.writers[subject]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(b.brokers...),
		Topic:        subject,
		RequiredAcks: kafka.RequireAll,
	}
	b.writers[subject] = w
	return w
}

func (b *KafkaBus) Publish(ctx context.Context, subject string, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return b.writer(subject).WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.ID),
		Value: data,
	})
}

func (b *KafkaBus) Subscribe(ctx context.Context, subject string, group string) (<-chan *Event, error) {
	ch := make(chan *Event, 256)

	key := subject + ":" + group
	r, ok := b.readers[key]
	if !ok {
		r = kafka.NewReader(kafka.ReaderConfig{
			Brokers: b.brokers,
			GroupID: group,
			Topic:   subject,
		})
		b.readers[key] = r
	}

	go func() {
		defer close(ch)
		for {
			m, err := r.FetchMessage(ctx)
			if err != nil {
				return
			}
			var evt Event
			if json.Unmarshal(m.Value, &evt) == nil {
				ch <- &evt
			}
			r.CommitMessages(ctx, m)
		}
	}()

	return ch, nil
}

func (b *KafkaBus) Close() error {
	var errs []error
	for _, w := range b.writers {
		if err := w.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	for _, r := range b.readers {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("kafka close errors: %v", errs)
	}
	return nil
}
