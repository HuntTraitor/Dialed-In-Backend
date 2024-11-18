package mocks

type MockMailer struct{}

func (m MockMailer) Send(email, template string, data interface{}) error {
	return nil
}
