package webp

import (
	"context"
	"fmt"
	"os/exec"
	"path"
)

func Encode(ctx context.Context, name string, outName string, dir string, frames []string, delays []int) error {
	webpFile := path.Join(dir, fmt.Sprintf("%s.webp", outName))

	if len(delays) == 1 {
		out, err := exec.CommandContext(ctx, "cwebp", "-z", "5", "-preset", "icon", "-sharpness", "3", path.Join(dir, "frames", name, frames[0]), "-o", webpFile).CombinedOutput()
		if err != nil {
			err = fmt.Errorf("cwebp failed: %s : %s", err.Error(), out)
		}

		return err
	}

	const argOffset = 11
	args := make([]string, len(delays)*3+argOffset)
	args[0] = "-o"
	args[1] = webpFile
	args[2] = "-loop"
	args[3] = "0"
	args[4] = "-mixed"
	args[5] = "-m"
	args[6] = "6"
	args[7] = "-kmax"
	args[8] = "0"
	args[9] = "-q"
	args[10] = "75"
	for i, v := range delays {
		args[argOffset+i*3] = "-d"
		args[argOffset+i*3+1] = fmt.Sprint(v * 10)
		args[argOffset+i*3+2] = path.Join(dir, "frames", name, frames[i])
	}

	out, err := exec.CommandContext(ctx, "img2webp", args...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("img2webp failed: %s : %s", err.Error(), out)
	}

	return err
}
