package users


var JwtInstance = NewJWT()

type Users struct {
	JWT *JwtService
}

func New() *Users {
	return &Users{
		JWT: JwtInstance,
	}
}
