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

	"github.com/emicklei/dot"
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

	// Set debug flag
	if req.Debug {
		*debugFlag = true
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

	// Generate Mermaid output instead of JSON
	mermaidCode, stats, err := generateMermaidCallgraph(analysis)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error generating callgraph: %v", err)),
			},
			IsError: true,
		}, nil
	}

	// Calculate duration (optional usage)
	duration := time.Since(start)
	stats.DurationMs = int(duration.Milliseconds())

	// Return Mermaid flowchart code directly
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(mermaidCode),
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
	// main package should not be considered standard library
	if path == "main" {
		return false
	}
	// Packages with dots are typically user packages (e.g., github.com/user/repo)
	if strings.Contains(path, ".") {
		return false
	}
	// Packages with slashes are typically user packages (e.g., user/repo)
	// BUT standard library also has subpackages like io/fs, math/bits
	if strings.Contains(path, "/") {
		// Check if it's a standard library subpackage
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			// Common standard library top-level packages
			stdPkgs := []string{
				"archive", "bufio", "builtin", "bytes", "compress", "container",
				"context", "crypto", "database", "debug", "embed", "encoding",
				"errors", "expvar", "flag", "fmt", "go", "hash", "html", "image",
				"index", "io", "log", "math", "mime", "net", "os", "path",
				"plugin", "reflect", "regexp", "runtime", "sort", "strconv",
				"strings", "sync", "syscall", "testing", "text", "time",
				"unicode", "unsafe",
			}
			for _, stdPkg := range stdPkgs {
				if parts[0] == stdPkg {
					return true // It's a standard library subpackage
				}
			}
		}
		return false // User package with slash
	}
	// Standard library packages (single word without dots or slashes)
	return true
}

// isInternalPkg checks if a package is an internal runtime package
func isInternalPkg(path string) bool {
	return strings.HasPrefix(path, "internal/") || 
		   strings.Contains(path, "/internal/") ||
		   path == "runtime" ||
		   strings.HasPrefix(path, "runtime/") ||
		   path == "sync" ||
		   strings.HasPrefix(path, "sync/")
}

