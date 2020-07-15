package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xsteadfastx/carson/internal/ddns"
	"github.com/xsteadfastx/carson/internal/errs"
	"github.com/xsteadfastx/carson/internal/token"
)

const secret = "foobar"
const hostname = "foohost"

func TestCreate(t *testing.T) {
	assert := assert.New(t)
	recordType := "A"
	expected := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJob3N0bmFtZSI6ImZvb2hvc3QiLCJ0eXBlIjoiQSIsIlRhcmdldCI6IiJ9." +
		"B7XUJbxPXNLT3pVT29ygMRONmHQG2q25RM0tUlQK-mg"

	tk := token.Token{}

	res, err := tk.Create(secret, hostname, recordType)

	assert.Nil(err)
	assert.Equal(expected, res)
}

func TestCreateTokenWrongType(t *testing.T) {
	assert := assert.New(t)
	recordType := "FOO"

	tk := token.Token{}

	res, err := tk.Create(secret, hostname, recordType)

	assert.Equal(errs.ErrWrongRecordType, err)
	assert.Equal("", res)
}

func TestParseToken(t *testing.T) {
	assert := assert.New(t)
	expected := ddns.Record{Hostname: "foohost"}
	tkn := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJob3N0bmFtZSI6ImZvb2hvc3QifQ." +
		"PDIExcwW0p-fVZLLnYJjfRXB7-Yt-Mq0m18eoBmBJnA"

	tk := token.Token{}

	c, err := tk.Parse(secret, tkn)

	assert.Nil(err)
	assert.Equal(expected, c)
}

func TestParseTokenWrongToken(t *testing.T) {
	assert := assert.New(t)
	tkn := "wrong.token"

	tk := token.Token{}

	c, err := tk.Parse(secret, tkn)

	assert.Equal(ddns.Record{}, c)
	assert.EqualError(err, "token contains an invalid number of segments")
}
