package model

import (
	"fmt"
)

type ResultSelector struct {
	Selector     Selector
	SubSelectors []*ResultSelector
	Result       *interface{}
}

func NewResultSelector(s Selector) *ResultSelector {
	if s.SubSelectors != nil {
		var subSelectors []*ResultSelector
		for _, ss := range s.SubSelectors {
			subSelectors = append(subSelectors, NewResultSelector(ss))
		}
		return &ResultSelector{
			s,
			subSelectors,
			nil,
		}
	}
	return &ResultSelector{
		s,
		nil,
		nil,
	}
}

func (s *ResultSelector) Iterate(cb func(*ResultSelector)) {
	cb(s)
	if s.SubSelectors != nil {
		for _, ss := range s.SubSelectors {
			ss.Iterate(cb)
		}
	}
}

func (s *ResultSelector) GetJS(runtimeVars []*RuntimeVar, parantType SelectorType, hideDefaultValueKeys *bool) (jsString string) {

	isRootSelector := !parantType.IsValid()

	if isRootSelector {
		jsString += `
		let removeDefaultValueKeys = false;
		function cleanup(result) {
			if (result) {
				if (result.forEach !== undefined) {
					result.forEach(cleanup)
				} else if (typeof result === 'object') {
					if (removeDefaultValueKeys) delete result.__value
					delete result.node
					Object.keys(result).forEach(key => {
						if (result[key]) result[key] = result[key].trim()
						cleanup(result[key])
					})
				}
			}
		}; 
		`
	}

	if isRootSelector && s.Selector.CSSSelector != nil {
		jsString += fmt.Sprintf(`const result = Array.from(document.querySelectorAll("%s"))`, *s.Selector.CSSSelector)

		switch s.Selector.Type {
		case SelectorTypeObjectArray:
			jsString += `.map(node => ({ __value: node, node: node }))`
			jsString += s.getResultNodeMutations(``)
			jsString += s.getResultRegexMutations(``)
			for _, ss := range s.SubSelectors {
				jsString += s.getSubSelectorJS(ss, s.Selector.Type, runtimeVars, hideDefaultValueKeys)
			}
		case SelectorTypeStringArray:
			jsString += s.getResultNodeMutations(`.map(node => ({ node: node }))`)
			jsString += s.getResultRegexMutations(``)
		case SelectorTypeObjectProp:
			jsString += `.reduce((prev, next) => [{ __value: prev[0].__value + next.innerHTML }], [{ __value: "" }])`
			jsString += s.getResultRegexMutations(``)
		case SelectorTypeStringProp:
			jsString += `.map(node => node.innerHTML)`
			jsString += s.getResultRegexMutations(``)
			jsString += `.join('')`
		}
	}

	if !isRootSelector {
		switch parantType {
		case SelectorTypeObjectArray:
			if s.Selector.CSSSelector != nil {
				jsString += fmt.Sprintf(`Array.from(object.node.querySelectorAll("%s")).map(node => ({ __value: node, node: node }))`, *s.Selector.CSSSelector)
				jsString += s.getResultNodeMutations(``)
				jsString += s.getResultRegexMutations(``)
			} else {
				jsString += s.getResultNodeMutations(`[object]`)
				jsString += s.getResultRegexMutations(``)
			}
		}
	}

	switch s.Selector.Type {
	case SelectorTypeObjectProp:
		fallthrough
	case SelectorTypeStringProp:
		jsString += `.pop()`
	}

	if isRootSelector {
		if hideDefaultValueKeys != nil && *hideDefaultValueKeys {
			jsString += `; removeDefaultValueKeys = true`
		}
		jsString += `; cleanup(result); result`
	}

	return ReplaceRuntimeTemplates(runtimeVars, jsString)
}

func (s *ResultSelector) getSubSelectorJS(subSelector *ResultSelector, parantType SelectorType, runtimeVars []*RuntimeVar, hideDefaultValueKeys *bool) (jsString string) {

	switch parantType {
	case SelectorTypeObjectArray:
		jsString += fmt.Sprintf(`.map(object => ({ ...object, %s: %s }))`, subSelector.Selector.Key, subSelector.GetJS(runtimeVars, parantType, hideDefaultValueKeys))
	}

	return jsString
}

func (s *ResultSelector) getResultNodeMutations(startCode string) (jsString string) {

	jsString += startCode

	switch s.Selector.Type {
	case SelectorTypeObjectArray:
		if s.Selector.HTMLAttribute != nil {
			jsString += fmt.Sprintf(`.map(object => ({ node: object.node, __value: object.node.getAttribute("%s") }))`, *s.Selector.HTMLAttribute)
		} else {
			jsString += `.map(object => ({ node: object.node, __value: object.node.innerHTML }))`
		}
	case SelectorTypeStringArray:
		if s.Selector.HTMLAttribute != nil {
			jsString += fmt.Sprintf(`.map(object => object.node.getAttribute("%s"))`, *s.Selector.HTMLAttribute)
		} else {
			jsString += `.map(object => object.node.innerHTML)`
		}
	case SelectorTypeObjectProp:
		if s.Selector.HTMLAttribute != nil {
			jsString += fmt.Sprintf(`.map(object => ({ __value: object.node.getAttribute("%s") }))`, *s.Selector.HTMLAttribute)
		} else {
			jsString += `.map(object => ({ __value: object.node.innerHTML }))`
		}
	case SelectorTypeStringProp:
		if s.Selector.HTMLAttribute != nil {
			jsString += fmt.Sprintf(`.map(object => object.node.getAttribute("%s"))`, *s.Selector.HTMLAttribute)
		} else {
			jsString += `.map(object => object.node.innerHTML)`
		}
	}

	return jsString
}

func (s *ResultSelector) getResultRegexMutations(startCode string) (jsString string) {
	jsString += startCode

	if s.Selector.Regex != nil {
		switch s.Selector.Type {
		case SelectorTypeObjectArray:
			jsString += fmt.Sprintf(`.map(object => ({ ...object, __value: object.__value.match(%s) }))`, *s.Selector.Regex)
		case SelectorTypeStringArray:
			jsString += fmt.Sprintf(`.map(str => str.match(%s))`, *s.Selector.Regex)
		case SelectorTypeObjectProp:
			jsString += fmt.Sprintf(`.map(object => ({ __value: object.__value.match(%s) }))`, *s.Selector.Regex)
		case SelectorTypeStringProp:
			jsString += fmt.Sprintf(`.map(str => str.match(%s))`, *s.Selector.Regex)
		}

		jsString += `.filter(n => n).map(matches => matches.pop())`
	}

	return jsString
}
