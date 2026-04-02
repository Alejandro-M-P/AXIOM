// Package languages contains programming language items for the DEV slot.
package languages

// NodeJS represents Node.js with npm.
// Installation is defined in toml/nodejs.toml
type NodeJS struct{}

func (s *NodeJS) ID() string             { return "nodejs" }
func (s *NodeJS) Name() string           { return "Node.js" }
func (s *NodeJS) Description() string    { return "Runtime de JavaScript con npm" }
func (s *NodeJS) Category() string       { return "dev" }
func (s *NodeJS) SubCategory() string    { return "languages" }
func (s *NodeJS) Dependencies() []string { return []string{} }
