package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

var errDueToMode = errors.New("error due to error mode")

func (x *CLI) run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, x.command, x.args...)

	cmd.Stdin = x.Input
	cmd.Stderr = x.ErrOut

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe for command: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command %q: %w", x.command, err)
	}

	if proc := cmd.Process; proc == nil {
		panic("cmd.Process is nil after cmd.Start()")
	} else {
		ch := make(chan os.Signal, 64)
		defer close(ch)

		signal.Notify(ch)
		defer signal.Stop(ch)

		go func() {
			for sig := range ch {
				_ = proc.Signal(sig)
			}
		}()
	}

	var content bool

	{
		scanner := bufio.NewScanner(stdoutPipe)

		for scanner.Scan() {
			line := scanner.Text()

			var matched bool

			for _, re := range x.compiledPatterns {
				if re.MatchString(line) {
					matched = true
					break
				}
			}

			if x.invertMatch != matched {
				_, _ = fmt.Fprintln(x.Output, line)
				if !content {
					content = true
				}
			}
		}

		err = scanner.Err()
		if err != nil {
			return err
		}
	}

	err = stdoutPipe.Close()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	err = ctx.Err()
	if err != nil {
		return err
	}

	switch x.errorMode {
	case errorModeOnContent:
		if content {
			return errDueToMode
		}

	case errorModeNoContent:
		if !content {
			return errDueToMode
		}
	}

	return nil
}
