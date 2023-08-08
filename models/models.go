package models

type AWSConfig struct {
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type ProductDetails struct {
	Product  string     `json:"product"`
	Version  string     `json:"version"`
	MetaData []MetaData `json:"metadata"`
}

type MetaData struct {
	Architecture     string `json:"architecture"`
	Platform         string `json:"platform"`
	Platform_Version string `json:"platform_version"`
	SHA1             string `json:"sha1"`
	SHA256           string `json:"sha256"`
}

type Sku struct {
	Skus     string   `json:"skus"`
	Products []string `json:"products"`
}
