package semantic_analyzer

import (
	"oberon/parser"
	"os"

	"github.com/op/go-logging"
)

var LOG = logging.MustGetLogger("semantic_analyzer")
var log_backend = logging.NewLogBackend(os.Stdout, "", 0)
var log_format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{longfile}#%{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
)

var parser_log_backend_formatter = logging.NewBackendFormatter(
	log_backend,
	log_format,
)
var parserDebug = false

type AnnotatedTree struct {
	Children []*AnnotatedTree
}

func module(tree *parser.ParseNode) (*AnnotatedTree, error) {
	var moduleNode = new(AnnotatedTree)
	var childIndex = 0
	moduleName := tree.Children[childIndex]
	moduleNode.Children = append(moduleNode.Children, moduleNode)
	childIndex += 1
	if tree.Children[childIndex].Label == "importList" {
		err := importList(tree.Children[childIndex])
		if err != nil {
			return nil, err
		}
		childIndex += 1
	}
	if moduleName.Label != tree.Children[childIndex].Label {
		LOG.Error(tree.Children[len(tree.Children)-1])
		return nil, nil
	}
	childIndex += 1
	return moduleNode, nil
}

func Analyze(tree *parser.ParseNode, debug bool) (*AnnotatedTree, error) {
	logging.SetBackend(parser_log_backend_formatter)
	parserDebug = debug
	annotated_tree := new(AnnotatedTree)
	_moduleNode, err := module(tree)
	if err != nil {
		return nil, err
	}
	annotated_tree.Children = append(annotated_tree.Children, _moduleNode)

	return annotated_tree, nil
}
