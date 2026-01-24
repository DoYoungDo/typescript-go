package dcloud

import "github.com/microsoft/typescript-go/internal/compiler"

type CompilerHost interface{
	compiler.CompilerHost

	SetForeceParseSourceFile(f bool)
}


