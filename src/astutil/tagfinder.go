package astutil

import (
	"github.com/samoand/gogen/src/gogentypes"
	structutil "github.com/samoand/gostructutil"
)

func FindTags(root gogentypes.ASTNode,
	tag string, extraMatcher func(gogentypes.ASTNode) bool,
	maxdept int, stoponsuccess bool) [](gogentypes.ASTNode) {
	matcher := func(el gogentypes.ASTNode) bool {
		v, ok := el["__tag"]
		return ok && v.(string) == tag && (extraMatcher == nil || extraMatcher(el))
	}
	return structutil.FindMatching(root, matcher, maxdept, stoponsuccess)
}

func FindFirstParent(dict gogentypes.ASTNode, tag string) gogentypes.ASTNode {
	p, ok := dict["__parent"]
	if ok && p != nil {
		pdict := *(p.(*gogentypes.ASTNode))
		if pdict["__tag"] == tag {
			return pdict
		} else {
			return FindFirstParent(pdict, tag)
		}
	} else {
		return nil
	}
}
