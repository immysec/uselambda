# uselambda
AWS Lambad middleware, like gin.


# Example
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/immysec/uselambda"
)

func SayHello(ctx *uselambda.Context) (interface{}, error) {
	res := &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Hello!",
	}
	return ctx.Return(res)
}

func Middleware(ctx *uselambda.Context) (interface{}, error) {
	fmt.Println("enter middleware first")
	return ctx.Next()
}

func main() {
	lambda.StartHandler(uselambda.Use(Middleware).Handle(SayHello))
}

```