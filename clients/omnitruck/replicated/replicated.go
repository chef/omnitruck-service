package replicated

import (
	"net/http"

	"github.com/chef/omnitruck-service/models"
)

type IReplicated interface {
	SearchCustomersByEmail(email string, requestId string) (customers []models.Customer, err error)
	GetDowloadUrl(customer models.Customer, requestId string) (url string, err error)
	DownloadFromReplicated(url, requestId, authorization string) (res *http.Response, err error)
}
