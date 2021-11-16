package containers

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	nGif "image/gif"
	nPng "image/png"

	"github.com/hashicorp/go-multierror"
	"github.com/seventv/EmoteProcessor/src/configure"
	"github.com/seventv/EmoteProcessor/src/containers/avi"
	"github.com/seventv/EmoteProcessor/src/containers/avif"
	"github.com/seventv/EmoteProcessor/src/containers/flv"
	"github.com/seventv/EmoteProcessor/src/containers/gif"
	"github.com/seventv/EmoteProcessor/src/containers/jpeg"
	"github.com/seventv/EmoteProcessor/src/containers/mp4"
	"github.com/seventv/EmoteProcessor/src/containers/png"
	"github.com/seventv/EmoteProcessor/src/containers/tiff"
	"github.com/seventv/EmoteProcessor/src/containers/webm"
	"github.com/seventv/EmoteProcessor/src/containers/webp"
	"github.com/seventv/EmoteProcessor/src/image"
	"github.com/seventv/EmoteProcessor/src/utils"
)

var (
	ErrUnknownFormat      = fmt.Errorf("unknown image format")
	ErrBadResponseAvifDec = fmt.Errorf("bad response from avifdec")
	ErrBadResponseFFprobe = fmt.Errorf("bad response from ffprobe")
	ErrUnknown            = fmt.Errorf("unknown")
)

var (
	avifDumpRe = regexp.MustCompile(`\d+\s+(\d+)\.\d+`)
	webpMuxRe  = regexp.MustCompile(`\s+\d+:\s+\d+\s+\d+\s+\w+\s+\d+\s+\d+\s+(\d+)\s+\w+\s+\w+\s+\d+\s+\s+\w+`)
)

func ToType(data []byte) (image.ImageType, error) {
	if avi.Test(data) {
		return image.AVI, nil
	} else if flv.Test(data) {
		return image.FLV, nil
	} else if gif.Test(data) {
		return image.GIF, nil
	} else if jpeg.Test(data) {
		return image.JPEG, nil
	} else if mp4.Test(data) {
		return image.MP4, nil
	} else if png.Test(data) {
		return image.PNG, nil
	} else if tiff.Test(data) {
		return image.TIFF, nil
	} else if webm.Test(data) {
		return image.WEBM, nil
	} else if webp.Test(data) {
		return image.WEBP, nil
	} else if avif.Test(data) { // do this test last because its very loose
		return image.AVIF, nil
	}

	return "", ErrUnknownFormat
}

