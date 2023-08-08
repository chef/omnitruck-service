package awsutils

import "github.com/aws/aws-sdk-go/aws/session"

type MockAwsUtils struct {
	GetNewSessionfunc func() (*session.Session, error)
}

func (mau *MockAwsUtils) GetNewSession() (*session.Session, error) {
	return mau.GetNewSessionfunc()
}