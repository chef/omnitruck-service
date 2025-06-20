package utils

const (
	DBError                    = "Error while fetching the information for the product from DB."
	FetchVersionsError         = "Error while fetching product versions"
	FetchLatestOsVersionError  = "Error while fetching the latest opensource version for the product."
	BadRequestError            = "Product information not found. Please check the input parameters."
	ChannelParamsError         = "Channel can only be stable or current"
	ArchitectureParamsError    = "Architecture (m) params cannot be empty"
	BOMParamsError             = "BOM (bom) params cannot be empty"
	PlatformParamsError        = "Platfrom (p) params cannot be empty"
	PlatformVersionParamsError = "Platform Version (pv) params cannot be empty"

	OmnitruckDataNotFoundError = "Requested data is not found. Please check the input parameters"
	OmnitruckApiError          = "Error while fetching omnitruck data"
	OmnitruckReqError          = "Error while creating request for omnitruck"

	LicenseReqError                     = "Error while creating request for License validation"
	LicenseApiError                     = "Error while validating License"
	ErrorWhileFetchingLatestVersion     = "Error while fetching the latest version for the "
	ErrorLogUnsupportedPackageStructure = "GetProductPackages returned unsupported package structure"
	ErrorMsgUnsupportedPackageStructure = "Package details could not be interpreted. Please verify your request."
)
