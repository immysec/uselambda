package uselambda

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

const abortIndex int8 = math.MaxInt8 >> 1

const (
	charsetUTF8 = "charset=UTF-8"
)

const (
	HeaderContentType = "Content-Type"
)

const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMETextPlain                  = "text/plain"
	MIMETextPlainCharsetUTF8       = MIMETextPlain + "; " + charsetUTF8
)

type H map[string]interface{}

type Payload []byte

func (p Payload) MustUnmarshal(dst interface{}) interface{} {
	if err := json.Unmarshal(p, dst); err != nil {
		log.Fatalf("Payload.MustUnmarshal: %v", err)
	}
	return dst
}

func (p Payload) AsRequest() *events.APIGatewayProxyRequest {
	return p.MustUnmarshal(new(events.APIGatewayProxyRequest)).(*events.APIGatewayProxyRequest)
}

func (p Payload) AsWsRequest() *events.APIGatewayWebsocketProxyRequest {
	return p.MustUnmarshal(new(events.APIGatewayWebsocketProxyRequest)).(*events.APIGatewayWebsocketProxyRequest)
}

type Context struct {
	Payload  Payload
	index    int8
	parent   context.Context
	handlers []Handler
	lambda   *Lambda
	mu       sync.RWMutex
	Keys     map[string]interface{}
}

func (ctx *Context) Next() (res interface{}, err error) {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		if res, err = ctx.handlers[ctx.index](ctx); err != nil {
			return ctx.lambda.errHandler(err, ctx)
		}
		ctx.index++
	}
	return
}

func (ctx *Context) JSON(statusCode int, body interface{}) (interface{}, error) {
	buf := bytes.NewBuffer(nil)
	json.NewEncoder(buf).Encode(body)
	res := &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			HeaderContentType: MIMEApplicationJSONCharsetUTF8,
		},
		Body: buf.String(),
	}
	return ctx.Return(res)
}

func (ctx *Context) Base64(statusCode int, body []byte) (interface{}, error) {
	res := &events.APIGatewayProxyResponse{
		IsBase64Encoded: true,
		StatusCode:      statusCode,
		Headers: map[string]string{
			HeaderContentType: MIMETextPlainCharsetUTF8,
		},
		Body: base64.StdEncoding.EncodeToString(body),
	}
	return ctx.Return(res)
}

func (ctx *Context) String(statusCode int, body string) (interface{}, error) {
	res := &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			HeaderContentType: MIMETextPlainCharsetUTF8,
		},
		Body: body,
	}
	return ctx.Return(res)
}

func (ctx *Context) Return(res interface{}) (interface{}, error) {
	ctx.Abort()
	return res, nil
}

func (ctx *Context) Abort() {
	ctx.index = abortIndex
}

func (ctx *Context) IsAborted() bool {
	return ctx.index >= abortIndex
}

func (ctx *Context) Set(key string, value interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.Keys == nil {
		ctx.Keys = make(map[string]interface{})
	}
	ctx.Keys[key] = value
}

func (ctx *Context) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	value, exists := ctx.Keys[key]
	return value, exists
}

func (ctx *Context) MustGet(key string) interface{} {
	value, exists := ctx.Get(key)
	if exists {
		return value
	}
	panic(fmt.Sprintf(`Key "%s" does not exist`, key))
}

func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.parent.Deadline()
}

func (ctx *Context) Done() <-chan struct{} {
	return ctx.parent.Done()
}

func (ctx *Context) Err() error {
	return ctx.parent.Err()
}

func (ctx *Context) Value(key interface{}) interface{} {
	if ks, ok := key.(string); ok {
		if value, exists := ctx.Get(ks); exists {
			return value
		}
	}
	return ctx.parent.Value(key)
}
