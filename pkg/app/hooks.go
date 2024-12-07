package app

import "context"

// A ShutdownHook can be registered with a [Server] to be invoked right before or after the [Server] is about to be shutdown
type ShutdownHook func(ctx context.Context)

// A ShutdownErrorHook is another type of [ShutdownHook] that can return an error
type ShutdownErrorHook func(ctx context.Context) error

func shutdownHookAdapter(h ShutdownHook) ShutdownErrorHook {
	return func(ctx context.Context) error {
		h(ctx)
		return nil
	}
}
