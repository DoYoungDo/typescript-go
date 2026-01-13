package dcloud

import "github.com/microsoft/typescript-go/internal/compiler"

type Program struct {
	*compiler.Program
}

func NewProgram(program *compiler.Program) *Program {
	return &Program{
		Program: program,
	}
}