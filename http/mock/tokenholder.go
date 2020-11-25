package mock

import (
	"github.com/stretchr/testify/mock"
)

// TokenHolder mocks an object implementing the `http.tokenHolder` interface.
type TokenHolder struct {
	mock.Mock
}

// InvalidateToken mocks the implementation of the real method.
func (t *TokenHolder) InvalidateToken() {
	t.Called()
}

// Raw mocks the implementation of the real method.
func (t *TokenHolder) RawToken() string {
	args := t.Called()
	return args.String(0)
}

// Refresh mocks the implementation of the real method.
func (t *TokenHolder) RefreshToken() error {
	args := t.Called()
	return args.Error(0)
}
