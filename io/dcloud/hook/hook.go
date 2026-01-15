package hook

import "sync"


type Resolver interface {
	Enable(fileName string) bool
	ResolveModuleNames(moduleNames []string, containingFile string) []string
}

type hook struct {
	resolvers []Resolver
}
var _hook = sync.OnceValue(func() *hook {
	return &hook{}
})

func RegisterResolver(resolver Resolver) {
	_hook().resolvers = append(_hook().resolvers, resolver)
}

func GetResolver(fileName string) Resolver {
	for _, resolver := range _hook().resolvers {
		if resolver.Enable(fileName) {
			return resolver
		}
	}
	return nil
}