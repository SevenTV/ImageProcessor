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
	"time"

	nGif "image/gif"
	nPng "image/png"

	"github.com/hashicorp/go-multierror"
	"github.com/seventv/ImageProcessor/src/configure"
	"github.com/seventv/ImageProcessor/src/containers/avi"
	"github.com/seventv/ImageProcessor/src/containers/avif"
	"github.com/seventv/ImageProcessor/src/containers/flv"
	"github.com/seventv/ImageProcessor/src/containers/gif"
	"github.com/seventv/ImageProcessor/src/containers/jpeg"
	"github.com/seventv/ImageProcessor/src/containers/mov"
	"github.com/seventv/ImageProcessor/src/containers/mp4"
	"github.com/seventv/ImageProcessor/src/containers/png"
	"github.com/seventv/ImageProcessor/src/containers/tiff"
	"github.com/seventv/ImageProcessor/src/containers/webm"
	"github.com/seventv/ImageProcessor/src/containers/webp"
	"github.com/seventv/ImageProcessor/src/image"
	"github.com/seventv/ImageProcessor/src/job"
	"github.com/seventv/ImageProcessor/src/utils"
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
	} else if mov.Test(data) {
		return image.MOV, nil
	} else if avif.Test(data) { // do this test last because its very loose
		return image.AVIF, nil
	}

	return "", ErrUnknownFormat
}

