package main

import (
	"cmp"
	"fmt"
	"os"
	"runtime/debug"

	"afastchat.com/config"
	"afastchat.com/logger"
	"afastchat.com/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		printVersion()
		return
	}

	cfg := config.ParseConfig()
	logger.Setup(config.Dev, cfg.LogLevel)
	server.StartServer(cfg)
}

func printVersion() {
	settings := map[string]string{}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			settings[setting.Key] = setting.Value
		}
	}

	commit := cmp.Or(settings["vcs.revision"], "unknown")
	if settings["vcs.modified"] == "true" {
		commit += " (with uncommitted changes)"
	}
	fmt.Printf("commit %s\ndate   %s\n", commit, cmp.Or(settings["vcs.time"], "unknown"))
}
