package pubsub

import (
	"context"
	"errors"
	"time"
)

type pullJob struct {
	subscription     *Subscriber
	maxPullMessages  int32
	longPullDuration *int32
	interval         time.Duration
	bufferSize       int
	errHandler       func(err error) bool
}

var ErrMissingCallback = NewConfigurationError("callback function is required", nil)

type PullJobOption func(*pullJob)

func WithPullMaxMessages(maximum int32) PullJobOption {
	return func(b *pullJob) {
		b.maxPullMessages = maximum
	}
}

func WithPullLongPullDuration(milliseconds int32) PullJobOption {
	return func(b *pullJob) {
		b.longPullDuration = &milliseconds
	}
}

func WithInterval(interval time.Duration) PullJobOption {
	return func(b *pullJob) {
		b.interval = interval
	}
}

func WithChannelBuffer(size int) PullJobOption {
	return func(b *pullJob) {
		b.bufferSize = size
	}
}

// WithErrorHandler allows the user to handle non-transient background errors
// and dictate whether the subscriber loop should continue (return true) or halt (return false).
func WithErrorHandler(handler func(err error) bool) PullJobOption {
	return func(b *pullJob) {
		if handler != nil {
			b.errHandler = handler
		}
	}
}

func newPullJob(s *Subscriber, opts []PullJobOption) *pullJob {
	b := &pullJob{
		subscription:    s,
		maxPullMessages: 10,
		interval:        time.Second * 5,
		bufferSize:      0,
		errHandler: func(err error) bool {
			s.logger.Error(err, "fatal background error")
			return true
		},
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *pullJob) runLoop(ctx context.Context, handler func(context.Context, PullMessages)) {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pullOpts := []PullOption{WithMaxMessages(b.maxPullMessages)}
			if b.longPullDuration != nil {
				pullOpts = append(pullOpts, WithLongPullDuration(*b.longPullDuration))
			}
			messages, err := b.subscription.Pull(ctx, pullOpts...)
			if err != nil { //nolint:nestif
				var sdkErr SDKError          // Declare the target variable
				if errors.As(err, &sdkErr) { // Pass a pointer to sdkErr
					if !sdkErr.IsTransient() {
						// Only exit the loop if the users error handler returns false
						if !b.errHandler(err) {
							return
						}
						continue
					}

					b.subscription.logger.Error(err, "transient error, retrying")
					continue
				}

				b.subscription.logger.Error(err, "unknown error, retrying")
				continue
			}

			if len(messages) > 0 {
				handler(ctx, messages)
			}
		}
	}
}

func (s *Subscriber) PullJobCallback(
	ctx context.Context,
	callback func(ctx context.Context, messages PullMessages),
	opts ...PullJobOption,
) error {
	if callback == nil {
		return ErrMissingCallback
	}

	job := newPullJob(s, opts)
	s.wg.Go(func() {
		job.runLoop(ctx, callback)
	})

	s.logger.Info(
		"new pull callback job created",
		"max_messages", int(job.maxPullMessages),
		"interval", job.interval,
	)
	return nil
}

func (s *Subscriber) PullJobChan(ctx context.Context, opts ...PullJobOption) (<-chan PullMessages, error) {
	job := newPullJob(s, opts)
	out := make(chan PullMessages, job.bufferSize)

	s.wg.Go(func() {
		defer close(out)

		adapter := func(innerCtx context.Context, msgs PullMessages) {
			select {
			case out <- msgs:
			case <-innerCtx.Done():
			}
		}

		job.runLoop(ctx, adapter)
	})

	s.logger.Info(
		"new pull job channel created",
		"max_messages", int(job.maxPullMessages),
		"interval", job.interval,
		"buffer_size", job.bufferSize,
	)
	return out, nil
}
