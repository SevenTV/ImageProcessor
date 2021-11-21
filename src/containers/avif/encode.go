package avif

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/seventv/ImageProcessor/src/configure"
)

func Encode(ctx context.Context, config *configure.Config, name string, outName string, dir string, delays []int) error {
	// ffmpeg -y -i input.gif -vsync 1 -pix_fmt yuva444p -f yuv4mpegpipe -strict -1 - | avifenc --stdin output.avif
	avifFile := path.Join(dir, fmt.Sprintf("%s.avif", outName))
	var ffmpegCmd *exec.Cmd
	if len(delays) == 1 {
		ffmpegCmd = exec.CommandContext(
			ctx,
			"ffmpeg",
			"-i", path.Join(dir, "frames", name, "dump_0000.png"),
			"-vsync", "0",
			"-f", "yuv4mpegpipe",
			"-pix_fmt", "yuva444p",
			"-strict", "-1",
			"pipe:1",
		)
	} else {
		ffmpegCmd = exec.CommandContext(
			ctx,
			"ffmpeg",
			"-f", "image2",
			"-i", path.Join(dir, "frames", name, "dump_%04d.png"),
			"-vsync", "0",
			"-f", "yuv4mpegpipe",
			"-pix_fmt", "yuva444p",
			"-strict", "-1",
			"pipe:1",
		)
	}

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

	avifEncOut := bytes.NewBuffer(nil)
	ffmpegOut := bytes.NewBuffer(nil)

	ffmpegCmd.Stderr = ffmpegOut
	avifEncCmd.Stdout = avifEncOut
	avifEncCmd.Stderr = avifEncOut

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

	err = multierror.Append(<-done, <-done).ErrorOrNil()
	if err != nil {
		err = fmt.Errorf("avifenc failed: %s : %s : %s", err.Error(), ffmpegOut.String(), avifEncOut.String())
	}

	return err
}
