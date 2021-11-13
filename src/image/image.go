package image

type Image struct {
	Dir    string
	Width  uint16
	Height uint16
	Delays []int
}

type ImageType string

const (
	AVI  ImageType = "avi"
	AVIF ImageType = "avif"
	FLV  ImageType = "flv"
	GIF  ImageType = "gif"
	JPEG ImageType = "jpeg"
	MP4  ImageType = "mp4"
	PNG  ImageType = "png"
	TIFF ImageType = "tiff"
	WEBM ImageType = "webm"
	WEBP ImageType = "webp"
)

type ImageSize string

const (
	OneX   ImageSize = "1x"
	TwoX   ImageSize = "2x"
	ThreeX ImageSize = "3x"
	FourX  ImageSize = "4x"
)

var ImageSizes = []ImageSize{
	OneX,
	TwoX,
	ThreeX,
	FourX,
}

var ImageSizesMap = map[ImageSize][3]uint16{
	OneX:   {OneXHeight, OneXMinWidth, OneXMaxWidth},
	TwoX:   {TwoXHeight, TwoXMinWidth, TwoXMaxWidth},
	ThreeX: {ThreeXHeight, ThreeXMinWidth, ThreeXMaxWidth},
	FourX:  {FourXHeight, FourXMinWidth, FourXMaxWidth},
}

const (
	OneXHeight   uint16 = 32
	TwoXHeight   uint16 = 64
	ThreeXHeight uint16 = 96
	FourXHeight  uint16 = 128

	OneXMinWidth   uint16 = OneXHeight
	TwoXMinWidth   uint16 = TwoXHeight
	ThreeXMinWidth uint16 = ThreeXHeight
	FourXMinWidth  uint16 = FourXHeight

	OneXMaxWidth   uint16 = OneXMinWidth * 3
	TwoXMaxWidth   uint16 = TwoXMinWidth * 3
	ThreeXMaxWidth uint16 = ThreeXMinWidth * 3
	FourXMaxWidth  uint16 = FourXMinWidth * 3
)
