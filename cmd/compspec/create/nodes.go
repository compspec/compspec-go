package create

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/compspec/compspec-go/pkg/graph"
	"github.com/compspec/compspec-go/pkg/utils"
	ep "github.com/compspec/compspec-go/plugins/extractors"
)

// Nodes will read in one or more node extraction metadata files and generate a single nodes JGF graph
// This is intentended for a registration command.
// TODO this should be converted to a creation (converter) plugin
func Nodes(nodesDir, clusterName, nodeOutFile string) error {

	// Read in each node into a plugins.Result
	// 	Results map[string]extractor.ExtractorData `json:"extractors,omitempty"`
	nodes := map[string]ep.Result{}

	nodeFiles, err := os.ReadDir(nodesDir)
	if err != nil {
		return err
	}
	for _, f := range nodeFiles {
		fmt.Printf("Loading %s\n", f.Name())
		result := ep.Result{}
		fullpath := filepath.Join(nodesDir, f.Name())

		// Be forgiving if extra files are there...
		err := result.Load(fullpath)
		if err != nil {
			fmt.Printf("Warning, filename %s is not in the correct format. Skipping\n", f.Name())
			continue
		}
		// Add to nodes, if we don't error
		nodes[f.Name()] = result
	}

	// When we get here, no nodes, no graph
	if len(nodes) == 0 {
		fmt.Println("There were no nodes for the graph.")
		return nil
	}

	// Prepare a graph that will describe our cluster
	g, err := graph.NewClusterGraph(clusterName)
	if err != nil {
		return err
	}

	// This is the root node, we reference it as a parent to the rack
	root := g.Graph.Nodes["0"]

	// Right now assume we have just one rack with all nodes
	// https://github.com/flux-framework/flux-sched/blob/master/t/data/resource/jgfs/tiny.json#L4
	// Note that these are flux specific, and we can make them more generic if needed

	// resource (e.g., rack, node)
	// name (usually the same as the resource)
	// size (usually 1)
	// exclusive (usually false)
	// unit (usually empty or an amount)
	rack := *g.AddNode("rack", "rack", 1, false, "")

	// Connect the rack to the parent, both ways.
	// I think this is because fluxion is Depth First and Upwards (dfu)
	// "The root cluster contains a rack"
	g.AddEdge(root, rack, "contains")

	// "The rack is in a cluster"
	g.AddEdge(rack, root, "in")

	// Read in each node and add to the rack.
	// There are several levels here:
	// /tiny0/rack0/node0/socket0/core1
	for nodeFile, meta := range nodes {

		// We must have extractors, nfd, and sections
		nfd, ok := meta.Results["nfd"]
		if !ok || len(nfd.Sections) == 0 {
			fmt.Printf("node %s is missing extractors->nfd data, skipping\n", nodeFile)
			continue
		}

		// We also need system -> sections -> processor
		system, ok := meta.Results["system"]
		if !ok || len(system.Sections) == 0 {
			fmt.Printf("node %s is missing extractors->system data, skipping\n", nodeFile)
			continue
		}
		processor, ok := system.Sections["processor"]
		if !ok || len(processor) == 0 {
			fmt.Printf("node %s is missing extractors->system->processor, skipping\n", nodeFile)
			continue
		}
		cpu, ok := system.Sections["cpu"]
		if !ok || len(cpu) == 0 {
			fmt.Printf("node %s is missing extractors->system->cpu, skipping\n", nodeFile)
			continue
		}

		// IMPORTANT: this is runtime nproces, which might be physical and virtual
		// we need hwloc for just physical I think
		cores, ok := cpu["cores"]
		if !ok {
			fmt.Printf("node %s is missing extractors->system->cpu->cores, skipping\n", nodeFile)
			continue
		}
		cpuCount, err := strconv.Atoi(cores)
		if err != nil {
			fmt.Printf("node %s cannot convert cores, skipping\n", nodeFile)
			continue
		}

		// First add the rack -> node
		node := *g.AddNode("node", "node", 1, false, "")
		g.AddEdge(rack, node, "contains")
		g.AddEdge(node, rack, "in")

		// Now add the socket. We need hwloc for this
		// nfd has a socket count, but we can't be sure which CPU are assigned to which?
		// This isn't good enough, see https://github.com/compspec/compspec-go/issues/19
		// For the prototype we will use the nfd socket count and split cores across it
		// cpu metadata from ndf
		socketCount := 1

		nfdCpu, ok := nfd.Sections["cpu"]
		if ok {
			sockets, ok := nfdCpu["topology.socket_count"]
			if ok {
				sCount, err := strconv.Atoi(sockets)
				if err == nil {
					socketCount = sCount
				}
			}
		}

		// Get the processors, assume we divide between the sockets
		// TODO we should also get this in better detail, physical vs logical cores
		items := []string{}
		for i := 0; i < cpuCount; i++ {
			items = append(items, fmt.Sprintf("%s", i))
		}
		// Mapping of socket to cores
		chunks := utils.Chunkify(items, socketCount)
		for _, chunk := range chunks {

			// Create each socket attached to the node
			// rack -> node -> socket
			socketNode := *g.AddNode("socket", "socket", 1, false, "")
			g.AddEdge(node, socketNode, "contains")
			g.AddEdge(socketNode, node, "in")

			// Create each core attached to the socket
			for _, _ = range chunk {
				coreNode := *g.AddNode("core", "core", 1, false, "")
				g.AddEdge(socketNode, coreNode, "contains")
				g.AddEdge(coreNode, socketNode, "in")

			}
		}
	}

	// Save graph if given a file
	if nodeOutFile != "" {
		err = g.SaveGraph(nodeOutFile)
		if err != nil {
			return err
		}
	} else {
		toprint, _ := json.MarshalIndent(g.Graph, "", "\t")
		fmt.Println(string(toprint))
		return nil
	}
	return nil
}
