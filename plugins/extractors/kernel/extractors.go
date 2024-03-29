package kernel

import (
	"fmt"
	"os"
	"strings"

	"github.com/compspec/compspec-go/pkg/plugin"
	"github.com/compspec/compspec-go/pkg/utils"
	kernelParser "github.com/moby/moby/pkg/parsers/kernel"
)

// Locations to get information about the kernel
const (
	// Parameters given at boot time
	kernelBootFile = "/proc/cmdline"

	// Full boot configuration (key value pairs) prefix for a kernel version
	// This is not usually present in containers, an empty directory
	kernelConfigPrefix = "/boot/config-"

	// Directory with metadata about kernel modules (drivers/versions/params)!
	kernelModules = "/sys/module"
)

// getKernelBootParams loads parameters given to the kernel at boot time
func getKernelBootParams() (plugin.PluginSection, error) {

	raw, err := os.ReadFile(kernelBootFile)
	if err != nil {
		return nil, err
	}

	args := strings.Split(strings.TrimSpace(string(raw)), " ")
	return utils.SplitDelimiterList(args, "=")
}

// getKernelBootConfig loads key value pairs from the kernel config
func getKernelBootConfig() (plugin.PluginSection, error) {

	version, err := kernelParser.GetKernelVersion()
	if err != nil {
		return nil, err
	}

	// What about other files in this directory (older or not active versions?)
	configPath := fmt.Sprintf("/boot/config-%s", version)
	return utils.ParseConfigFile(configPath, "#", "=")
}

// getKernelModules flattens the list of kernel modules (drivers) into
// the name (and if enabled) and version. I don't know if we need more than that.
func getKernelModules() (plugin.PluginSection, error) {
	version, err := kernelParser.GetKernelVersion()
	if err != nil {
		return nil, err
	}

	// The directories in this folder are the modules!
	moduleDirs, err := os.ReadDir(kernelModules)
	if err != nil {
		return nil, err
	}

	// modules is a flattened list of key values pair, for each:
	// module.<name> = <version>
	// module.parameter.<param> = value
	// TODO will this work?
	modules := plugin.PluginSection{}
	for _, moduleDir := range moduleDirs {

		// Don't look unless it's a directory
		if !moduleDir.IsDir() {
			continue
		}

		// Get the name, and then we can create a module
		// This parses the version and name
		moduleName := moduleDir.Name()
		module := NewModule(moduleName, version.String())

		// This can error, and we'd want to know about it
		err := module.SetParameters()
		if err != nil {
			return nil, err
		}
		// Add module paramters to our data
		modules[module.Key()] = module.Version
		for param, value := range module.Parameters {
			moduleParam := fmt.Sprintf("%s.parameter.%s", module.Key(), param)
			modules[moduleParam] = value
		}
	}
	return modules, nil
}
