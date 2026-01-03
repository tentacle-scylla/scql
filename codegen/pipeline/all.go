package pipeline

import (
	"github.com/pierre-borckmans/scql/codegen"
)

// All runs all pipeline steps in order.
func All(dirs *codegen.Dirs) error {
	if err := Download(dirs); err != nil {
		return err
	}
	if err := Extract(dirs); err != nil {
		return err
	}
	if err := Diff(dirs); err != nil {
		return err
	}
	if err := Patch(dirs); err != nil {
		return err
	}
	if err := Generate(dirs); err != nil {
		return err
	}
	if err := GenerateComplete(dirs); err != nil {
		return err
	}
	return Test(dirs)
}
