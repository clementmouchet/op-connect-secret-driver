package main

import (
	"fmt"
	"os"

	"github.com/docker/go-plugins-helpers/secrets"

	"op-connect-secret-driver/internal/driver"
)

var version = "dev"

func main() {
	pluginDriver, err := driver.NewDriver()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to create 1Password Connect Driver: %v\n", err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] starting plugin handler on /run/docker/plugins/opcsd.sock (version %s)\n", version)
	handler := secrets.NewHandler(pluginDriver)
	if err := handler.ServeUnix("/run/docker/plugins/opcsd.sock", 0); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] error serving plugin: %v\n", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] closed\n")
}
