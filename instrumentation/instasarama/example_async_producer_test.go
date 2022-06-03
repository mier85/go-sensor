// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama_test

import (
	"github.com/Shopify/sarama"
	instana "github.com/mier85/go-sensor"
	"github.com/mier85/go-sensor/instrumentation/instasarama"
	"github.com/opentracing/opentracing-go/ext"
)

// This example demonstrates how to instrument an async Kafka producer using instasarama.
// Error handling is omitted for brevity.
func Example_asyncProducer() {
	sensor := instana.NewSensor("my-service")
	brokers := []string{"localhost:9092"}

	config := sarama.NewConfig()
	// enable the use record headers added in kafka v0.11.0 and used to propagate
	// trace context
	config.Version = sarama.V0_11_0_0

	// create a new instrumented instance of sarama.SyncProducer
	producer, _ := instasarama.NewAsyncProducer(brokers, config, sensor)

	// start a new entry span
	sp := sensor.Tracer().StartSpan("my-producing-method")
	ext.SpanKind.Set(sp, "entry")

	msg := &sarama.ProducerMessage{
		// ...
	}

	// inject the span before passing the message to producer
	producer.Input() <- instasarama.ProducerMessageWithSpan(msg, sp)
}
