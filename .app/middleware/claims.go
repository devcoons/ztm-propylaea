package middleware

type SJWTClaims struct {
	Auth    bool   `json:"auth"`
	UserId  int    `json:"userid"`
	Role    int    `json:"role"`
	Service string `json:"service"`
	Hop     int    `json:"hop"`
}

type UJWTClaims struct {
	Auth   bool   `json:"a"`
	UserId int    `json:"u"`
	Role   int    `json:"r"`
	Nonce  string `json:"n"`
}

type UJWTClaimsMinimal struct {
	A bool   `json:"a"`
	U int    `json:"u"`
	R int    `json:"r"`
	N string `json:"n"`
}
