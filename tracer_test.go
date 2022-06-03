// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"testing"

	"github.com/instana/testify/assert"
	instana "github.com/mier85/go-sensor"
	ot "github.com/opentracing/opentracing-go"
)

func TestTracerAPI(t *testing.T) {
	tracer := instana.NewTracer()
	assert.NotNil(t, tracer)

	recorder := instana.NewTestRecorder()

	tracer = instana.NewTracerWithEverything(&instana.Options{}, recorder)
	assert.NotNil(t, tracer)

	tracer = instana.NewTracerWithOptions(&instana.Options{})
	assert.NotNil(t, tracer)
}

func TestTracerBasics(t *testing.T) {
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("test")
	sp.SetBaggageItem("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
}

func TestTracer_StartSpan_SuppressTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test", instana.SuppressTracing())

	sc := sp.Context().(instana.SpanContext)
	assert.True(t, sc.Suppressed)
}

func TestTracer_StartSpan_WithCorrelationData(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test", ot.ChildOf(instana.SpanContext{
		Correlation: instana.EUMCorrelationData{
			Type: "type1",
			ID:   "id1",
		},
	}))

	sc := sp.Context().(instana.SpanContext)
	assert.Equal(t, instana.EUMCorrelationData{}, sc.Correlation)
}
