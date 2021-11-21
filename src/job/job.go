package job

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Job struct {
	ID string `json:"id"`

	AspectRatioXY []int                `json:"aspect_ratio_xy"`
	Sizes         map[string]ImageSize `json:"sizes"`
	Settings      uint64               `json:"settings"`

	RawProvider           RawProvider         `json:"raw_provider"`
	RawProviderDetails    jsoniter.RawMessage `json:"raw_provider_details"`
	ResultConsumer        ResultConsumer      `json:"result_consumer"`
	ResultConsumerDetails jsoniter.RawMessage `json:"result_consumer_details"`
}

const (
	EnableOutputAnimatedGIF uint64 = 1 << iota
	EnableOutputAnimatedWEBP
	EnableOutputAnimatedAVIF
	EnableOutputStaticWEBP
	EnableOutputStaticAVIF
	EnableOutputStaticPNG
	EnableOutputAnimated
	EnableOutputAnimatedThumbanils
	AllSettings uint64 = (1 << iota) - 1
)

type File struct {
	Name        string        `json:"name"`
	Size        int           `json:"size"`
	ContentType string        `json:"content_type"`
	Animated    bool          `json:"animated"`
	TimeTaken   time.Duration `json:"time_taken"`
}

type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ImageVariant struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type RawProviderDetailsAws struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

type RawProviderDetailsLocal struct {
	Path string `json:"path"`
}

type ResultConsumerDetailsAws struct {
	Bucket    string `json:"bucket"`
	KeyFolder string `json:"key_folder"`
}

type ResultConsumerDetailsLocal struct {
	PathFolder string `json:"path_folder"`
}

type RawProvider string

const (
	AwsProvider   RawProvider = "aws"
	LocalProvider RawProvider = "local"
)

type ResultConsumer string

const (
	AwsConsumer   ResultConsumer = "aws"
	LocalConsumer ResultConsumer = "local"
)
