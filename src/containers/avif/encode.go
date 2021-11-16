package avif

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/seventv/EmoteProcessor/src/configure"
	"github.com/seventv/EmoteProcessor/src/image"
)

func Encode(ctx context.Context, config *configure.Config, imgSize image.ImageSize, dir string, delays []int) error {
	// ffmpeg -y -i input.gif -vsync 1 -pix_fmt yuva444p -f yuv4mpegpipe -strict -1 - | avifenc --stdin output.avif
	avifFile := path.Join(dir, fmt.Sprintf("%s.avif", string(imgSize)))
	ffmpegCmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-f", "image2",
		"-i", path.Join(dir, "frames", string(imgSize), "dump_%04d.png"),
		"-vsync", "0",
		"-f", "yuv4mpegpipe",
		"-pix_fmt", "yuva444p",
		"-strict", "-1",
		"pipe:1",
	)

	durations := make([]string, len(delays))
	for i, v := range delays {
		if v == 0 {
			v = 1
		}
		durations[i] = strconv.Itoa(v)
	}

	encoder := config.Av1Encoder
	if encoder == "" {
		encoder = "rav1e"
	}

	avifEncCmd := exec.CommandContext(
		ctx,
		"avifenc",
		"--stdin-durations", strconv.Itoa(len(delays)), strings.Join(durations, ","),
		"--speed", "3",
		"--timescale", "100",
		"--min", "10",
		"--max", "20",
		"--minalpha", "10",
		"--maxalpha", "20",
		"--jobs", "all",
		"--codec", encoder,
		"--stdin", avifFile,
	)

	r, w := io.Pipe()
	ffmpegCmd.Stdout = w
	avifEncCmd.Stdin = r

	err := avifEncCmd.Start()
	if err != nil {
		return err
	}

	err = ffmpegCmd.Start()
	if err != nil {
		return err
	}

	done := make(chan error)

	go func() {
		done <- ffmpegCmd.Wait()
		w.Close()
	}()

	go func() {
		done <- avifEncCmd.Wait()
		r.Close()
	}()

	return multierror.Append(<-done, <-done).ErrorOrNil()
}
