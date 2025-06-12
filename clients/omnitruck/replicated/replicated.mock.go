package replicated

import (
	"net/http"

	"github.com/chef/omnitruck-service/models"
)

type MockReplicated struct {
	SearchCustomersByEmailFunc func(email string, requestId string) (customers []models.Customer, err error)
	GetDowloadUrlFunc          func(customer models.Customer, requestId string) (url string, err error)
	DownloadFromReplicatedFunc func(url, requestId, authorization string) (res *http.Response, err error)
}

func (m MockReplicated) SearchCustomersByEmail(email string, requestId string) (customers []models.Customer, err error) {
	return m.SearchCustomersByEmailFunc(email, requestId)
}

func (m MockReplicated) GetDowloadUrl(customer models.Customer, requestId string) (url string, err error) {
	return m.GetDowloadUrlFunc(customer, requestId)
}

func (m MockReplicated) DownloadFromReplicated(url, requestId, authorization string) (res *http.Response, err error) {
	return m.DownloadFromReplicatedFunc(url, requestId, authorization)
}
