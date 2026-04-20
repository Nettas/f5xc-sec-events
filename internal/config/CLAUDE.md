# Package: internal/config

## Purpose
Load and validate configuration from environment variables and CLI flag overrides.

## Key File
- config.go — Config struct and Load() function

## Config Fields
- APIKey    string  From F5XC_API_KEY env var (required)
- Tenant    string  From F5XC_TENANT env var (default "f5-sa")
- Namespace string  From F5XC_NAMESPACE env var (default "s-iannetta")

## Notes
- Load() reads os.Getenv for each field
- Return an error if APIKey is empty
- CLI flags (--namespace, --lb) override the loaded config values in main.go
