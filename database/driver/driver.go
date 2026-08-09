package driver

import (
	"errors"
	"net/url"

	"gorm.io/gorm"
)

type dialectBuilder = interface {
	Build(*url.URL) (gorm.Dialector, error)
}

var (
	builders             = map[string]dialectBuilder{}
	ErrorDialectNotFound = errors.New("dialect builder notfound")
)

func MakeConnection(connection string) (gorm.Dialector, error) {
	u, err := url.Parse(connection)
	if err != nil {
		return nil, err
	}
	builder, ok := builders[u.Scheme]
	if !ok {
		return nil, ErrorDialectNotFound
	}
	return builder.Build(u)
}
