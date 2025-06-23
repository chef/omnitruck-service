package config

type ServiceConfig struct {
	LicenseServiceUrl          string           `json:"licenseServiceUrl"`
	RelatedProductsTable       string           `json:"relatedProductsTable"`
	MetadataDetailsTable       string           `json:"metadataDetailsTable"`
	AWSConfig                  AWSConfig        `json:"awsConfig"`
	ReplicatedConfig           ReplicatedConfig `json:"replicatedConfig"`
	ReadWriteTimeout           int64            `json:"readWriteTimeout"`
	PackageManagersTable       string           `json:"packageManagersTable"`
	PackageDetailsCurrentTable string           `json:"packageDetailsCurrentTable"`
	PackageDetailsStableTable  string           `json:"packageDetailsStableTable"`
}

type ReplicatedConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
	AppID string `json:"appId"`
}

type AWSConfig struct {
	Region    string   `json:"region"`
	AccessKey string   `json:"access_key"`
	SecretKey string   `json:"secret_access_key"`
	S3Config  S3Config `json:"s3_config"`
}

type S3Config struct {
	Bucket      string `json:"bucket"`
	RoleArn     string `json:"role_arn"`
	StablePath  string `json:"stable_path"`
	CurrentPath string `json:"current_path"`
}
