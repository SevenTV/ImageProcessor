package gif

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"github.com/seventv/emote-processor/src/image"
)

func Encode(ctx context.Context, imgSize image.ImageSize, dir string, delays []int) error {
	file := path.Join(dir, "frames", string(imgSize), "dump_%04d.png")

	gifFile := path.Join(dir, fmt.Sprintf("%s.gif", imgSize))
	if err := exec.CommandContext(ctx, "ffmpeg", "-f", "image2", "-i", file, "-filter_complex", "geq=a=255:r=if(lt(alpha(X\\,Y)\\,128)\\,0\\,r(X\\,Y)):g=if(lt(alpha(X\\,Y)\\,128)\\,255\\,if(eq(r(X\\,Y)*b(X\\,Y)\\,0)*eq(g(X\\,Y)\\,255)\\,250\\,g(X\\,Y))):b=if(lt(alpha(X\\,Y)\\,128)\\,0\\,b(X\\,Y)),split[s0][s1];[s0]palettegen=reserve_transparent=1:transparency_color=#010101[p];[s1][p]paletteuse=new=1", gifFile).Run(); err != nil {
		return err
	}

	if err := exec.CommandContext(ctx, "gifsicle", "-b", "--colors=255", gifFile).Run(); err != nil {
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

	if err := exec.CommandContext(ctx, "gifsicle", "-b", "-U", "--disposal=previous", "--transparent=#00FF00", "-O3", gifFile).Run(); err != nil {
		return err
	}

	return nil
}
