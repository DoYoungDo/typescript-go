// package hook

// import (
// 	"sync"

// 	"github.com/microsoft/typescript-go/internal/core"
// 	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
// )

// type Resolver interface {
// 	dis.Disposable
// 	Enable(fileName string) bool
// 	ResolveModuleName(moduleName string, containingFile string, resolutionMode core.ResolutionMode, redirectedReference any) (string, string, bool, bool)
// }

// type Checker interface {
// 	dis.Disposable
// 	// Enable(fileName string) bool
// 	CheckExpression(expression any) (any, bool)
// }

// type hook struct {
// 	resolvers []*dis.Box[Resolver]
// 	checkers []*dis.Box[Checker]
// }
// var _hook = sync.OnceValue(func() *hook {
// 	return &hook{}
// })

// func RegisterResolver(resolver Resolver) {
// 	_hook().resolvers = append(_hook().resolvers, resolver)
// }

// func GetResolver(fileName string) []*dis.Box[Resolver] {
// 	hk := _hook()
// 	var resolvers []*dis.Box[Resolver]
// 	for _, resolver := range hk.resolvers {
// 		if resolver.Value() == nil || !resolver.Value().Enable(fileName){
// 			continue
// 		}
// 		resolvers = append(resolvers, resolver)
// 	}
// 	return resolvers
// }

// func GetCheckers() []*dis.Box[Checker] {
// 	hk := _hook()
// 	var checkers []*dis.Box[Checker]
// 	for _, checker := range hk.checkers {
// 		if checker.Value() == nil {
// 			continue
// 		}
// 		checkers = append(checkers, checker)
// 	}
// 	return checkers
// }
