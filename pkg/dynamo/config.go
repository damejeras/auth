package dynamo

var (
	DefaultTableConfig = TableConfig{
		BasicCname:   "oauth2_basic",
		AccessCName:  "oauth2_access",
		RefreshCName: "oauth2_refresh",
	}
)

type TableConfig struct {
	BasicCname   string
	AccessCName  string
	RefreshCName string
}
