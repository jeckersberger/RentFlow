package cache

// RedisCache is kept as a source-compatible alias for callers and tests that
// still use the previous cache type name. New code should use RedisClient.
type RedisCache = RedisClient
