package totp

var TotpInstance = New()

type Totp struct{}

func New() *Totp {
	return &Totp{}
}
