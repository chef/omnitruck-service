package models

type AWSConfig struct {
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type ProductDetails struct {
	Product  string `json:"product"`
	Version  string `json:"version"`
	MetaData []MetaData `json:"metadata"`
}

type MetaData struct {
	Architecture     string
	Platform         string
	Platform_Version string
	SHA1             string
	SHA256           string
}

type Sku struct {
	Skus     string
	Products []string
}
