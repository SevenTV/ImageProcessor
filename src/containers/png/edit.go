package png

import (
	"context"
	"fmt"
	"os/exec"
	"path"
)

func Edit(ctx context.Context, frames []string, dir string, name string, width uint16, height uint16) error {
	args := make([]string, len(frames)+4)
	args[0] = "-o"
	args[1] = path.Join(name, "%s.png")
	args[2] = "--size"
	args[3] = fmt.Sprintf("%dx%d", width, height)

	for i, file := range frames {
		args[i+4] = path.Join(dir, "frames", file)
	}

	out, err := exec.CommandContext(ctx, "vipsthumbnail", args...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("vipsthumbnail failed: %s : %s", err.Error(), out)
	}

	return err
}
