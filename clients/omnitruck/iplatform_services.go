package omnitruck

type IPlatformServices interface {
	PlatformVersionLatest(*RequestParams, int) (ProductVersion, error)
	PlatformVersionsAll(*RequestParams, int) ([]ProductVersion, error)
	PlatformPackages(*RequestParams, int) (PackageList, error)
	PlatformMetadata(*RequestParams, int) (PackageMetadata, error)
	PlatformFilename(*RequestParams, int) (string, error)
}
