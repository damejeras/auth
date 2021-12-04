package dynamo

type Option func(store *tokenStore)

func WithTableConfig(config TableConfig) Option {
	return func(store *tokenStore) {
		store.tables = config
	}
}
