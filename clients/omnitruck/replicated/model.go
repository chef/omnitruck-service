package replicated

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
