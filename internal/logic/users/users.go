package users

type Users struct {
	JWT *JwtService
}

func New() *Users {
	return &Users{
		JWT: NewJWT(),
	}
}
