Instana instrumentation for gin framework
=============================================

This module contains middleware to instrument HTTP services written with [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin).

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/mier85/go-sensor/instrumentation/instagin
```

Usage
-----

```go
// init gin engine
engine := gin.Default()

// create a sensor
sensor := instana.NewSensor("gin-sensor")

// add middleware to the gin handlers
instagin.AddMiddleware(sensor, engine)

// define API
engine.GET("/api", func(c *gin.Context) {}
...
```
[Full example][fullExample]



[godoc]: https://pkg.go.dev/github.com/mier85/go-sensor/instrumentation/instagin
[fullExample]: https://github.com/mier85/go-sensor/blob/master/example/gin/main.go
