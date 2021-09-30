package uselambda

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type ErrHandler func(error, *Context) (interface{}, error)

var DefaultErrHandler ErrHandler = func(err error, ctx *Context) (interface{}, error) {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError) + ": " + err.Error(),
	}, nil
}

type Handler func(ctx *Context) (interface{}, error)

type Lambda struct {
	handlers   []Handler
	errHandler ErrHandler
}

func (l *Lambda) SetErrHandler(handler ErrHandler) *Lambda {
	l.errHandler = handler
	return l
}

func (l *Lambda) Handle(handler Handler) *Lambda {
	l.handlers = append(l.handlers, handler)
	return l
}

func (l *Lambda) Invoke(parent context.Context, payload []byte) (out []byte, err error) {
	ctx := new(Context)
	ctx.index = -1
	ctx.Payload = payload
	ctx.handlers = l.handlers
	ctx.lambda = l
	ctx.parent = parent

	resp, err := ctx.Next()
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp)
}

func (l *Lambda) With(key string, value interface{}) *Lambda {
	l.handlers = append(l.handlers, func(ctx *Context) (interface{}, error) {
		ctx.Set(key, value)
		return ctx.Next()
	})
	return l
}

func (l *Lambda) Use(handlers ...Handler) *Lambda {
	l.handlers = append(l.handlers, handlers...)
	return l
}

func Use(handlers ...Handler) *Lambda {
	api := new(Lambda)
	if len(handlers) >= int(abortIndex)-1 {
		panic("too many handlers")
	}
	api.handlers = handlers
	api.errHandler = DefaultErrHandler
	return api
}