func ProcessStage1(ctx context.Context, config *configure.Config, file string, imgType image.ImageType, aspectRatioXY [2]int) (image.Image, error) {
	// we need to get infomation about frames for a few types.
	delay := []int{}
	frameCount := -1

	switch imgType {
	case image.GIF:
		// golang
		out, err := exec.CommandContext(ctx, "gifsicle", "-U", file, "-o", file).CombinedOutput()
		if err != nil {
			return image.Image{}, fmt.Errorf("gifsicle failed: %s : %s", err.Error(), out)
		}

		f, err := os.OpenFile(file, os.O_RDONLY, 0600)
		if err != nil {
			return image.Image{}, fmt.Errorf("read file failed: %s", err.Error())
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
		data, err := exec.CommandContext(ctx, "webpmux", "-info", file).CombinedOutput()
		if err != nil {
			return image.Image{}, fmt.Errorf("webpmux failed: %s : %s", err.Error(), data)
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
	case image.AVI, image.FLV, image.JPEG, image.MP4, image.PNG, image.TIFF, image.WEBM, image.AVIF, image.MOV:
	default:
		return image.Image{}, ErrUnknownFormat
	}

	dir := path.Dir(file)
	frameDir := path.Join(dir, "frames")
	if err := os.MkdirAll(frameDir, 0700); err != nil {
		return image.Image{}, fmt.Errorf("mkdir failed: %s", err.Error())
	}

	// this will get all the frames.
	switch imgType {
	case image.AVI, image.FLV, image.GIF, image.JPEG, image.MP4, image.TIFF, image.WEBM, image.PNG, image.MOV:
		// ffmpeg
		if out, err := exec.CommandContext(ctx, "ffmpeg", "-i", file, "-vsync", "0", "-f", "image2", "-start_number", "0", fmt.Sprintf("%s/%s", frameDir, "dump_%04d.png")).CombinedOutput(); err != nil {
			return image.Image{}, fmt.Errorf("ffmpeg failed: %s : %s", err.Error(), out)
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
				return image.Image{}, fmt.Errorf("filepath walk failed: %s", err.Error())
			}
			delay = make([]int, frameCount)
			if frameCount == 0 {
				return image.Image{}, ErrUnknown
			}
			if frameCount > 1 {
				// we need to calculate the frame timings, by looking at the old fps.
				// :)
				fpsData, err := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-select_streams", "v", "-of", "default=noprint_wrappers=1:nokey=1", "-show_entries", "stream=r_frame_rate", file).CombinedOutput()
				if err != nil {
					return image.Image{}, fmt.Errorf("ffprobe failed: %s : %s", err.Error(), fpsData)
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
		).CombinedOutput(); err != nil {
			return image.Image{}, fmt.Errorf("avifdump failed: %s : %s", err.Error(), out)
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
		if out, err := exec.CommandContext(ctx, "anim_dump", "-folder", frameDir, file).CombinedOutput(); err != nil {
			return image.Image{}, fmt.Errorf("anim_dump failed: %s : %s", err.Error(), out)
		}
	default:
		return image.Image{}, ErrUnknownFormat
	}

	out, err := exec.CommandContext(ctx,
		"ffmpeg",
		"-f", "image2",
		"-start_number", "0",
		"-i", fmt.Sprintf("%s/%s", frameDir, "dump_%04d.png"),
		"-vf", fmt.Sprintf("format=rgba,pad=h=if(gt(iw/ih\\,%d)\\,iw/%d\\,ih):w=if(lt(iw/ih\\,%d)\\,ih/%d\\,iw):x=0:y=(oh-ih):color=#00000000", aspectRatioXY[0], aspectRatioXY[0], aspectRatioXY[1], aspectRatioXY[1]),
		"-f", "image2",
		"-start_number", "0",
		"-y", fmt.Sprintf("%s/%s", frameDir, "dump_%04d.png"),
	).CombinedOutput()
	if err != nil {
		return image.Image{}, fmt.Errorf("ffmpeg failed: %s : %s", err.Error(), out)
	}

	// we now at this point know how many frames are in the emote and also the timings.
	pngFile, err := os.OpenFile(path.Join(frameDir, "dump_0000.png"), os.O_RDONLY, 0600)
	if err != nil {
		return image.Image{}, fmt.Errorf("open file failed: %s", err.Error())
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

func ProcessStage2(ctx context.Context, config *configure.Config, img image.Image, sizes map[string]job.ImageSize) error {
	for v := range sizes {
		dir := path.Join(img.Dir, "frames", v)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("mkdir failed: %s", err.Error())
		}
	}

	errCh := make(chan error, 5)
	defer close(errCh)

	for name, size := range sizes {
		go func(name string, size job.ImageSize) {
			errCh <- png.Edit(ctx, name, img.Dir, uint16(size.Width), uint16(size.Height), len(img.Delays))
		}(name, size)
	}

	var err error
	for i := 0; i < len(sizes); i++ {
		err = multierror.Append(err, <-errCh).ErrorOrNil()
	}

	return err
}

func ProcessStage3(ctx context.Context, config *configure.Config, img image.Image, sizes map[string]job.ImageSize, settings uint64) ([]job.File, error) {
	errCh := make(chan error)

	wg := sync.WaitGroup{}

	isAnimated := len(img.Delays) > 1

	files := []job.File{}
	fileChan := make(chan job.File)
	start := time.Now()

	// AVIF
	if (settings&job.EnableOutputAnimatedAVIF != 0 && isAnimated) || (settings&job.EnableOutputStaticAVIF != 0 && !isAnimated) {
		for name, size := range sizes {
			wg.Add(1)
			go func(name string, size job.ImageSize) {
				defer wg.Done()
				err := avif.Encode(ctx, config, name, name, img.Dir, img.Delays)
				if err == nil {
					info, err := os.Stat(path.Join(img.Dir, fmt.Sprintf("%s.avif", name)))
					if err != nil {
						errCh <- err
						return
					}

					fileChan <- job.File{
						Name:        fmt.Sprintf("%s.avif", name),
						ContentType: "image/avif",
						Size:        int(info.Size()),
						Animated:    isAnimated,
						Width:       int(float64(size.Height) / float64(img.Height) * float64(img.Width)),
						Height:      size.Height,
						TimeTaken:   time.Since(start),
					}
				}
				errCh <- err
			}(name, size)
		}
	}

	// WEBP
	if (settings&job.EnableOutputAnimatedWEBP != 0 && isAnimated) || (settings&job.EnableOutputStaticWEBP != 0 && !isAnimated) {
		for name, size := range sizes {
			wg.Add(1)
			go func(name string, size job.ImageSize) {
				defer wg.Done()
				err := webp.Encode(ctx, name, name, img.Dir, img.Delays)
				if err == nil {
					info, err := os.Stat(path.Join(img.Dir, fmt.Sprintf("%s.webp", name)))
					if err != nil {
						errCh <- err
						return
					}

					fileChan <- job.File{
						Name:        fmt.Sprintf("%s.webp", name),
						ContentType: "image/webp",
						Size:        int(info.Size()),
						Animated:    isAnimated,
						Width:       int(float64(size.Height) / float64(img.Height) * float64(img.Width)),
						Height:      size.Height,
						TimeTaken:   time.Since(start),
					}
				}
				errCh <- err
			}(name, size)
		}
	}

	// GIF
	if settings&job.EnableOutputAnimatedGIF != 0 && isAnimated {
		for name, size := range sizes {
			wg.Add(1)
			go func(name string, size job.ImageSize) {
				defer wg.Done()
				err := gif.Encode(ctx, name, name, img.Dir, img.Delays)
				if err == nil {
					info, err := os.Stat(path.Join(img.Dir, fmt.Sprintf("%s.gif", name)))
					if err != nil {
						errCh <- err
						return
					}

					fileChan <- job.File{
						Name:        fmt.Sprintf("%s.gif", name),
						ContentType: "image/gif",
						Size:        int(info.Size()),
						Animated:    isAnimated,
						Width:       int(float64(size.Height) / float64(img.Height) * float64(img.Width)),
						Height:      size.Height,
						TimeTaken:   time.Since(start),
					}
				}
				errCh <- err
			}(name, size)
		}
	}

	// PNG
	if settings&job.EnableOutputStaticPNG != 0 && !isAnimated {
		for name, size := range sizes {
			wg.Add(1)
			go func(name string, size job.ImageSize) {
				defer wg.Done()
				err := png.Encode(ctx, path.Join(img.Dir, "frames", name, "dump_0000.png"), path.Join(img.Dir, fmt.Sprintf("%s.png", name)))
				if err == nil {
					info, err := os.Stat(path.Join(img.Dir, fmt.Sprintf("%s.png", name)))
					if err != nil {
						errCh <- err
						return
					}

					fileChan <- job.File{
						Name:        fmt.Sprintf("%s.png", name),
						ContentType: "image/png",
						Size:        int(info.Size()),
						Animated:    false,
						Width:       int(float64(size.Height) / float64(img.Height) * float64(img.Width)),
						Height:      size.Height,
						TimeTaken:   time.Since(start),
					}
				}
				errCh <- err
			}(name, size)
		}
	}

	// THUMBNAILS
	if isAnimated && settings&job.EnableOutputAnimatedThumbanils != 0 {
		for name, size := range sizes {
			if settings&job.EnableOutputStaticAVIF != 0 {
				wg.Add(1)
				go func(name string, size job.ImageSize) {
					defer wg.Done()
					err := avif.Encode(ctx, config, name, fmt.Sprintf("%s_static", name), img.Dir, img.Delays[:1])
					if err == nil {
						info, err := os.Stat(path.Join(img.Dir, fmt.Sprintf("%s_static.avif", name)))
						if err != nil {
							errCh <- err
							return
						}

						fileChan <- job.File{
							Name:        fmt.Sprintf("%s_static.avif", name),
							ContentType: "image/avif",
							Size:        int(info.Size()),
							Animated:    false,
							Width:       int(float64(size.Height) / float64(img.Height) * float64(img.Width)),
							Height:      size.Height,
							TimeTaken:   time.Since(start),
						}
					}
					errCh <- err
				}(name, size)
			}
			if settings&job.EnableOutputStaticWEBP != 0 {
				wg.Add(1)
				go func(name string, size job.ImageSize) {
					defer wg.Done()
					err := webp.Encode(ctx, name, fmt.Sprintf("%s_static", name), img.Dir, img.Delays[:1])
					if err == nil {
						info, err := os.Stat(path.Join(img.Dir, fmt.Sprintf("%s_static.webp", name)))
						if err != nil {
							errCh <- err
							return
						}

						fileChan <- job.File{
							Name:        fmt.Sprintf("%s_static.webp", name),
							ContentType: "image/webp",
							Size:        int(info.Size()),
							Animated:    false,
							Width:       int(float64(size.Height) / float64(img.Height) * float64(img.Width)),
							Height:      size.Height,
							TimeTaken:   time.Since(start),
						}
					}
					errCh <- err
				}(name, size)

			}
			if settings&job.EnableOutputStaticPNG != 0 {
				wg.Add(1)
				go func(name string, size job.ImageSize) {
					defer wg.Done()
					err := png.Encode(ctx, path.Join(img.Dir, "frames", name, "dump_0000.png"), path.Join(img.Dir, fmt.Sprintf("%s_static.png", name)))
					if err == nil {
						info, err := os.Stat(path.Join(img.Dir, fmt.Sprintf("%s_static.png", name)))
						if err != nil {
							errCh <- err
							return
						}

						fileChan <- job.File{
							Name:        fmt.Sprintf("%s_static.png", name),
							ContentType: "image/png",
							Size:        int(info.Size()),
							Animated:    false,
							Width:       int(float64(size.Height) / float64(img.Height) * float64(img.Width)),
							Height:      size.Height,
							TimeTaken:   time.Since(start),
						}
					}
					errCh <- err
				}(name, size)
			}
		}
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(fileChan)
	}()

	wg2 := sync.WaitGroup{}
	wg2.Add(2)
	var err error

	go func() {
		defer wg2.Done()
		for e := range errCh {
			err = multierror.Append(err, e).ErrorOrNil()
		}
	}()

	go func() {
		defer wg2.Done()
		for f := range fileChan {
			files = append(files, f)
		}
	}()

	wg2.Wait()

	return files, multierror.Append(err, os.RemoveAll(path.Join(img.Dir, "frames"))).ErrorOrNil()
}
