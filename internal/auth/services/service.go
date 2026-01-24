package services

type AuthService struct {
	accountRepo      AccountRepository
	refreshTokenRepo RefreshTokenRepository
	cache            Cache
	txManager        TxManager
	secretKey        string
}

func NewAuthService(accountRepo AccountRepository, refreshTokenRepo RefreshTokenRepository, cache Cache, txManager TxManager, secretKey string) *AuthService {
	return &AuthService{
		accountRepo:      accountRepo,
		refreshTokenRepo: refreshTokenRepo,
		cache:            cache,
		txManager:        txManager,
		secretKey:        secretKey,
	}
}
