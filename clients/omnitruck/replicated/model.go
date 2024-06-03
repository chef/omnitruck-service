package replicated

type ResponseSearchCustomer struct {
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

type RequestCustomer struct {
	AppId                            string              `json:"app_id"`
	ChannelId                        string              `json:"channel_id"`
	Name                             string              `json:"name"`
	CustomerType                     string              `json:"type"`
	Email                            string              `json:"email"`
	EntitlementValues                []EntitlementValues `json:"entitlementValues"`
	ExpiresAt                        string              `json:"expires_at"`
	IsAirgapEnabled                  bool                `json:"is_airgap_enabled"`
	IsEmbeddedClusterDownloadEnabled bool                `json:"is_embedded_cluster_download_enabled"`
	IsGeoaxisSupported               bool                `json:"is_geoaxis_supported"`
	IsGitopsSupported                bool                `json:"is_gitops_supported"`
	IsHelmvmDownloadEnabled          bool                `json:"is_helmvm_download_enabled"`
	IsIdentityServiceSupported       bool                `json:"is_identity_service_supported"`
	IsKotsInstallEnabled             bool                `json:"is_kots_install_enabled"`
	IsSnapshotSupported              bool                `json:"is_snapshot_supported"`
	IsSupportBundleUploadEnabled     bool                `json:"is_support_bundle_upload_enabled"`
}
