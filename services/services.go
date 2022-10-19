package services

import "sync"

type Config struct {
	Listen string
}

type Service interface {
	Name() string
	Start(*sync.WaitGroup) error
	Stop() error
}
