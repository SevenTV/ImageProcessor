package png

import (
	"context"
	"fmt"
	"os/exec"
	"path"
)

func Edit(ctx context.Context, name string, dir string, width uint16, height uint16, frameCount int) error {
	files := make([]string, frameCount+4)
	for i := 0; i < frameCount; i++ {
		files[i+4] = path.Join(dir, "frames", fmt.Sprintf("dump_%04d.png", i))
	}
	files[0] = "-o"
	files[1] = path.Join(name, "%s.png")
	files[2] = "--size"
	files[3] = fmt.Sprintf("%dx%d", width, height)

	return exec.CommandContext(ctx, "vipsthumbnail", files...).Run()
}
