package replicated

type MockReplicated struct {
	SearchCustomersByEmailFunc func(email string, requestId string) (customers []Customer, err error)
}

func (m MockReplicated) SearchCustomersByEmail(email string, requestId string) (customers []Customer, err error) {
	return m.SearchCustomersByEmailFunc(email, requestId)
}
