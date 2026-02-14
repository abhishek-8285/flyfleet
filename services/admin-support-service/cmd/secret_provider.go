package main

type envSecretProvider struct {
	secret string
}

func (p *envSecretProvider) GetJWTSecret() string {
	return p.secret
}
