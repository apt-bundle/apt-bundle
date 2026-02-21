package testutil

// MockExecutor implements the Executor interface (defined in internal/apt) for testing.
type MockExecutor struct {
	RunFunc     func(name string, args ...string) error
	OutputFunc  func(name string, args ...string) ([]byte, error)
	RunCalls    [][]string
	OutputCalls [][]string
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		RunCalls:    [][]string{},
		OutputCalls: [][]string{},
	}
}

func (m *MockExecutor) Run(name string, args ...string) error {
	call := append([]string{name}, args...)
	m.RunCalls = append(m.RunCalls, call)
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return nil
}

func (m *MockExecutor) Output(name string, args ...string) ([]byte, error) {
	call := append([]string{name}, args...)
	m.OutputCalls = append(m.OutputCalls, call)
	if m.OutputFunc != nil {
		return m.OutputFunc(name, args...)
	}
	return nil, nil
}
