package uselambda

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestFallthroughMiddlewares(t *testing.T) {
	respBody := "hello"

	wg := new(sync.WaitGroup)

	var seq []int
	middlewares := []Handler{
		func(ctx *Context) (interface{}, error) {
			require.NotNil(t, ctx.Payload)
			wg.Done()
			seq = append(seq, 1)
			return ctx.Next()
		},
		func(ctx *Context) (interface{}, error) {
			require.NotNil(t, ctx.Payload)
			wg.Done()
			seq = append(seq, 2)
			return ctx.Next()
		},
		func(ctx *Context) (interface{}, error) {
			require.NotNil(t, ctx.Payload)
			wg.Done()
			seq = append(seq, 3)
			return ctx.Next()
		},
	}
	gt := Use(middlewares...)
	wg.Add(len(middlewares) + 1)

	gt.Handle(func(ctx *Context) (interface{}, error) {
		require.NotNil(t, ctx.Payload)
		wg.Done()
		seq = append(seq, 4)
		return ctx.Return(&events.APIGatewayProxyResponse{Body: respBody})
	})

	out, err := gt.Invoke(context.Background(), []byte("{}"))
	require.NoError(t, err)

	var resp events.APIGatewayProxyResponse
	err = json.Unmarshal(out, &resp)
	require.NoError(t, err)
	require.Equal(t, resp.Body, respBody)

	wg.Wait()
	require.Equal(t, []int{1, 2, 3, 4}, seq)
}

func TestStopMiddlewares(t *testing.T) {
	respBody := "stop"

	wg := new(sync.WaitGroup)

	var seq []int
	middlewares := []Handler{
		func(ctx *Context) (interface{}, error) {
			require.NotNil(t, ctx.Payload)
			wg.Done()
			seq = append(seq, 1)
			return ctx.Return(&events.APIGatewayProxyResponse{Body: respBody})
		},
	}

	gt := Use(middlewares...)
	wg.Add(len(middlewares))

	gt.Handle(func(ctx *Context) (interface{}, error) {
		require.NotNil(t, ctx.Payload)
		wg.Done()
		seq = append(seq, 2)
		return ctx.Return(&events.APIGatewayProxyResponse{})
	})

	out, err := gt.Invoke(context.Background(), []byte("{}"))
	require.NoError(t, err)

	var resp events.APIGatewayProxyResponse
	err = json.Unmarshal(out, &resp)
	require.NoError(t, err)
	require.Equal(t, resp.Body, respBody)

	wg.Wait()
	require.Equal(t, []int{1}, seq)
}

func TestWith(t *testing.T) {
	gt := Use(func(ctx *Context) (interface{}, error) {
		return ctx.Next()
	})
	gt.With("k", "v").Handle(func(ctx *Context) (interface{}, error) {
		v := ctx.MustGet("k").(string)
		require.Equal(t, "v", v)
		return "", nil
	})
	_, err := gt.Invoke(context.Background(), []byte("{}"))
	require.NoError(t, err)
}

func TestDefaultErrHandler(t *testing.T) {
	gt := Use()
	err := errors.New("internal error")
	gt.Handle(func(ctx *Context) (interface{}, error) {
		return nil, err
	})
	out, e := gt.Invoke(context.Background(), []byte("{}"))
	require.NoError(t, e)
	require.Contains(t, string(out), "internal error")
}
