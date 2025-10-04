package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// Shared types moved from main package
type CallGraphType string

const (
	CallGraphTypeStatic CallGraphType = "static"
	CallGraphTypeCha    CallGraphType = "cha"
	CallGraphTypeRta    CallGraphType = "rta"
)

type renderOpts struct {
	cacheDir string
	focus    string
	group    []string
	ignore   []string
	include  []string
	limit    []string
	nointer  bool
	refresh  bool
	nostd    bool
	algo     CallGraphType
}

type analysis struct {
	opts      *renderOpts
	prog      *ssa.Program
	pkgs      []*ssa.Package
	mainPkg   *ssa.Package
	callgraph *callgraph.Graph
}

// MCPCallgraphRequest represents the input parameters for the callgraph tool via MCP
type MCPCallgraphRequest struct {
	ModuleArgs []string `json:"moduleArgs"`           // Package/module arguments (e.g., "./...")
	Dir        string   `json:"dir,omitempty"`        // Working directory
	Focus      string   `json:"focus,omitempty"`      // Focus specific package
	Group      string   `json:"group,omitempty"`      // Grouping functions by packages and/or types
	Limit      string   `json:"limit,omitempty"`      // Limit package paths to given prefixes
	Ignore     string   `json:"ignore,omitempty"`     // Ignore package paths containing given prefixes
	Include    string   `json:"include,omitempty"`    // Include package paths with given prefixes
	NoStd      bool     `json:"nostd,omitempty"`      // Omit calls to/from packages in standard library
	NoInter    bool     `json:"nointer,omitempty"`    // Omit calls to unexported functions
	Tests      bool     `json:"tests,omitempty"`      // Include test code
	Algo       string   `json:"algo,omitempty"`       // Algorithm: static, cha, rta
	Tags       []string `json:"tags,omitempty"`       // Build tags
	Debug      bool     `json:"debug,omitempty"`      // Enable verbose log
}

// MCPCallgraphResponse represents the output of the callgraph tool via MCP
type MCPCallgraphResponse struct {
	Algorithm string                 `json:"algorithm"`
	Focus     *string                `json:"focus"`
	Filters   MCPCallgraphFilters    `json:"filters"`
	Stats     MCPCallgraphStats      `json:"stats"`
	Graph     MCPCallgraphData       `json:"graph"`
	Error     string                 `json:"error,omitempty"`
}

type MCPCallgraphFilters struct {
	Limit   []string `json:"limit"`
	Ignore  []string `json:"ignore"`
	Include []string `json:"include"`
	NoStd   bool     `json:"nostd"`
	NoInter bool     `json:"nointer"`
	Group   []string `json:"group"`
}

type MCPCallgraphStats struct {
	NodeCount   int `json:"nodeCount"`
	EdgeCount   int `json:"edgeCount"`
	DurationMs  int `json:"durationMs"`
}

type MCPCallgraphData struct {
	Nodes []MCPCallgraphNode `json:"nodes"`
	Edges []MCPCallgraphEdge `json:"edges"`
}

type MCPCallgraphNode struct {
	ID           string  `json:"id"`
	Func         string  `json:"func"`
	PackagePath  string  `json:"packagePath"`
	PackageName  string  `json:"packageName"`
	File         string  `json:"file"`
	Line         int     `json:"line"`
	IsStd        bool    `json:"isStd"`
	Exported     bool    `json:"exported"`
	ReceiverType *string `json:"receiverType"`
}

