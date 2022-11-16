package kswitch

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/tjamet/kubectl-switch/pkg/kubectl"
	"github.com/tjamet/kubectl-switch/pkg/server"

	"github.com/spf13/cobra"
	//utilflag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var exit = os.Exit

func run(rg server.RestConfigGetter) {
	v := server.GetVersionFromConfig(rg)
	os := runtime.GOOS
	arch := runtime.GOARCH
	code, err := run_with_version(v, os, arch)
	// If we are in a Mac OS with an arm64 just download the amd64 and use
	// Rosseta.
	if errors.Is(err, kubectl.ErrNoBinaryFound) && os == "darwin" && arch == "arm64" {
		code, err = run_with_version(v, os, "amd64")
	}
	if err != nil {
		fmt.Printf("Failed to download kubectl version %s: %v\n", v, err.Error())
		exit(1)
	}
	exit(code)
}

func run_with_version(v string, osName string, arch string) (int, error) {
	if !kubectl.Installed(v) {
		err := kubectl.Download(v, osName, arch)
		if err != nil {
			return 0, fmt.Errorf("failed to download kubectl version %s: %w", v, err)

		}
	}
	r := kubectl.Exec(v, os.Args[1:]...)
	return r, nil
}

type nopWriter struct{}

func (n nopWriter) Write(a []byte) (int, error) {
	return len(a), nil
}

func Main() {
	cmds := &cobra.Command{}

	flags := cmds.PersistentFlags()
	//flags.SetNormalizeFunc(utilflag.WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	//flags.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)

	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	kubeConfigFlags.AddFlags(flags)
	cmds.Run = func(cmd *cobra.Command, args []string) {
		run(kubeConfigFlags)
	}
	cmds.SetUsageFunc(func(*cobra.Command) error {
		run(kubeConfigFlags)
		return nil
	})
	cmds.SetOutput(nopWriter{})
	cmds.Execute()
}
