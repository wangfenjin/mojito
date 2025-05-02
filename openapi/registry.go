package openapi

import (
	"context"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-chi/chi/v5"
)

var handlerFuncs = make(map[string]FuncInfo)
var mws = make(map[string][]string)

func RegisterMws(r chi.Routes) {
	clear(mws)
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		for _, mw := range middlewares {
			mws[method+":"+route] = append(mws[method+":"+route], runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).Name())
		}
		return nil
	})
}

// RegisterHandler adds route information to the registry
func RegisterHandler[Req any, Resp any](method, pattern string, handlerFunc func(context.Context, Req) (Resp, error)) {
	// return if is production env
	if os.Getenv("ENV") == "production" {
		return
	}
	if strings.Contains(pattern, "/test/") {
		return
	}
	if _, ok := handlerFuncs[method+":"+pattern]; ok {
		return
	}
	funcInfo := buildFuncInfo(method, pattern, handlerFunc)
	handlerFuncs[method+":"+pattern] = funcInfo
}

func requireAuth(method, pattern string) bool {
	for _, mw := range mws[method+":"+pattern] {
		if strings.Contains(mw, "RequireAuth") {
			return true
		}
	}
	return false
}

type FuncInfo struct {
	RequestType  reflect.Type
	ResponseType reflect.Type
	Tag          string `json:"tag"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Pkg          string `json:"pkg"`
	Func         string `json:"func"`
	Comment      string `json:"comment"`
	Summary      string `json:"summary"`
	File         string `json:"file,omitempty"`
	Line         int    `json:"line,omitempty"`
	Anonymous    bool   `json:"anonymous,omitempty"`
	Unresolvable bool   `json:"unresolvable,omitempty"`
	RequireAuth  bool   `json:"require_auth,omitempty"`
}

func getGoPath() string {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = build.Default.GOPATH
	}
	return goPath
}
func buildFuncInfo[Req any, Resp any](method, path string, i func(context.Context, Req) (Resp, error)) FuncInfo {
	fi := FuncInfo{}
	fi.Method = method
	fi.Path = path
	// tag defaults to path first part after trim /api/v1/
	fi.Tag = strings.Split(path, "/")[3]
	fi.RequireAuth = requireAuth(method, path)
	fi.RequestType = reflect.TypeOf((*Req)(nil)).Elem()
	fi.ResponseType = reflect.TypeOf((*Resp)(nil)).Elem()
	frame := getCallerFrame(i)
	goPathSrc := filepath.Join(getGoPath(), "src")

	if frame == nil {
		fi.Unresolvable = true
		return fi
	}

	pkgName := getPkgName(frame.File)
	if pkgName == "chi" {
		fi.Unresolvable = true
	}
	funcPath := frame.Func.Name()

	idx := strings.Index(funcPath, "/"+pkgName)
	if idx > 0 {
		fi.Pkg = funcPath[:idx+1+len(pkgName)]
		fi.Func = funcPath[idx+2+len(pkgName):]
	} else {
		fi.Func = funcPath
	}

	if strings.Index(fi.Func, ".func") > 0 {
		fi.Anonymous = true
	}

	fi.File = frame.File
	fi.Line = frame.Line
	if filepath.HasPrefix(fi.File, goPathSrc) {
		fi.File = fi.File[len(goPathSrc)+1:]
	}

	// Check if file info is unresolvable
	if !strings.Contains(funcPath, pkgName) {
		fi.Unresolvable = true
	}

	if !fi.Unresolvable {
		fi.Comment = getFuncComment(frame.File, frame.Line)
		parseComment(&fi)
	}

	return fi
}

func parseComment(fi *FuncInfo) {
	// split by new line and parse the line start with @tag
	lines := strings.Split(fi.Comment, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(line), "@tag") {
			fi.Tag = strings.TrimSpace(line[5:])
		} else if strings.HasPrefix(strings.ToLower(line), "@summary") {
			fi.Summary = strings.TrimSpace(line[9:])
		}
	}
}

func getCallerFrame(i interface{}) *runtime.Frame {
	value := reflect.ValueOf(i)
	if value.Kind() != reflect.Func {
		return nil
	}
	pc := value.Pointer()
	frames := runtime.CallersFrames([]uintptr{pc})
	if frames == nil {
		return nil
	}
	frame, _ := frames.Next()
	if frame.Entry == 0 {
		return nil
	}
	return &frame
}

func getFuncComment(file string, line int) string {
	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		return ""
	}

	if len(astFile.Comments) == 0 {
		fmt.Printf("No comments found in file %s\n", file)
		return ""
	}

	for _, cmt := range astFile.Comments {
		if fset.Position(cmt.End()).Line+1 == line {
			return cmt.Text()
		}
	}

	fmt.Printf("No comment found at line %d in file %s\n", line, file)
	return ""
}

func getPkgName(file string) string {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, file, nil, parser.PackageClauseOnly)
	if err != nil {
		return ""
	}
	if astFile.Name == nil {
		return ""
	}
	return astFile.Name.Name
}
