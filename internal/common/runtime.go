package common

import "archil.lab.tech.com/internal/config"

// Keep a single process-wide pointer to your loaded config.
var appCfg *config.Config

// SetRuntime is called once on boot (from server.New or main).
func SetRuntime(cfg *config.Config) { appCfg = cfg }

// Runtime returns the current runtime config for handlers and other packages.
func Runtime() *config.Config { return appCfg }
