package png

import (
	"context"
	"os/exec"
)

func Encode(ctx context.Context, input string, output string) error {
	err := exec.CommandContext(ctx, "cp", input, output).Run()
	if err != nil {
		return err
	}

	return exec.CommandContext(ctx, "optipng", "-o7", output).Run()
}
