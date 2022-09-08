package mqtt

import (
	"context"
	"github.com/c0olix/goChan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

//go:generate mockgen -destination ../gensrc/mocks/mqtt/paho.mqtt.golang_client.go -source $GOPATH/pkg/mod/github.com/eclipse/paho.mqtt.golang@v1.4.1/client.go
//go:generate mockgen -destination ../gensrc/mocks/mqtt/paho.mqtt.golang_token.go -source $GOPATH/pkg/mod/github.com/eclipse/paho.mqtt.golang@v1.4.1/token.go

type Channel struct {
	name              string
	qos               int
	reader            mqtt.Client
	writer            mqtt.Client
	readerMiddleWares []goChan.Middleware
	writerMiddleWares []goChan.Middleware
	errorCallBack     func(ctx context.Context, err error)
}

func (c Channel) Consume(handler goChan.Handler) {
	ctx := context.Background()
	handler = goChan.WrapMiddleware(c.readerMiddleWares, handler)
	callback := func(client mqtt.Client, message mqtt.Message) {
		err := handler(ctx, message)
		if err != nil {
			c.errorCallBack(ctx, errors.Wrap(err, "handler encountered an error"))
		}
	}
	if token := c.reader.Subscribe(c.name, byte(c.qos), callback); token.Wait() && token.Error() != nil {
		c.errorCallBack(ctx, errors.Wrap(token.Error(), "unable to subscribe to topic"))
	}
}

func (c Channel) Produce(ctx context.Context, messageInterface goChan.MessageInterface) error {
	handler := func(context.Context, goChan.MessageInterface) error {
		token := c.writer.Publish(c.name, byte(c.qos), true, messageInterface)
		if token.Wait() && token.Error() != nil {
			return errors.Wrap(token.Error(), "unable to publish")
		}
		return nil
	}

	handler = goChan.WrapMiddleware(c.writerMiddleWares, handler)

	return handler(ctx, messageInterface)
}

func (c Channel) SetReaderMiddleWares(mw ...goChan.Middleware) {
	c.readerMiddleWares = mw
}

func (c Channel) SetWriterMiddleWares(mw ...goChan.Middleware) {
	c.writerMiddleWares = mw
}
