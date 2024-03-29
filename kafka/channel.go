package kafka

import (
	"context"
	"github.com/c0olix/goChan"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

//go:generate mockgen -destination ../gensrc/mocks/kafka/$GOFILE -source $GOFILE

type ReaderInterface interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, messages ...kafka.Message) error
}

type WriterInterface interface {
	WriteMessages(ctx context.Context, messages ...kafka.Message) error
}

type Channel struct {
	reader            ReaderInterface
	writer            WriterInterface
	readerMiddleWares []goChan.Middleware
	writerMiddleWares []goChan.Middleware
}

func (channel *Channel) Consume(handler goChan.Handler) chan error {
	handler = goChan.WrapMiddleware(channel.readerMiddleWares, handler)
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		for {
			ctx := context.Background()
			message, err := channel.reader.FetchMessage(ctx)
			if err != nil {
				errChan <- errors.Wrap(err, "error while fetching Kafka message")
				continue
			}

			err = handler(ctx, message)
			if err != nil {
				errChan <- errors.Wrap(err, "error while calling handler")
				continue
			}

			err = channel.reader.CommitMessages(ctx, message)
			if err != nil {
				errChan <- errors.Wrap(err, "error while committing Kafka message")
				continue
			}
		}
	}()
	return errChan
}

func (channel *Channel) Produce(ctx context.Context, proto goChan.MessageInterface) error {
	handler := func(ctx context.Context, proto goChan.MessageInterface) error {
		message, ok := proto.(kafka.Message)
		if !ok {
			return errors.New("type is not kafka message")
		}
		err := channel.writer.WriteMessages(ctx, message)
		if err != nil {
			return errors.Wrap(err, "unable to write event")
		}
		return nil
	}

	handler = goChan.WrapMiddleware(channel.writerMiddleWares, handler)

	return handler(ctx, proto)
}

func (channel *Channel) SetReaderMiddleWares(mw ...goChan.Middleware) {
	channel.readerMiddleWares = mw
}

func (channel *Channel) SetWriterMiddleWares(mw ...goChan.Middleware) {
	channel.writerMiddleWares = mw
}
