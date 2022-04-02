package png

import (
	"context"
	"fmt"
	"os/exec"
	"path"
)

func Edit(ctx context.Context, frames []string, dir string, name string, width uint16, height uint16) error {
	files := make([]string, len(frames)+4)
	for i := 0; i < len(frames); i++ {
		files[i+4] = path.Join(dir, "frames", frames[i])
	}
	files[0] = "-o"
	files[1] = path.Join(name, "%s.png")
	files[2] = "--size"
	files[3] = fmt.Sprintf("%dx%d", width, height)

	out, err := exec.CommandContext(ctx, "vipsthumbnail", files...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("vipsthumbnail failed: %s : %s", err.Error(), out)
	}

	for _, file := range frames {
		out, err = exec.CommandContext(ctx, "optipng", "-o7", path.Join(name, file), "-fix", "-force", "-clobber", "-silent").CombinedOutput()
		if err != nil {
			return fmt.Errorf("optipng failed: %s %s", err.Error(), out)
		}
	}

	return err
}
