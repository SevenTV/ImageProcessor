package gif

import (
	"context"
	"fmt"
	"os/exec"
	"path"
)

func Encode(ctx context.Context, name string, outName string, dir string, frames []string, delays []int) error {
	gifFile := path.Join(dir, fmt.Sprintf("%s.gif", outName))

	args := make([]string, len(delays)+2)
	dlay := make([]string, len(delays))
	for i := range delays {
		dlay[i] = fmt.Sprint(delays[i])
		args[i] = path.Join(dir, "frames", name, frames[i])
	}

	args[len(args)-2] = "--output"
	args[len(args)-1] = gifFile

	if out, err := exec.CommandContext(ctx, "gifski", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("gifski failed: %s : %s", err.Error(), out)
	}

	args = make([]string, len(delays)*2+2)
	args[0] = "-b"
	args[1] = gifFile
	for i, v := range delays {
		args[2+i*2] = fmt.Sprintf("--delay=%d", v)
		args[2+i*2+1] = fmt.Sprintf("#%d", i)
	}

	if out, err := exec.CommandContext(ctx, "gifsicle", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("gifsicle failed: %s : %s", err.Error(), out)
	}

	if out, err := exec.CommandContext(ctx, "gifsicle", "-b", "-O3", gifFile).CombinedOutput(); err != nil {
		return fmt.Errorf("gifsicle failed: %s : %s", err.Error(), out)
	}

	return nil
}
