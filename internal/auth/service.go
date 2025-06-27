package auth


type service interface {
	Register(req RegisterRequest) (*TokenResponse, error)
}