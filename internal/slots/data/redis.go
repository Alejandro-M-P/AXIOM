// Package data contains database installation items.
package data

// Redis implements the Slot interface for Redis.
// Installation is defined in toml/redis.toml
type Redis struct{}

func (s *Redis) ID() string             { return "redis" }
func (s *Redis) Name() string           { return "Redis" }
func (s *Redis) Description() string    { return "Base de datos en memoria, cache y message broker" }
func (s *Redis) Category() string       { return "data" }
func (s *Redis) SubCategory() string    { return "databases" }
func (s *Redis) Dependencies() []string { return []string{} }
