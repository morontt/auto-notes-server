package security

const (
	UserContextKey = "user_ctx_key"
)

type UserClaims struct {
	ID       uint   `json:"uid"`
	Username string `json:"uname,required"`
}
