package app

type Config struct {
	AdminConfig struct {
		Port string `default:":9097"`
	} `fig:"admin""`
	Oauth2Config struct {
		Port string `default:":9096"`
	} `fig:"app"`
	AWSConfig struct {
		Region   string `validate:"required"`
		Endpoint string `validate:"required"`
		ID       string `validate:"required"`
		Secret   string `validate:"required"`
	} `fig:"aws"`
	ConsentProviderConfig struct {
		Address string `default:"http://localhost:8888/consent"`
	} `fig:"consent_provider"`
	IdentityProviderConfig struct {
		Address string `default:"http://localhost:8888/auth"`
	} `fig:"identity_provider"`
}
