package png

import (
	"context"
	"fmt"
	"os/exec"
)

func Encode(ctx context.Context, input string, output string) error {
	out, err := exec.CommandContext(ctx, "cp", input, output).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cp failed: %s %s", err.Error(), out)
	}

	out, err = exec.CommandContext(ctx, "optipng", "-o7", output).CombinedOutput()
	if err != nil {
		return fmt.Errorf("optipng failed: %s %s", err.Error(), out)
	}

	return nil
}
