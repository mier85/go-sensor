// (c) Copyright IBM Corp. 2022

//go:build go1.16
// +build go1.16

package instaredigo

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

type instaRedigoConn struct {
	redis.Conn
	sensor   *instana.Sensor
	address  string
	prevSpan *prevSpan
	mu       sync.Mutex
}

type prevSpan struct {
	span      ot.Span
	batchCmds []string
}

func isTransactionEnding(commandName string) bool {
	return strings.ToUpper(commandName) == "EXEC" || strings.ToUpper(commandName) == "DISCARD"
}

func isTransactionBeginning(commandName string) bool {
	return strings.ToUpper(commandName) == "MULTI"
}

func genericDo(conn *redis.Conn, ctx context.Context, timeout time.Duration, commandName string, args ...interface{}) (interface{}, error) {
	switch c := (*conn).(type) {
	case redis.ConnWithContext:
		return c.DoContext(ctx, commandName, args...)
	case redis.ConnWithTimeout:
		return c.DoWithTimeout(timeout, commandName, args...)
	// case redis.Conn:
	default:
		return c.Do(commandName, args...)
	}
}

func handleMultiTransaction(c *instaRedigoConn, commandName string, 
    ctx context.Context, conn redis.Conn, timeout time.Duration, 
    args ...interface{}) (reply interface{}, err error){
        reply, err = genericDo(&conn, ctx, time.Millisecond, commandName, args)
		if err != nil {
			c.prevSpan.span.SetTag("redis.error", err.Error())
			c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
		}

		if !isTransactionEnding(commandName) {
			c.prevSpan.batchCmds = append(c.prevSpan.batchCmds, commandName)
			c.prevSpan.span.SetTag("redis.subCommands", c.prevSpan.batchCmds)
		} else {
			c.prevSpan.span.Finish()
			c.prevSpan = nil
		}
        return reply, err
}

func genericHandler(c *instaRedigoConn, commandName string, ctx context.Context, conn redis.Conn, timeout time.Duration, args ...interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

    //Handle multi statement transaction
	if c.prevSpan != nil {
	    return handleMultiTransaction(c, commandName, ctx, conn, timeout, args...)	
	} else {
		//Fetching the tracer
		tracer := c.sensor.Tracer()
		opts := []ot.StartSpanOption{
			ot.Tags{
				"redis.connection": c.address,
			},
		}

		//Fetching the span from the context if it's there and starting it
		if ps, ok := instana.SpanFromContext(ctx); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}
		span := tracer.StartSpan("redis", opts...)

		//Checking whether it is a multi command
		if isTransactionBeginning(commandName) {
			c.prevSpan = &prevSpan{span, nil}
		} else {
			defer span.Finish()
		}

		//Setting span tags and executing the command
		span.SetTag("redis.command", commandName)
		reply, err := genericDo(&conn, ctx, time.Millisecond, commandName, args)
		if err != nil {
			span.SetTag("redis.error", err.Error())
			span.LogFields(otlog.Object("error", err.Error()))
		}

		return reply, err
	}
}

// Do sends a command to the server and returns the received reply and collect
//the instrumentation details. Inorder to capture the correlated span information,
//create a context from the parent span and pass it as an argument along with
//the other arguments. The Do API will retrieve the span information from the
//context and record the correlated span information.
func (c *instaRedigoConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	//Separating the passed context and the arguments
	var cmdArgs []interface{}
	ctx := context.Background()
	for _, a := range args {
		if _, ok := a.(context.Context); ok {
			ctx = a.(context.Context)
		} else {
			cmdArgs = append(cmdArgs, a)
		}
	}

	return genericHandler(c, commandName, ctx, c.Conn, time.Microsecond, args)
}

//DoContext sends a command to server and returns the received reply along with
//the instrumentation.
func (c *instaRedigoConn) DoContext(ctx context.Context, commandName string,
	args ...interface{}) (reply interface{}, err error) {
	if cwc, ok := c.Conn.(redis.ConnWithContext); ok {
		return genericHandler(c, commandName, ctx, cwc, time.Microsecond, args)
	}

	return nil, errors.New("redis: connection does not support ConnWithContext")
}

