package constants

const (
	SKU_PARTITION_KEY                    = "bom"
	PRODUCT_PARTITION_KEY                = "product"
	PRODUCT_SORT_KEY                     = "version"
	AUTOMATE_PRODUCT                     = "automate"
	HABITAT_PRODUCT                      = "habitat"
	CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT = "chef-ice"
	MIGRATE_ICE                          = "migrate-ice"
	LATEST                               = "latest"
	PLATFORM_SERVICE                     = "chef-360"
	PLATFORM_SERVICE_PRODUCT             = "chef-360"
	PLATFORM_ERROR                       = "chef-360 not available for the trial and opensource"
	REPLICATED_DOWNLOAD_URL              = "https://replicated.app/embedded"
	OCTET_STREAM                         = "application/octet-stream"
	PLATFORM_SERVICE_CONTENT_DISPOSITION = "attachment;filename=chef-360.tar.gz"
	CHUNKED                              = "chunked"
	CURRENT_CHANNEL                      = "current"
	STABLE_CHANNEL                       = "stable"
	PLATFORM_VERSION_KEY                 = "pv"
	DOWNLOAD_URL                         = "https://packages.chef.io/files/%s/%s/%s/%s"
	CHEF_AUTOMATE_CLI                    = "chef-automate-cli"
	AUTOMATE_CLI_VERSION                 = "latest"
	AUTOMATE_CHANNEL                     = "current"
	DUMMY_PACKAGE_MANAGER                = "pm"
)

const (
	UNMARSHAL_ERR_MSG                  = "error on unmarshal.\n[ERROR] -"
	SUCCESS_RESPONSE_FROM_FILENAME_MSG = "Returning success response from fileName API for "
	REPLICATED_CUSTOMER_ERROR          = "error while searching customer in replicated"
	REPLICATED_DOWNLOAD_ERROR          = "error while downloading from replicated"
	ERR_VALIDATING                     = "Error while validating params:"
)

type ApiType int

const (
	Trial ApiType = iota
	Opensource
	Commercial
)
