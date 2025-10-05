package main

import "strings"

// isStdPkgPath checks if a package path is standard library (shared utility)
func isStdPkgPath(path string) bool {
    if path == "main" {
        return false
    }
    if strings.Contains(path, ".") {
        return false
    }
    if strings.Contains(path, "/") {
        parts := strings.Split(path, "/")
        if len(parts) >= 2 {
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
                    return true
                }
            }
        }
        return false
    }
    return true
}