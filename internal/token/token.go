package token

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/xsteadfastx/carson/internal/ddns"
	"github.com/xsteadfastx/carson/internal/errs"
)

type Token struct{}

// Create creates token.
func (tk Token) Create(secret, hostname, recordType string) (string, error) {
	if recordType != "A" && recordType != "AAAA" {
		return "", errs.ErrWrongRecordType
	}

	s := []byte(secret)
	t := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&ddns.Record{
			Hostname: hostname,
			Type:     recordType,
		},
	)
	tString, err := t.SignedString(s)

	if err != nil {
		return "", err
	}

	return tString, nil
}

// Parse parses token.
func (tk Token) Parse(secret, token string) (ddns.Record, error) {
	s := []byte(secret)
	p, err := jwt.ParseWithClaims(token, &ddns.Record{}, func(tn *jwt.Token) (interface{}, error) {
		return s, nil
	})

	if err != nil {
		return ddns.Record{}, err
	}

	claims, ok := p.Claims.(*ddns.Record)
	if !ok || !p.Valid {
		return ddns.Record{}, errs.ErrFoo
	}

	if claims.Hostname == "" {
		return ddns.Record{}, errs.ErrNoHostname
	}

	return *claims, nil
}