type MCPCallgraphEdge struct {
	Caller    string `json:"caller"`
	Callee    string `json:"callee"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Synthetic bool   `json:"synthetic"`
}

// HandleCallgraphRequest processes the MCP callgraph request
func HandleCallgraphRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	
	// Parse the arguments
	var req MCPCallgraphRequest
	
	// Convert arguments to JSON bytes first
	argsBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling arguments: %v", err)),
			},
			IsError: true,
		}, nil
	}
	
	if err := json.Unmarshal(argsBytes, &req); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error parsing arguments: %v", err)),
			},
			IsError: true,
		}, nil
	}

	// Validate required parameters
	if len(req.ModuleArgs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: moduleArgs is required"),
			},
			IsError: true,
		}, nil
	}

	// Set defaults
	if req.Algo == "" {
		req.Algo = "static"
	}
	if req.Group == "" {
		req.Group = "pkg"
	}

	// Map MCP request to internal analysis options
	opts := mapMCPRequestToRenderOpts(req)
	
	// Set up build tags if provided
	if len(req.Tags) > 0 {
		build.Default.BuildTags = req.Tags
	}

	// Initialize analysis
	analysis := &analysis{opts: opts}
	
	// Perform analysis
	algo := CallGraphType(req.Algo)
	if err := analysis.DoAnalysis(algo, req.Dir, req.Tests, req.ModuleArgs); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Analysis failed: %v", err)),
			},
			IsError: true,
		}, nil
	}

	// Process list arguments (comma-separated strings to slices)
	if err := analysis.ProcessListArgs(); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error processing arguments: %v", err)),
			},
			IsError: true,
		}, nil
	}

	// Generate JSON output instead of DOT
	jsonData, stats, err := generateJSONCallgraph(analysis)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error generating callgraph: %v", err)),
			},
			IsError: true,
		}, nil
	}

	// Calculate duration
	duration := time.Since(start)
	stats.DurationMs = int(duration.Milliseconds())

	// Build response
	response := MCPCallgraphResponse{
		Algorithm: req.Algo,
		Focus:     nil,
		Filters: MCPCallgraphFilters{
			Limit:   analysis.opts.limit,
			Ignore:  analysis.opts.ignore,
			Include: analysis.opts.include,
			NoStd:   analysis.opts.nostd,
			NoInter: analysis.opts.nointer,
			Group:   analysis.opts.group,
		},
		Stats: stats,
		Graph: jsonData,
	}

	if req.Focus != "" {
		response.Focus = &req.Focus
	}

	// Convert response to JSON
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling response: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(responseJSON)),
		},
	}, nil
}

// mapMCPRequestToRenderOpts converts MCP request to internal renderOpts
func mapMCPRequestToRenderOpts(req MCPCallgraphRequest) *renderOpts {
	return &renderOpts{
		cacheDir: "", // Not used in MCP mode
		focus:    req.Focus,
		group:    []string{req.Group},
		ignore:   []string{req.Ignore},
		include:  []string{req.Include},
		limit:    []string{req.Limit},
		nointer:  req.NoInter,
		refresh:  false, // Not used in MCP mode
		nostd:    req.NoStd,
		algo:     CallGraphType(req.Algo),
	}
}

// Helper functions moved from main package
var debugFlag = &[]bool{false}[0] // Default to false

// logf is a simple logging function
func logf(format string, args ...interface{}) {
	if *debugFlag {
		log.Printf(format, args...)
	}
}

// mainPackages returns the main packages to analyze.
func mainPackages(pkgs []*ssa.Package) ([]*ssa.Package, error) {
	var mains []*ssa.Package
	for _, p := range pkgs {
		if p != nil && p.Pkg.Name() == "main" && p.Func("main") != nil {
			mains = append(mains, p)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages")
	}
	return mains, nil
}

// initFuncs returns all package init functions
func initFuncs(pkgs []*ssa.Package) ([]*ssa.Function, error) {
	var inits []*ssa.Function
	for _, p := range pkgs {
		if p == nil {
			continue
		}
		for name, member := range p.Members {
			fun, ok := member.(*ssa.Function)
			if !ok {
				continue
			}
			if name == "init" || strings.HasPrefix(name, "init#") {
				inits = append(inits, fun)
			}
		}
	}
	return inits, nil
}

func (a *analysis) DoAnalysis(
	algo CallGraphType,
	dir string,
	tests bool,
	args []string,
) error {
	logf("begin analysis")
	defer logf("analysis done")

	cfg := &packages.Config{
		Mode:       packages.LoadAllSyntax,
		Tests:      tests,
		Dir:        dir,
		BuildFlags: getBuildFlags(),
	}

	logf("loading packages")

	initial, err := packages.Load(cfg, args...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(initial) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	logf("loaded %d initial packages, building program", len(initial))

	// Create and build SSA-form program representation.
	mode := ssa.InstantiateGenerics
	prog, pkgs := ssautil.AllPackages(initial, mode)
	
	prog.Build()

	logf("build done, computing callgraph (algo: %v)", algo)

	var graph *callgraph.Graph
	var mainPkg *ssa.Package

	switch algo {
	case CallGraphTypeStatic:
		graph = static.CallGraph(prog)
	case CallGraphTypeCha:
		graph = cha.CallGraph(prog)
	case CallGraphTypeRta:
		mains, err := mainPackages(prog.AllPackages())
		if err != nil {
			return err
		}
		var roots []*ssa.Function
		mainPkg = mains[0]
		for _, main := range mains {
			roots = append(roots, main.Func("main"))
		}

		inits, err := initFuncs(prog.AllPackages())
		if err != nil {
			return err
		}
		for _, init := range inits {
			roots = append(roots, init)
		}

		graph = rta.Analyze(roots, true).CallGraph
	default:
		return fmt.Errorf("invalid call graph type: %s", algo)
	}

	logf("callgraph resolved with %d nodes", len(graph.Nodes))

	a.prog = prog
	a.pkgs = pkgs
	a.mainPkg = mainPkg
	a.callgraph = graph
	return nil
}

func (a *analysis) ProcessListArgs() (e error) {
	var groupBy []string
	var ignorePaths []string
	var includePaths []string
	var limitPaths []string

	for _, g := range strings.Split(a.opts.group[0], ",") {
		g := strings.TrimSpace(g)
		if g == "" {
			continue
		}
		if g != "pkg" && g != "type" {
			e = errors.New("invalid group option")
			return
		}
		groupBy = append(groupBy, g)
	}

	for _, p := range strings.Split(a.opts.ignore[0], ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			ignorePaths = append(ignorePaths, p)
		}
	}

	for _, p := range strings.Split(a.opts.include[0], ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			includePaths = append(includePaths, p)
		}
	}

	for _, p := range strings.Split(a.opts.limit[0], ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			limitPaths = append(limitPaths, p)
		}
	}

	a.opts.group = groupBy
	a.opts.ignore = ignorePaths
	a.opts.include = includePaths
	a.opts.limit = limitPaths

	return
}

func getBuildFlags() []string {
	buildFlagTags := getBuildFlagTags(build.Default.BuildTags)
	if len(buildFlagTags) == 0 {
		return nil
	}

	return []string{buildFlagTags}
}

func getBuildFlagTags(buildTags []string) string {
	if len(buildTags) > 0 {
		return "-tags=" + strings.Join(buildTags, ",")
	}

	return ""
}

// isSynthetic checks if an edge is synthetic
func isSynthetic(edge *callgraph.Edge) bool {
	return edge.Caller.Func.Pkg == nil || edge.Callee.Func.Pkg == nil || edge.Callee.Func.Synthetic != ""
}

// inStd checks if a node is in standard library
func inStd(node *callgraph.Node) bool {
	return isStdPkgPath(node.Func.Pkg.Pkg.Path())
}

// isStdPkgPath checks if a package path is standard library
func isStdPkgPath(path string) bool {
	if strings.Contains(path, ".") {
		return false
	}
	return true
}

// generateJSONCallgraph generates JSON callgraph data instead of DOT format
func generateJSONCallgraph(a *analysis) (MCPCallgraphData, MCPCallgraphStats, error) {
	var nodes []MCPCallgraphNode
	var edges []MCPCallgraphEdge
	
	nodeMap := make(map[string]*MCPCallgraphNode)
	edgeMap := make(map[string]*MCPCallgraphEdge)

	// Get focus package if specified
	var focusPkg *types.Package
	if a.opts.focus != "" {
		if ssaPkg := a.prog.ImportedPackage(a.opts.focus); ssaPkg != nil {
			focusPkg = ssaPkg.Pkg
		} else {
			// Try to find package by name
			for _, p := range a.pkgs {
				if p.Pkg.Name() == a.opts.focus {
					if ssaPkg := a.prog.ImportedPackage(p.Pkg.Path()); ssaPkg != nil {
						focusPkg = ssaPkg.Pkg
						break
					}
				}
			}
		}
	}

	// Delete synthetic nodes
	a.callgraph.DeleteSyntheticNodes()

	// Define filter functions
	var isFocused = func(edge *callgraph.Edge) bool {
		caller := edge.Caller
		callee := edge.Callee
		if focusPkg != nil && (caller.Func.Pkg.Pkg.Path() == focusPkg.Path() || callee.Func.Pkg.Pkg.Path() == focusPkg.Path()) {
			return true
		}
		fromFocused := false
		for _, e := range caller.In {
			if !isSynthetic(e) && focusPkg != nil && e.Caller.Func.Pkg.Pkg.Path() == focusPkg.Path() {
				fromFocused = true
				break
			}
		}
		toFocused := false
		for _, e := range callee.Out {
			if !isSynthetic(e) && focusPkg != nil && e.Callee.Func.Pkg.Pkg.Path() == focusPkg.Path() {
				toFocused = true
				break
			}
		}
		return fromFocused && toFocused
	}

	var inIncludes = func(node *callgraph.Node) bool {
		pkgPath := node.Func.Pkg.Pkg.Path()
		for _, p := range a.opts.include {
			if strings.HasPrefix(pkgPath, p) {
				return true
			}
		}
		return false
	}

	var inLimits = func(node *callgraph.Node) bool {
		pkgPath := node.Func.Pkg.Pkg.Path()
		for _, p := range a.opts.limit {
			if strings.HasPrefix(pkgPath, p) {
				return true
			}
		}
		return false
	}

	var inIgnores = func(node *callgraph.Node) bool {
		pkgPath := node.Func.Pkg.Pkg.Path()
		for _, p := range a.opts.ignore {
			if strings.Contains(pkgPath, p) {
				return true
			}
		}
		return false
	}

	// Process all edges
	for _, node := range a.callgraph.Nodes {
		for _, edge := range node.Out {
			caller := edge.Caller
			callee := edge.Callee

			// Skip synthetic edges
			if isSynthetic(edge) {
				continue
			}

			// Apply filters
			if a.opts.nostd && (inStd(caller) || inStd(callee)) {
				continue
			}

			if a.opts.nointer && (!caller.Func.Object().Exported() || !callee.Func.Object().Exported()) {
				continue
			}

			if len(a.opts.include) > 0 && !(inIncludes(caller) || inIncludes(callee)) {
				continue
			}

			if len(a.opts.limit) > 0 && !(inLimits(caller) || inLimits(callee)) {
				continue
			}

			if len(a.opts.ignore) > 0 && (inIgnores(caller) || inIgnores(callee)) {
				continue
			}

			if focusPkg != nil && !isFocused(edge) {
				continue
			}

			// Create nodes if they don't exist
			callerID := fmt.Sprintf("%s", caller.Func)
			calleeID := fmt.Sprintf("%s", callee.Func)

			if _, exists := nodeMap[callerID]; !exists {
				pos := a.prog.Fset.Position(caller.Func.Pos())
				nodeMap[callerID] = createJSONNode(caller, pos)
			}

			if _, exists := nodeMap[calleeID]; !exists {
				pos := a.prog.Fset.Position(callee.Func.Pos())
				nodeMap[calleeID] = createJSONNode(callee, pos)
			}

			// Create edge
			edgeID := fmt.Sprintf("%s->%s", callerID, calleeID)
			if _, exists := edgeMap[edgeID]; !exists {
				pos := a.prog.Fset.Position(edge.Pos())
				edgeMap[edgeID] = &MCPCallgraphEdge{
					Caller:    callerID,
					Callee:    calleeID,
					File:      pos.Filename,
					Line:      pos.Line,
					Synthetic: isSynthetic(edge),
				}
			}
		}
	}

	// Convert maps to slices
	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}
	for _, edge := range edgeMap {
		edges = append(edges, *edge)
	}

	stats := MCPCallgraphStats{
		NodeCount: len(nodes),
		EdgeCount: len(edges),
	}

	data := MCPCallgraphData{
		Nodes: nodes,
		Edges: edges,
	}

	return data, stats, nil
}

func createJSONNode(node *callgraph.Node, pos token.Position) *MCPCallgraphNode {
	fn := node.Func
	pkg := fn.Pkg.Pkg
	
	var receiverType *string
	if fn.Signature.Recv() != nil {
		recvType := fn.Signature.Recv().Type().String()
		receiverType = &recvType
	}

	return &MCPCallgraphNode{
		ID:           fmt.Sprintf("%s", fn),
		Func:         fn.Name(),
		PackagePath:  pkg.Path(),
		PackageName:  pkg.Name(),
		File:         filepath.Base(pos.Filename),
		Line:         pos.Line,
		IsStd:        isStdPkgPath(pkg.Path()),
		Exported:     fn.Object() != nil && fn.Object().Exported(),
		ReceiverType: receiverType,
	}
}