package security

type UserClaims struct {
	ID       uint   `json:"uid"`
	Username string `json:"uname,required"`
}
