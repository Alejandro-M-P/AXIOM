package main

import (
	"axiom/internal/router"
)

// Re-export router types and functions for backwards compatibility
type (
	Router = router.Router
)

const (
	CmdCreate  = router.CmdCreate
	CmdDelete  = router.CmdDelete
	CmdList    = router.CmdList
	CmdStop    = router.CmdStop
	CmdPrune   = router.CmdPrune
	CmdBuild   = router.CmdBuild
	CmdRebuild = router.CmdRebuild
	CmdHelp    = router.CmdHelp
	CmdInfo    = router.CmdInfo
	CmdReset   = router.CmdReset
	CmdEnter   = router.CmdEnter
	CmdInit    = router.CmdInit
)

var NewRouter = router.NewRouter
var KnownCommand = router.KnownCommand
