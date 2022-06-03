module github.com/mier85/go-sensor/instrumentation/instasarama/example

go 1.9

require (
	github.com/Shopify/sarama v1.19.0
	github.com/mier85/go-sensor v1.42.0
	github.com/mier85/go-sensor/instrumentation/instasarama v1.1.0
	github.com/opentracing/opentracing-go v1.1.0
)

replace github.com/mier85/go-sensor => ../../../

replace github.com/mier85/go-sensor/instrumentation/instasarama => ../
