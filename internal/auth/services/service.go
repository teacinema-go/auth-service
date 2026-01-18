package services

type AuthService struct {
	accountRepo      AccountRepository
	refreshTokenRepo RefreshTokenRepository
	cache            Cache
	txManager        TxManager
}

func NewAuthService(accountRepo AccountRepository, refreshTokenRepo RefreshTokenRepository, cache Cache, txManager TxManager) *AuthService {
	return &AuthService{
		accountRepo:      accountRepo,
		refreshTokenRepo: refreshTokenRepo,
		cache:            cache,
		txManager:        txManager,
	}
}
