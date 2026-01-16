package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type LanguageService interface {
	disposable.Disposable
	GetProvideCompletion(defaultLs *ls.LanguageService)(func(ctx context.Context,documentURI lsproto.DocumentUri,LSPPosition lsproto.Position,context *lsproto.CompletionContext) (lsproto.CompletionResponse, error))
}
