// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal_test

import (
	"testing"

	"github.com/instana/testify/assert"
	"github.com/mier85/go-sensor/autoprofile/internal"
)

func TestGenerateUUID_Unique(t *testing.T) {
	assert.NotEqual(t, internal.GenerateUUID(), internal.GenerateUUID())
}
