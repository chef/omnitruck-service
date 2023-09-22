package config

type ServiceConfig struct {
	LicenseServiceUrl    string    `json:"licenseServiceUrl"`
	RelatedProductsTable string    `json:"relatedProductsTable"`
	MetadataDetailsTable string    `json:"metadataDetailsTable"`
	AWSConfig            AWSConfig `json:"awsConfig"`
}

type AWSConfig struct {
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_access_key"`
}
