package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
)

type LanguageService struct {
	*ls.LanguageService
}

func (l *LanguageService) ProvideCompletion(
	ctx context.Context,
	documentURI lsproto.DocumentUri,
	LSPPosition lsproto.Position,
	context *lsproto.CompletionContext,
) (lsproto.CompletionResponse, error) {
	res, err := l.LanguageService.ProvideCompletion(ctx, documentURI, LSPPosition, context)

	if res.List != nil {
		kind := lsproto.CompletionItemKindSnippet
		res.List.Items = append(res.List.Items, &lsproto.CompletionItem{
			Label: "DCloud_Snippet",
			Kind:  &kind,
		})
	}
	return res, err
}