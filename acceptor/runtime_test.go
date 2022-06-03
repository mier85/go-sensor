// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor_test

import (
	"testing"

	"github.com/instana/testify/assert"
	"github.com/mier85/go-sensor/acceptor"
)

func TestNewGoProcessPluginPayload(t *testing.T) {
	data := acceptor.GoProcessData{
		PID: 42,
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.golang",
		EntityID: "42",
		Data:     data,
	}, acceptor.NewGoProcessPluginPayload(data))
}
