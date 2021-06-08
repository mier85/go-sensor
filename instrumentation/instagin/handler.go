// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.11

package instagin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
)

// AddMiddleware adds the tracing middleware to the list of Gin handlers. Unlike the gin.Use method, it puts the middleware
// to the beginning of the list to allow tracing default handlers added by the gin.Default() call.
func AddMiddleware(sensor *instana.Sensor, engine *gin.Engine) {
	f := middleware(sensor)
	engine.Handlers = append([]gin.HandlerFunc{f}, engine.Handlers...)

	// trigger engine.rebuild404Handlers and engine.rebuild405Handlers
	engine.Use()
}

type statusWriter interface {
	SetStatus(status int)
}

// middleware wraps gin's handlers execution. Adds tracing context and handles entry span.
func middleware(sensor *instana.Sensor) gin.HandlerFunc {
	return func(gc *gin.Context) {
		instana.TracingHandlerFunc(sensor, gc.FullPath(), func(writer http.ResponseWriter, request *http.Request) {
			gc.Request = request
			gc.Next()

			// set status from gc.Writer to instana.statusCodeRecorder which is used by instana.TracingHandlerFunc
			if v, ok := writer.(statusWriter); ok {
				v.SetStatus(gc.Writer.Status())
			}
		})(gc.Writer, gc.Request)
	}
}
