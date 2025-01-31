// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build go1.11
// +build go1.11

package pubsub_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	instana "github.com/mier85/go-sensor"
	"github.com/mier85/go-sensor/instrumentation/cloud.google.com/go/pubsub"
	"github.com/opentracing/opentracing-go"
)

func TestTracingHandler(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)

	payload, err := ioutil.ReadFile("testdata/message.json")
	require.NoError(t, err)

	var (
		numCalls int
		reqSpan  opentracing.Span
	)
	h := pubsub.TracingHandlerFunc(sensor, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		var ok bool
		reqSpan, ok = instana.SpanFromContext(req.Context())
		require.True(t, ok)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(payload), string(body))

		w.WriteHeader(http.StatusAccepted)
	})

	rec := httptest.NewRecorder()

	h(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload)))

	assert.Equal(t, http.StatusAccepted, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	_ = reqSpan

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	gcpsSpan := spans[0]

	// new trace started
	assert.NotEmpty(t, gcpsSpan.TraceID)
	assert.Empty(t, gcpsSpan.ParentID)
	assert.NotEmpty(t, gcpsSpan.SpanID)

	// span tags
	assert.Equal(t, "gcps", gcpsSpan.Name)
	assert.EqualValues(t, instana.EntrySpanKind, gcpsSpan.Kind)
	assert.Equal(t, 0, gcpsSpan.Ec)

	require.IsType(t, instana.GCPPubSubSpanData{}, gcpsSpan.Data)

	data := gcpsSpan.Data.(instana.GCPPubSubSpanData)
	assert.Equal(t, instana.GCPPubSubSpanTags{
		Operation:    "CONSUME",
		ProjectID:    "myproject",
		Subscription: "mysubscription",
		MessageID:    "136969346945",
	}, data.Tags)
}

func TestTracingHandlerFunc_TracePropagation(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)

	payload, err := ioutil.ReadFile("testdata/message_with_context.json")
	require.NoError(t, err)

	var numCalls int
	h := pubsub.TracingHandlerFunc(sensor, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		_, ok := instana.SpanFromContext(req.Context())
		assert.True(t, ok)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(payload), string(body))

		w.WriteHeader(http.StatusAccepted)
	})

	rec := httptest.NewRecorder()

	h(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload)))

	assert.Equal(t, http.StatusAccepted, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	gcpsSpan := spans[0]

	// trace continuation
	assert.EqualValues(t, 0x1234, gcpsSpan.TraceID)
	assert.EqualValues(t, 0x5678, gcpsSpan.ParentID)
	assert.NotEmpty(t, gcpsSpan.SpanID)

	// span tags
	assert.Equal(t, "gcps", gcpsSpan.Name)
	assert.EqualValues(t, instana.EntrySpanKind, gcpsSpan.Kind)
	assert.Equal(t, 0, gcpsSpan.Ec)

	require.IsType(t, instana.GCPPubSubSpanData{}, gcpsSpan.Data)

	data := gcpsSpan.Data.(instana.GCPPubSubSpanData)
	assert.Equal(t, instana.GCPPubSubSpanTags{
		Operation:    "CONSUME",
		ProjectID:    "myproject",
		Subscription: "mysubscription",
		MessageID:    "136969346945",
	}, data.Tags)
}

func TestTracingHandlerFunc_NotPubSub(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)

	var numCalls int
	h := pubsub.TracingHandlerFunc(sensor, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		_, ok := instana.SpanFromContext(req.Context())
		assert.True(t, ok)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Equal(t, "request payload", string(body))

		w.WriteHeader(http.StatusAccepted)
	})

	rec := httptest.NewRecorder()

	h(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("request payload")))

	assert.Equal(t, http.StatusAccepted, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "g.http", spans[0].Name)
}
