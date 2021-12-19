package app

type Config struct {
	API struct {
		Port string `default:":9097"`
	}
	App struct {
		Port string `default:":9096"`
	}
	ConsentProvider struct {
		Address string `default:"http://localhost:8888/consent"`
	}
	IdentityProvider struct {
		Address string `default:"http://localhost:8888/auth"`
	}
	AWS struct {
		Region   string `validate:"required"`
		Endpoint string `validate:"required"`
		ID       string `validate:"required"`
		Secret   string `validate:"required"`
	}
}
