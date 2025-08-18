package core

const (
	MogolyRatelimiter MiddleWareName = "mogoly:ratelimiter"
)

var MiddlewaresList MiddlewareSets = MiddlewareSets{
	MogolyRatelimiter: struct {
		fn   MogolyMiddleware
		conf any
	}{
		fn:   RateLimiterMiddleware,
		conf: &RateLimitMiddlewareConfig{},
	},
}
