package pipeparser

import (
	"github.com/zclconf/go-cty/cty"
)

const rootDependencyNode = "rootDependencyNode"

type ParseModFlag uint32

const (
	CreateDefaultMod ParseModFlag = 1 << iota
	CreatePseudoResources
)

/*
	ReferenceTypeValueMap is the raw data used to build the evaluation context

When resolving hcl references like :
- query.q1
- var.v1
- mod1.query.my_query.sql

ReferenceTypeValueMap is keyed  by resource type, then by resource name
*/
type ReferenceTypeValueMap map[string]map[string]cty.Value