func ProcessStage1(ctx context.Context, config *configure.Config, file string, imgType image.ImageType) (image.Image, error) {
	// we need to get infomation about frames for a few types.
	delay := []int{}
	frameCount := -1

	switch imgType {
	case image.GIF:
		// golang
		err := exec.CommandContext(ctx, "gifsicle", "-U", file, "-o", file).Run()
		if err != nil {
			return image.Image{}, err
		}

		f, err := os.OpenFile(file, os.O_RDONLY, 0600)
		if err != nil {
			return image.Image{}, err
		}
		defer f.Close()

		decGIF, err := nGif.DecodeAll(f)
		if err != nil {
			return image.Image{}, err
		}

		delay = decGIF.Delay
		frameCount = len(delay)
	case image.WEBP:
		// webpmux -info
		// avifdec -i
		data, err := exec.CommandContext(ctx, "webpmux", "-info", file).Output()
		if err != nil {
			return image.Image{}, err
		}

		matches := webpMuxRe.FindAllStringSubmatch(utils.B2S(data), -1)
		if len(matches) == 0 {
			// this is a static webp only 1 frame
			frameCount = 1
			delay = make([]int, 1)
		} else {
			frameCount = len(matches)
			delay = make([]int, frameCount)
			for i, m := range matches {
				delay[i], _ = strconv.Atoi(m[1])
				delay[i] /= 10
			}
		}
	case image.AVI, image.FLV, image.JPEG, image.MP4, image.PNG, image.TIFF, image.WEBM, image.AVIF:
	default:
		return image.Image{}, ErrUnknownFormat
	}

	dir := path.Dir(file)
	frameDir := path.Join(dir, "frames")
	if err := os.MkdirAll(frameDir, 0700); err != nil {
		return image.Image{}, err
	}

	// this will get all the frames.
	switch imgType {
	case image.AVI, image.FLV, image.GIF, image.JPEG, image.MP4, image.TIFF, image.WEBM, image.PNG:
		// ffmpeg
		if err := exec.CommandContext(ctx, "ffmpeg", "-i", file, "-vsync", "0", "-f", "image2", "-start_number", "0", fmt.Sprintf("%s/%s", frameDir, "dump_%04d.png")).Run(); err != nil {
			return image.Image{}, err
		}
		// we need to count them here tho.
		if frameCount == -1 {
			frameCount = 0
			if err := filepath.Walk(frameDir, func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}

				if filepath.Ext(path) == ".png" {
					frameCount++
				}
				return nil
			}); err != nil {
				return image.Image{}, err
			}
			delay = make([]int, frameCount)
			if frameCount == 0 {
				return image.Image{}, ErrUnknown
			}
			if frameCount > 1 {
				// we need to calculate the frame timings, by looking at the old fps.
				// :)
				fpsData, err := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-select_streams", "v", "-of", "default=noprint_wrappers=1:nokey=1", "-show_entries", "stream=r_frame_rate", file).Output()
				if err != nil {
					return image.Image{}, err
				}

				fpsSplits := strings.Split(utils.B2S(fpsData), "/")
				if len(fpsSplits) != 2 {
					return image.Image{}, ErrBadResponseFFprobe
				}

				fpsNum, err := strconv.Atoi(strings.TrimSpace(fpsSplits[0]))
				if err != nil {
					return image.Image{}, err
				}

				fpsDenom, err := strconv.Atoi(strings.TrimSpace(fpsSplits[1]))
				if err != nil {
					return image.Image{}, err
				}

				d := int(math.Floor(100 / (float64(fpsNum) / float64(fpsDenom))))
				for i := 0; i < frameCount; i++ {
					delay[i] = d
				}
			}
		}
	case image.AVIF:
		// avifdec
		decoder := config.Av1Decoder
		if decoder == "" {
			decoder = "dav1d"
		}

		if out, err := exec.CommandContext(
			ctx,
			"avifdump",
			"--codec", decoder,
			"--png-compress", "0",
			"--jobs", "all",
			"--depth", "16",
			file,
			fmt.Sprintf("%s/dump_%%04d.png", frameDir),
		).Output(); err != nil {
			return image.Image{}, err
		} else {
			matches := avifDumpRe.FindAllStringSubmatch(utils.B2S(out), -1)
			if len(matches) == 0 {
				return image.Image{}, ErrBadResponseAvifDec
			}

			frameCount = len(matches)
			delay = make([]int, frameCount)
			for i, m := range matches {
				delay[i], _ = strconv.Atoi(m[1])
				delay[i] /= 10
			}
		}
	case image.WEBP:
		// anim_dump
		if err := exec.CommandContext(ctx, "anim_dump", "-folder", frameDir, file).Run(); err != nil {
			return image.Image{}, err
		}
	default:
		return image.Image{}, ErrUnknownFormat
	}

	err := exec.CommandContext(ctx,
		"ffmpeg",
		"-f", "image2",
		"-start_number", "0",
		"-i", fmt.Sprintf("%s/%s", frameDir, "dump_%04d.png"),
		"-vf", "format=rgba,pad=h=if(gt(iw/ih\\,3)\\,iw/3\\,ih):w=if(lt(iw/ih\\,1)\\,ih\\,iw):x=0:y=(oh-ih):color=#00000000",
		"-f", "image2",
		"-start_number", "0",
		"-y", fmt.Sprintf("%s/%s", frameDir, "dump_%04d.png"),
	).Run()
	if err != nil {
		return image.Image{}, err
	}

	// we now at this point know how many frames are in the emote and also the timings.
	pngFile, err := os.OpenFile(path.Join(frameDir, "dump_0000.png"), os.O_RDONLY, 0600)
	if err != nil {
		return image.Image{}, err
	}
	defer pngFile.Close()

	pngCfg, err := nPng.DecodeConfig(pngFile)
	if err != nil {
		return image.Image{}, err
	}

	return image.Image{
		Dir:    dir,
		Width:  uint16(pngCfg.Width),
		Height: uint16(pngCfg.Height),
		Delays: delay,
	}, nil
}

func ProcessStage2(ctx context.Context, config *configure.Config, img image.Image) error {
	for _, v := range image.ImageSizes {
		dir := path.Join(img.Dir, "frames", string(v))
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	errCh := make(chan error, 5)
	defer close(errCh)

	for _, t := range image.ImageSizes {
		go func(t image.ImageSize) {
			errCh <- png.Edit(ctx, t, img.Dir, image.ImageSizesMap[t][2], image.ImageSizesMap[t][0], len(img.Delays))
		}(t)
	}

	var err error
	for i := 0; i < len(image.ImageSizes); i++ {
		err = multierror.Append(err, <-errCh).ErrorOrNil()
	}

	return err
}

func ProcessStage3(ctx context.Context, config *configure.Config, img image.Image) error {
	errCh := make(chan error)

	wg := sync.WaitGroup{}

	for _, v := range image.ImageSizes {
		wg.Add(1)
		go func(v image.ImageSize) {
			defer wg.Done()
			errCh <- webp.Encode(ctx, v, img.Dir, img.Delays)
		}(v)
	}

	if len(img.Delays) > 1 {
		for _, v := range image.ImageSizes {
			wg.Add(1)
			go func(v image.ImageSize) {
				defer wg.Done()
				errCh <- gif.Encode(ctx, v, img.Dir, img.Delays)
			}(v)
		}
	} else {
		for _, v := range image.ImageSizes {
			wg.Add(1)
			go func(v image.ImageSize) {
				defer wg.Done()
				errCh <- exec.CommandContext(ctx, "cp", path.Join(img.Dir, "frames", string(v), "dump_0000.png"), path.Join(img.Dir, fmt.Sprintf("%s.png", v))).Run()
			}(v)
		}
	}

	for _, v := range image.ImageSizes {
		wg.Add(1)
		go func(v image.ImageSize) {
			defer wg.Done()
			errCh <- avif.Encode(ctx, config, v, img.Dir, img.Delays)
		}(v)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var err error
	for e := range errCh {
		err = multierror.Append(err, e).ErrorOrNil()
	}

	return multierror.Append(err, os.RemoveAll(path.Join(img.Dir, "frames"))).ErrorOrNil()
}
