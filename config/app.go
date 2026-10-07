package config

const (
	AppName    = "SIAKAD Mini API"
	AppVersion = "1.0.0"
)

// IsProduction reports whether the app is running with APP_ENV=production.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}
