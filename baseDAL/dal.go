package baseDAL

import (
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"context"
)

// DAL is struct with common dependencies for DAL implementations
type DAL struct {
	DB                 *sqlx.DB
	PreparedStatements map[string]*sqlx.Stmt
	Logger             logrus.FieldLogger
	Ctx                context.Context
}

// PreparedStatement returns initialized statement or one of errors: ErrStmtNotFound, ErrStmtNotInitialized
func (d DAL) PreparedStatement(name string) (*sqlx.Stmt, error) {
	stmt, ok := d.PreparedStatements[name]
	if !ok {
		return nil, ErrStmtNotFound
	} else if stmt == nil {
		return nil, ErrStmtNotInitialized
	}
	return stmt, nil
}