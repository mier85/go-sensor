// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"testing"

	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	instana "github.com/mier85/go-sensor"
)

func TestSpanFromContext_WithActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	span := tracer.StartSpan("test")
	ctx := instana.ContextWithSpan(context.Background(), span)

	sp, ok := instana.SpanFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, span, sp)
}

func TestSpanFromContext_NoActiveSpan(t *testing.T) {
	_, ok := instana.SpanFromContext(context.Background())
	assert.False(t, ok)
}
