package baseDAL

import "errors"

var (
	// ErrStmtNotFound will be returned is statement for request was not found in registry
	ErrStmtNotFound = errors.New("statement with such name was not found in statementRegistry")
	// ErrStmtNotInitialized will be returned is statement for request was not initialized
	ErrStmtNotInitialized = errors.New("statement was not initialized")
)
