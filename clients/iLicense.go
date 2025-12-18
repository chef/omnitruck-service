package clients

type ILicense interface {
	Get(url string) *Request
	Validate(id, licenseServiceUrl string, data *Response) *Request
	GetReplicatedCustomerEmail(licenseId, licenseServiceUrl string, data *Response) *Request
	IsTrial(l string) bool
	IsFree(l string) bool
}