// generateMermaidCallgraph builds a DOT graph using emicklei/dot and returns Mermaid flowchart code
func generateMermaidCallgraph(a *analysis) (string, MCPCallgraphStats, error) {
	var stats MCPCallgraphStats

	// Build a DOT graph (directed)
	g := dot.NewGraph(dot.Directed)
	g.Attr("label", "callgraph")

	// Helper maps
	nodeMap := make(map[string]*MCPCallgraphNode)
	edgeMap := make(map[string]*MCPCallgraphEdge)
	dotNodes := make(map[string]dot.Node)

	// Get focus package if specified
	var focusPkg *types.Package
	if a.opts.focus != "" {
		if ssaPkg := a.prog.ImportedPackage(a.opts.focus); ssaPkg != nil {
			focusPkg = ssaPkg.Pkg
		} else {
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

	// Filter helpers
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

	// Traverse callgraph and populate nodes/edges
	for _, n := range a.callgraph.Nodes {
		for _, e := range n.Out {
			caller := e.Caller
			callee := e.Callee

			if isSynthetic(e) {
				continue
			}
			
			if a.opts.nostd && (inStd(caller) || inStd(callee)) {
				continue
			}
			
			// Also filter internal runtime packages when nostd is true
			if a.opts.nostd && (isInternalPkg(caller.Func.Pkg.Pkg.Path()) || isInternalPkg(callee.Func.Pkg.Pkg.Path())) {
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
			if focusPkg != nil && !isFocused(e) {
				continue
			}

			callerID := fmt.Sprintf("%s", caller.Func)
			calleeID := fmt.Sprintf("%s", callee.Func)

			if _, ok := nodeMap[callerID]; !ok {
				pos := a.prog.Fset.Position(caller.Func.Pos())
				nodeMap[callerID] = createJSONNode(caller, pos)
			}
			if _, ok := nodeMap[calleeID]; !ok {
				pos := a.prog.Fset.Position(callee.Func.Pos())
				nodeMap[calleeID] = createJSONNode(callee, pos)
			}

			// Create DOT nodes
			dFromID := sanitizeMermaidID(callerID)
			dToID := sanitizeMermaidID(calleeID)
			if _, exists := dotNodes[dFromID]; !exists {
				dotNodes[dFromID] = g.Node(dFromID)
			}
			if _, exists := dotNodes[dToID]; !exists {
				dotNodes[dToID] = g.Node(dToID)
			}
			g.Edge(dotNodes[dFromID], dotNodes[dToID])

			edgeID := fmt.Sprintf("%s->%s", callerID, calleeID)
			if _, exists := edgeMap[edgeID]; !exists {
				pos := a.prog.Fset.Position(e.Pos())
				edgeMap[edgeID] = &MCPCallgraphEdge{
					Caller:    callerID,
					Callee:    calleeID,
					File:      pos.Filename,
					Line:      pos.Line,
					Synthetic: isSynthetic(e),
				}
			}
		}
	}

	// Build Mermaid flowchart text
	var sb strings.Builder
	// Direction: Left-to-Right (LR). Could be configurable.
	sb.WriteString("flowchart LR\n")

	// Determine grouping options
	hasPkg := false
	hasType := false
	for _, g := range a.opts.group {
		if g == "pkg" {
			hasPkg = true
		}
		if g == "type" {
			hasType = true
		}
	}

	// Helper to write a single node line and its file:line comment
	writeNode := func(id string, n *MCPCallgraphNode) {
		mid := sanitizeMermaidID(id)
		label := fmt.Sprintf("%s<br/>%s", n.Func, n.PackageName)
		sb.WriteString(fmt.Sprintf("%s[%q]\n", mid, label))
		// Append file:line as a Mermaid comment
		sb.WriteString(fmt.Sprintf("%%%% %s:%d\n", n.File, n.Line))
	}

	if hasPkg && hasType {
		// Nested grouping: pkg -> type -> nodes
		nested := make(map[string]map[string][]string)
		for id, n := range nodeMap {
			pkg := n.PackagePath
			typ := "func"
			if n.ReceiverType != nil && *n.ReceiverType != "" {
				typ = *n.ReceiverType
			}
			if _, ok := nested[pkg]; !ok {
				nested[pkg] = make(map[string][]string)
			}
			nested[pkg][typ] = append(nested[pkg][typ], id)
		}
		for pkg, typeMap := range nested {
			// Subgraph per package
			sb.WriteString(fmt.Sprintf("subgraph %q\n", "pkg:"+pkg))
			for typ, ids := range typeMap {
				// Subgraph per type within package
				sb.WriteString(fmt.Sprintf("subgraph %q\n", "type:"+typ))
				for _, id := range ids {
					writeNode(id, nodeMap[id])
				}
				sb.WriteString("end\n")
			}
			sb.WriteString("end\n")
		}
	} else if hasPkg {
		// Group by package only
		groups := make(map[string][]string)
		for id, n := range nodeMap {
			groups[n.PackagePath] = append(groups[n.PackagePath], id)
		}
		for pkg, ids := range groups {
			sb.WriteString(fmt.Sprintf("subgraph %q\n", "pkg:"+pkg))
			for _, id := range ids {
				writeNode(id, nodeMap[id])
			}
			sb.WriteString("end\n")
		}
	} else if hasType {
		// Group by type (receiver) only
		groups := make(map[string][]string)
		for id, n := range nodeMap {
			typ := "func"
			if n.ReceiverType != nil && *n.ReceiverType != "" {
				typ = *n.ReceiverType
			}
			groups[typ] = append(groups[typ], id)
		}
		for typ, ids := range groups {
			sb.WriteString(fmt.Sprintf("subgraph %q\n", "type:"+typ))
			for _, id := range ids {
				writeNode(id, nodeMap[id])
			}
			sb.WriteString("end\n")
		}
	} else {
		// No grouping, declare all nodes at top level
		for id, n := range nodeMap {
			writeNode(id, n)
		}
	}

	// Declare edges
	for _, ed := range edgeMap {
		from := sanitizeMermaidID(ed.Caller)
		to := sanitizeMermaidID(ed.Callee)
		sb.WriteString(fmt.Sprintf("%s --> %s\n", from, to))
	}

	stats.NodeCount = len(nodeMap)
	stats.EdgeCount = len(edgeMap)

	return sb.String(), stats, nil
}

// sanitizeMermaidID creates a safe identifier for Mermaid nodes
func sanitizeMermaidID(s string) string {
	// Replace characters not suitable for Mermaid identifiers
	replacer := strings.NewReplacer(
		" ", "_",
		"\t", "_",
		"\n", "_",
		"\r", "_",
		".", "_",
		"/", "_",
		"-", "_",
		":", "_",
		"(", "_",
		")", "_",
		"*", "_",
		"[", "_",
		"]", "_",
	)
	// Limit length to avoid extremely long ids
	safe := replacer.Replace(s)
	if len(safe) > 128 {
		return safe[:128]
	}
	return safe
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