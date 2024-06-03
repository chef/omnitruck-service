package replicated

type IReplicated interface {
	SearchCustomersByEmail(email string, requestId string) (customers []Customer, err error)
}
