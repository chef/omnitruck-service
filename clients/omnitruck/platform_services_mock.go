package omnitruck

type MockPlatformServices struct {
    PlatformVersionLatestFunc func(*RequestParams, int) (ProductVersion, error)
    PlatformVersionsAllFunc   func(*RequestParams, int) ([]ProductVersion, error)
    PlatformPackagesFunc      func(*RequestParams, int) (PackageList, error)
    PlatformMetadataFunc      func(*RequestParams, int) (PackageMetadata, error)
    PlatformFilenameFunc      func(*RequestParams, int) (string, error)
}

func (m *MockPlatformServices) PlatformVersionLatest(params *RequestParams, mode int) (ProductVersion, error) {
    if m.PlatformVersionLatestFunc != nil {
        return m.PlatformVersionLatestFunc(params, mode)
    }
    return "", nil
}

func (m *MockPlatformServices) PlatformVersionsAll(params *RequestParams, mode int) ([]ProductVersion, error) {
    if m.PlatformVersionsAllFunc != nil {
        return m.PlatformVersionsAllFunc(params, mode)
    }
    return nil, nil
}

func (m *MockPlatformServices) PlatformPackages(params *RequestParams, mode int) (PackageList, error) {
    if m.PlatformPackagesFunc != nil {
        return m.PlatformPackagesFunc(params, mode)
    }
    return nil, nil
}

func (m *MockPlatformServices) PlatformMetadata(params *RequestParams, mode int) (PackageMetadata, error) {
    if m.PlatformMetadataFunc != nil {
        return m.PlatformMetadataFunc(params, mode)
    }
    return PackageMetadata{}, nil
}

func (m *MockPlatformServices) PlatformFilename(params *RequestParams, mode int) (string, error) {
    if m.PlatformFilenameFunc != nil {
        return m.PlatformFilenameFunc(params, mode)
    }
    return "", nil
}
