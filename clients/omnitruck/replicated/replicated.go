package replicated

import (
	"net/http"
)

type IReplicated interface {
	SearchCustomersByEmail(email string, requestId string) (customers []Customer, err error)
	GetDowloadUrl(customer Customer, requestId string) (url string, err error)
	DownloadFromReplicated(url, requestId, authorization string) (res *http.Response, err error)
}
