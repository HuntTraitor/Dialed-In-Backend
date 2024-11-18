package mocks

type MockMailer struct {
	SendCalledCount int
}

func NewMockMailer() *MockMailer {
	return &MockMailer{}
}

func (m *MockMailer) Send(email, template string, data interface{}) error {
	m.SendCalledCount++
	return nil
}
