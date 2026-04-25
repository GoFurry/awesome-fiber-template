package writer

import "github.com/GoFurry/fiberx/internal/renderer"

type Result struct {
	DryRun       bool
	WrittenFiles int
}

type DryRunWriter struct{}

func NewDryRunWriter() DryRunWriter {
	return DryRunWriter{}
}

func (DryRunWriter) Write(rendered renderer.Result) (Result, error) {
	return Result{
		DryRun:       true,
		WrittenFiles: len(rendered.PreviewFiles),
	}, nil
}
