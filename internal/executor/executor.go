// Package executor execute command
package executor

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"time"

	"go.uber.org/zap"

	event "github.com/hatappi/go-recmd/internal/event"
)

// Executor represent executor interface
type Executor interface {
	Run(ctx context.Context, commands []string) error
}
type executor struct {
	logger    *zap.Logger
	eventChan chan *event.Event
}

// NewExecutor initialize executor
func NewExecutor(logger *zap.Logger, eventChan chan *event.Event) Executor {
	return &executor{
		logger:    logger,
		eventChan: eventChan,
	}
}

func (e *executor) Run(ctx context.Context, commands []string) error {
	var (
		execCtx context.Context
		cancel  context.CancelFunc
	)

	errChan := make(chan error)

	execCtx, cancel = context.WithCancel(ctx)
	defer cancel()

	go func() {
		err := e.runCommand(execCtx, commands)
		if err != nil {
			errChan <- err
		}
	}()

	lastExecutedAt := time.Now()

	for {
		select {
		case evt := <-e.eventChan:
			e.logger.Debug("receive event", zap.Any("event", evt))

			d := evt.CreatedAt.Sub(lastExecutedAt)
			// If it runs continuously for a short period of time, it will get an killed error.
			if d.Seconds() < 1.0 {
				continue
			}

			cancel()

			lastExecutedAt = evt.CreatedAt
			go func() {
				execCtx, cancel = context.WithCancel(ctx)
				err := e.runCommand(execCtx, commands)
				if err != nil {
					errChan <- err
				}
			}()
		case err := <-errChan:
			if err == context.Canceled || err == context.DeadlineExceeded {
				continue
			}
			return err
		case <-ctx.Done():
			e.logger.Debug("finish executor")
			return nil
		}
	}
}

func (e *executor) runCommand(ctx context.Context, commands []string) error {
	cmd := exec.CommandContext(ctx, commands[0], commands[1:]...)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(outReader)
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(errReader)
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
