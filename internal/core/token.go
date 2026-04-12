package core

import (
	"encoding/json"
	"time"
)

type OathToken struct {
	Header    Header
	Claims    Claims
	Payload   []byte
	Signature []byte
}

type Header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type Claims struct {
	Iss         string    `json:"iss"`
	Sub         string    `json:"sub"`
	Aud         string    `json:"aud"`
	Act         string    `json:"act"`
	Iat         time.Time `json:"iat"`
	Exp         time.Time `json:"exp"`
	JTI         string    `json:"jti"`
	Scope       []string  `json:"scope,omitempty"`
	Reason      string    `json:"reason,omitempty"`
	Src         *Source   `json:"src,omitempty"`
	DelegatedBy string    `json:"delegated_by,omitempty"`
	Env         string    `json:"env,omitempty"`
	Tenant      string    `json:"tenant,omitempty"`
	RQH         string    `json:"rqh,omitempty"`
}

func (c *Claims) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

func ClaimsFromJSON(data []byte) (*Claims, error) {
	var c Claims
	err := json.Unmarshal(data, &c)
	return &c, err
}