//DoWithTimeout executes a Redis command with the specified read timeout along
//with the instrumentation. If the connection does not satisfy the ConnWithTimeout
//interface, then an error is returned.
func (c *instaRedigoConn) DoWithTimeout(timeout time.Duration, commandName string,
	args ...interface{}) (reply interface{}, err error) {
	
        if cwt, ok := c.Conn.(redis.ConnWithTimeout); ok {
		//return cwt.DoWithTimeout(timeout, commandName, args...)
		var cmdArgs []interface{}
		ctx := context.Background()
		//Separating the passed context and the arguments
		for _, a := range args {
			if _, ok := a.(context.Context); ok {
				ctx = a.(context.Context)
			} else {
				cmdArgs = append(cmdArgs, a)
			}
		}

		return genericHandler(c, commandName, ctx, cwt, timeout, args)
	}

	return nil, errors.New("redis: connection does not support ConnWithTimeout")
}

//Dial connects to the Redis server at the given network and address using the
//specified options along with instrumentation code.
func Dial(sensor *instana.Sensor, network, address string, options ...redis.DialOption) (redis.Conn, error) {
	conn, err := redis.Dial(network, address, options...)
    if err != nil {
		return conn, err
	}
	if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}

	return &instaRedigoConn{Conn: conn, sensor: sensor, address: address, prevSpan: nil}, err
}

//DialURL wraps DialURLContext using context.Background along with the instrumentation code.
func DialURL(sensor *instana.Sensor, rawurl string, options ...redis.DialOption) (redis.Conn, error) {
	conn, err := redis.DialURL(rawurl, options...)
	if err != nil {
		return conn, err
	}

	return &instaRedigoConn{Conn: conn, sensor: sensor, address: rawurl, prevSpan: nil}, err
}

//DialURLContext connects to a Redis server at the given URL using the Redis
//URI scheme along with the instrumentation code.
func DialURLContext(sensor *instana.Sensor, ctx context.Context, rawurl string, options ...redis.DialOption) (redis.Conn, error) {
	conn, err := redis.DialURLContext(ctx, rawurl, options...)
	if err != nil {
		return conn, err
	}

	return &instaRedigoConn{Conn: conn, sensor: sensor, address: rawurl, prevSpan: nil}, err
}

//NewConn returns a new Redigo connection for the given net connection along with the instrumentation code.
func NewConn(sensor *instana.Sensor, netConn net.Conn, readTimeout, writeTimeout time.Duration) redis.Conn {
	addr := netConn.LocalAddr().String()
	conn := redis.NewConn(netConn, readTimeout, writeTimeout)

	return &instaRedigoConn{Conn: conn, sensor: sensor, address: addr, prevSpan: nil}
}

// Send writes the command to the client's output buffer and collect the
//instrumentation details.Inorder to capture the correlated span information,
//create a context from the parent span and pass it as an argument along with
//the other arguments. The Send API will retrieve the span information from the
//context and record the correlated span information.
func (c *instaRedigoConn) Send(commandName string, args ...interface{}) (err error) {
	var cmdArgs []interface{}
	//Separating the parent context from the arguments
	ctx := context.Background()
	for _, a := range args {
		if _, ok := a.(context.Context); ok {
			ctx = a.(context.Context)
		} else {
			cmdArgs = append(cmdArgs, a)
		}
	}

    //Handling the command if there exists a previous span
    c.mu.Lock()
    defer c.mu.Unlock()

	if c.prevSpan != nil {

		err = c.Conn.Send(commandName, cmdArgs...)
		if err != nil {
			c.prevSpan.span.SetTag("redis.error", err.Error())
			c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
		}

		if !isTransactionEnding(commandName) {
			c.prevSpan.batchCmds = append(c.prevSpan.batchCmds, commandName)
			c.prevSpan.span.SetTag("redis.subCommands", c.prevSpan.batchCmds)
		} else {
			c.prevSpan.span.Finish()
			c.prevSpan = nil
		}
		
	} else {
	
        //Fetching the tracer
        tracer := c.sensor.Tracer()
		opts := []ot.StartSpanOption{
			ot.Tags{
				"redis.connection": c.address,
			},
		}
	
        //Fetching the span and starting it
        if ps, ok := instana.SpanFromContext(ctx); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}
		span := tracer.StartSpan("redis", opts...)
        
        //Checking for multi command transaction
        if isTransactionBeginning(commandName) {
			c.prevSpan = &prevSpan{span, nil}
		} else {
            defer span.Finish()
        }

        //Setting the span tags and executing the command
		span.SetTag("redis.command", commandName)
		err = c.Conn.Send(commandName, cmdArgs...)
		if err != nil {
			span.SetTag("redis.error", err.Error())
			span.LogFields(otlog.Object("error", err.Error()))
		}
	}

	return err
}
