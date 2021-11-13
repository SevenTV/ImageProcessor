package job

import jsoniter "github.com/json-iterator/go"

type Job struct {
	ID                    string              `json:"id"`
	RawProvider           RawProvider         `json:"raw_provider"`
	RawProviderDetails    jsoniter.RawMessage `json:"raw_provider_details"`
	ResultConsumer        ResultConsumer      `json:"result_consumer"`
	ResultConsumerDetails jsoniter.RawMessage `json:"result_consumer_details"`
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
