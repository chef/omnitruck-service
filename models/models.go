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
	FileName         string `json:"filename"`
	Platform         string `json:"platform"`
	Platform_Version string `json:"platform_version"`
	SHA1             string `json:"sha1"`
	SHA256           string `json:"sha256"`
}

type RelatedProducts struct {
	Bom      string            `json:"bom"`
	Products map[string]string `json:"products"`
}

type ScriptParams struct {
	BaseUrl   string `json:"base_url"`
	LicenseId string `json:"licenseId"`
}

//Replicated models

type CustomerSearchResponse struct {
	Query     string     `json:"query"`
	TotalHits int        `json:"total_hits"`
	Customers []Customer `json:"customers"`
}

type GetCustomerResponse struct {
	Customer Customer `json:"customer"`
}

type Customer struct {
	ID             string    `json:"id"`
	TeamID         string    `json:"teamId"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	CustomID       string    `json:"customId"`
	ExpiresAt      string    `json:"expiresAt"`
	CustomerType   string    `json:"type"`
	Airgap         bool      `json:"airgap"`
	InstallationId string    `json:"installationId"`
	Channels       []Channel `json:"channels"`
}

type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AppID       string `json:"appId"`
	AppSlug     string `json:"appSlug"`
	AppName     string `json:"appName"`
	ChannelSlug string `json:"channelSlug"`
}

type EntitlementValues struct {
	IsDefault bool   `json:"isDefault"`
	Name      string `json:"name"`
	Value     string `json:"value"`
}

type ApiType int

const (
	Trial ApiType = iota
	Opensource
	Commercial
)
