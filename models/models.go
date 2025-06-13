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

type PackageDetails struct {
    Product  string              `json:"product"`  
    Version  string              `json:"version"`  
    Metadata map[string]Platform `json:"metadata"` 
}

type Platform struct {
    Architectures map[string]Architecture `json:"architectures"` 
}

type Architecture struct {
    PackageTypes map[string]PackageType `json:"package_types"` 
}

type PackageType struct {
    FileName       string `json:"filename"`        
    InstallMessage string `json:"install_message"` 
    SHA1           string `json:"sha1"`            
    SHA256         string `json:"sha256"`          
}
