package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kohkimakimoto/loglv"
	"github.com/kohkimakimoto/sshdocker/sshdocker"
	"github.com/pkg/errors"
)

func main() {
	os.Exit(realMain())
}

func realMain() (status int) {
	defer func() {
		if err := recover(); err != nil {
			printError(err)
			status = 1
		}
	}()
	// parse flags...
	var optConfig string
	var optVersion bool

	flag.StringVar(&optConfig, "config-file", "", "")
	flag.StringVar(&optConfig, "c", "", "")
	flag.BoolVar(&optVersion, "v", false, "")
	flag.BoolVar(&optVersion, "version", false, "")

	flag.Usage = func() {
		fmt.Println(`Usage: ` + sshdocker.Name + ` [OPTIONS...]

` + sshdocker.Name + ` : Run docker containers over SSH.
version ` + sshdocker.Version + ` (` + sshdocker.CommitHash + `)

Options:
  -c, -config-file=FILE    Load configuration from the file.
  -h, -help                Show help.
  -v, -version             Print the version.

See: https://github.com/kohkimakimoto/sshdocker for updates, code and issues.
`)
	}
	flag.Parse()

	if optVersion {
		// show version
		fmt.Println(sshdocker.Name + " version " + sshdocker.Version + " (" + sshdocker.CommitHash + ")")
		return 0
	}

	// config
	c := sshdocker.NewConfig()

	// load config file path from a environment variable.
	if optConfig == "" {
		if confpath := os.Getenv("SSHDOCKER_CONFIG_PATH"); confpath != "" {
			optConfig = confpath
		}
	}

	if optConfig != "" {
		if err := c.LoadConfigFile(optConfig); err != nil {
			printError(errors.Wrapf(err, "failed to load config from the file '%s'", c))
			return 1
		}
	}

	loglv.Init()

	if lv := os.Getenv("SSHDOCKER_DEBUG"); lv != "" {
		c.Debug = true
	}

	if c.Debug {
		loglv.SetLv(loglv.LvDebug)
	}

	// log.Printf("Loaded configuration from: %s", optConfig)

	s := sshdocker.NewServer(c)
	defer s.Close()

	if err := s.Run(); err != nil {
		printError(err)
		return 1
	}

	return status
}

func printError(err interface{}) {
	fmt.Fprintf(os.Stderr, "[error] "+sshdocker.Name+" aborted!\n")
	fmt.Fprintf(os.Stderr, "[error] %v\n", err)
}
