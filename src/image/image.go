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
	MOV  ImageType = "mov"
)
