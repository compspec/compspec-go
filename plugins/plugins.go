package plugins

import (
	"strings"

	"github.com/supercontainers/compspec-go/plugins/extractors/kernel"
	"github.com/supercontainers/compspec-go/plugins/extractors/system"
)

// Add new plugin names here. They should correspond with the package name, then NewPlugin()
var (
	KernelExtractor = "kernel"
	SystemExtractor = "system"
	pluginNames     = []string{KernelExtractor, SystemExtractor}
)

// parseSections will return sections from the name string
// We could use regex here instead
func parseSections(raw string) (string, []string) {

	sections := []string{}

	// If no opening bracker, it's just the name
	if !strings.Contains(raw, "[") {
		return raw, sections
	}
	// Get rid of last piece ]
	raw = strings.ReplaceAll(raw, "]", "")

	// Split into two pieces, the
	parts := strings.SplitN(raw, "[", 2)
	name, raw := parts[0], parts[1]

	// Split sections by comma
	sections = strings.Split(raw, ",")
	return name, sections
}

// Get plugins parses a request and returns a list of plugins
// We honor the order that the plugins and sections are provided in
func GetPlugins(names []string) (PluginsRequest, error) {

	if len(names) == 0 {
		names = pluginNames
	}

	request := PluginsRequest{}

	// Prepare an extractor for each, and validate the requested sections
	// TODO: this could also be done with an init -> Register pattern
	for _, name := range names {

		// If we are given a list of section names, parse.
		name, sections := parseSections(name)

		if strings.HasPrefix(name, KernelExtractor) {
			p, err := kernel.NewPlugin(sections)
			if err != nil {
				return request, err
			}
			// Save the name, the instantiated interface, and sections
			pr := PluginRequest{Name: name, Extractor: p, Sections: sections}
			request = append(request, pr)
		}

		if strings.HasPrefix(name, SystemExtractor) {
			p, err := system.NewPlugin(sections)
			if err != nil {
				return request, err
			}
			// Save the name, the instantiated interface, and sections
			pr := PluginRequest{Name: name, Extractor: p, Sections: sections}
			request = append(request, pr)
		}
	}
	return request, nil
}