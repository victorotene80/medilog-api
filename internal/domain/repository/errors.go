package repository

import "errors"

// ErrNotFound is the repository contract's signal that an operation matched no
// rows.
//
// Write-side methods previously returned gorm.ErrRecordNotFound directly, which
// leaked the ORM's sentinel through a domain port. Nothing above infrastructure
// can reasonably import GORM to test for it, so httperr.StatusFrom fell through
// to its default and a concurrent update or delete surfaced as 500 where 404
// was intended.
var ErrNotFound = errors.New("record not found")
