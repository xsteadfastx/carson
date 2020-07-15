package errs

import "errors"

var (
	ErrCouldNotExtractHostname = errors.New("could not extract hostname")
	ErrCouldNotExtractTarget   = errors.New("could not extract target")
	ErrFoo                     = errors.New("something went wrong")
	ErrMissingRecordHostname   = errors.New("missing hostname in token")
	ErrMissingRecordType       = errors.New("missing record type in token")
	ErrNoHostname              = errors.New("could not find hostname")
	ErrNoSerialFound           = errors.New("could not find serial in zone")
	ErrNoZoneForHostname       = errors.New("no zone for hostname")
	ErrWrongRecordType         = errors.New("only A or AAAA allowed")
)
