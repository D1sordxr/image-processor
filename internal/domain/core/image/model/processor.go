package model

import "time"

type ProcessingOptions struct {
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	Quality       int    `json:"quality,omitempty"`
	Format        string `json:"format,omitempty"`
	Thumbnail     bool   `json:"thumbnail,omitempty"`
	WatermarkText string `json:"watermark_text,omitempty"`
}

type ProcessingResult struct {
	ProcessedData  []byte
	Format         string
	Width          int
	Height         int
	Size           int64
	ProcessingTime time.Duration
}
