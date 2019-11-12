package astutil

import (
	"github.com/samoand/gogen/src/gogentypes"
	structutil "github.com/samoand/gostructutil"
)

func ProcessTags(root gogentypes.ASTNode, tag string,
	extraMatcher func(gogentypes.ASTNode) bool,
	mutator func(gogentypes.ASTNode) error,
	maxdept int, stoponsuccess bool) error {
	matcher := func(el gogentypes.ASTNode) bool {
		v, ok := el["__tag"]
		extraMatch := extraMatcher == nil || extraMatcher(el)
		return ok && v.(string) == tag && extraMatch
	}
	_handler := func(in gogentypes.ASTNode) (interface{}, error) {
		err := mutator(in)
		return nil, err
	}
	_, err := structutil.VisitMatching(root, matcher, _handler, nil, maxdept, stoponsuccess)
	return err
}
