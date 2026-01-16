package hook

import (
	"sync"

	"github.com/microsoft/typescript-go/internal/core"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)


type Resolver interface {
	dis.Disposable
	Enable(fileName string) bool
	ResolveModuleName(moduleName string, containingFile string, resolutionMode core.ResolutionMode, redirectedReference any) (string, string, bool, bool)
}

type hook struct {
	resolvers []*dis.Box[Resolver]
}
var _hook = sync.OnceValue(func() *hook {
	return &hook{}
})

// func RegisterResolver(resolver Resolver) {
// 	_hook().resolvers = append(_hook().resolvers, resolver)
// }

func GetResolver(fileName string) []Resolver {
	var resolvers []Resolver
	var refResolvers []*dis.Box[Resolver]
	hk := _hook()
	for _, resolver := range hk.resolvers {
		if(resolver == nil || resolver.Value() == nil) {
			continue
		}
		resolvers = append(resolvers, resolver.Value())
		refResolvers = append(refResolvers, resolver)
	}
	hk.resolvers = refResolvers
	
	return resolvers
}