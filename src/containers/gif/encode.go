package gif

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"github.com/davecgh/go-spew/spew"
)

func Encode(ctx context.Context, name string, outName string, dir string, delays []int) error {
	gifFile := path.Join(dir, fmt.Sprintf("%s.gif", outName))

	files := make([]string, len(delays)+2)
	for i := range delays {
		files[i] = path.Join(dir, "frames", name, fmt.Sprintf("dump_%04d.png", i))
	}

	files[len(files)-2] = "--output"
	files[len(files)-1] = gifFile

	if out, err := exec.CommandContext(ctx, "gifski", files...).CombinedOutput(); err != nil {
		spew.Dump(out)
		return err
	}

	args := make([]string, len(delays)*2+2)
	args[0] = "-b"
	args[1] = gifFile
	for i, v := range delays {
		args[2+i*2] = fmt.Sprintf("--delay=%d", v)
		args[2+i*2+1] = fmt.Sprintf("#%d", i)
	}

	if err := exec.CommandContext(ctx, "gifsicle", args...).Run(); err != nil {
		return err
	}

	if err := exec.CommandContext(ctx, "gifsicle", "-b", "-O3", gifFile).Run(); err != nil {
		return err
	}

	return nil
}
