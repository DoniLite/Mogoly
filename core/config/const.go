package config

const (
	ENV_PREFIX string = "MOGOLY_"
)

const (
	BALANCER_STRATEGY     string = "BALANCER_STRATEGY"
	HEALTHCHECK_INTERVAL  string = "HEALTHCHECK_INTERVAL"
	LOG_LEVEL             string = "LOG_LEVEL"
	LOG_FORMAT            string = "LOG_FORMAT"
	LOG_OUTPUT            string = "LOG_OUTPUT"
	LOG_FILE              string = "LOG_FILE"
	MOGOLY_CERT_CACHE_DIR string = "CERT_CACHE_DIR"
	MOGOLY_EMAIL          string = "EMAIL"
	MOGOLY_ENV            string = "ENV"
	// Path to the hosts file
	// Default: /etc/hosts on Linux, C:\Windows\System32\drivers\etc\hosts on Windows
	// Can be overridden by setting the MOGOLY_HOSTS_PATH environment variable
	// Be careful when overriding this value as it may affect the system's ability to resolve hostnames
	MOGOLY_HOSTS_PATH     string = "HOSTS_PATH"
)
