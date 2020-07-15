package ddns

import "context"

//go:generate mockery --all

type Nower interface {
	Now() string
}

type Tokenizer interface {
	Create(secret, hostname, recordType string) (string, error)
	Parse(secret, token string) (Record, error)
}

type ZoneRefresher interface {
	Refresh(ctx context.Context, ddns *DDNS, r Record, remoteAddr string)
}
