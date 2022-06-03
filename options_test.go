// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	"github.com/instana/testify/assert"
	instana "github.com/mier85/go-sensor"
)

func TestDefaultOptions(t *testing.T) {
	assert.Equal(t, &instana.Options{
		AgentHost:                   "localhost",
		AgentPort:                   42699,
		MaxBufferedSpans:            instana.DefaultMaxBufferedSpans,
		ForceTransmissionStartingAt: instana.DefaultForceSpanSendAt,
		Tracer:                      instana.DefaultTracerOptions(),
	}, instana.DefaultOptions())
}
