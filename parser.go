package main

import (
	"fmt"
	"os"

	"github.com/op/go-logging"

	lexer "oberon/lexer"
)

var PARSER_LOG = logging.MustGetLogger("parser")
var parser_log_backend = logging.NewLogBackend(os.Stdout, "", 0)
var parser_log_format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{longfile}#%{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
)

var parser_log_backend_formatter = logging.NewBackendFormatter(parser_log_backend, parser_log_format)
var parserDebug = false

type ParseNode struct {
	Label    string
	Children []*ParseNode
}

func print_parse_tree(root *ParseNode, indentation int) {
	if nil == root {
		return
	}
	var space = ""
	for i := 0; i < indentation; i++ {
		space += " "
	}
	fmt.Println(fmt.Sprintf("%s%s", space, root.Label))
	for _, child := range root.Children {
		print_parse_tree(child, indentation+1)
	}
}

func parse_error(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) error {
	if *position < len(*lexemes) {
		return fmt.Errorf("parse error: expected %s, found %v", message, (*lexemes)[*position])
	} else {
		return fmt.Errorf("parse error: expected %s, but reached end of stream", message)
	}
}

func debug(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) {
	if !parserDebug {
		return
	}
	if *position < len(*lexemes) {
		PARSER_LOG.Debug(fmt.Sprintf("%s (current_token: %v, position: %d)", message, ((*lexemes)[*position]), *position))
	}
}

func attempt_log(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Attempting to match %s", message), lexemes, position)
}

func attempt_optionally_log(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Attempting to optionally match %s", message), lexemes, position)
}

func did_not_match_log(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Did not match %s", message), lexemes, position)
}

func did_not_match_optionally_log(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Did not optionally match %s", message), lexemes, position)
}

func matched_log(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Matched %s", message), lexemes, position)
}

func optionally_matched_log(
	message string,
	lexemes *[]lexer.Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Optionally matched %s", message), lexemes, position)
}

func matchReservedWord(
	lexemes *[]lexer.Lexeme,
	position *int,
	terminal string,
) *ParseNode {
	if *position >= len(*lexemes) {
		return nil
	}
	lexeme := (*lexemes)[*position]
	if lexeme.Label == terminal {
		var terminalNode = new(ParseNode)
		(*terminalNode).Label = lexeme.Label
		(*position)++
		return terminalNode
	}
	return nil
}

func matchOperator(
	lexemes *[]lexer.Lexeme,
	position *int,
	operator string,
) *ParseNode {
	if *position >= len(*lexemes) {
		return nil
	}
	lexeme := (*lexemes)[*position]
	if lexeme.Typ == lexer.OP_OR_DELIM && lexeme.Label == operator {
		var terminalNode = new(ParseNode)
		(*terminalNode).Label = lexeme.Label
		(*position)++
		return terminalNode
	}
	return nil
}

func matchtype(
	lexemes *[]lexer.Lexeme,
	position *int,
	lexemetype lexer.Lexemetype,
) *ParseNode {
	if *position >= len(*lexemes) {
		return nil
	}
	lexeme := (*lexemes)[*position]
	if lexeme.Typ == lexemetype {
		var terminalNode = new(ParseNode)
		(*terminalNode).Label = lexeme.Label
		(*position)++
		return terminalNode
	}
	return nil
}

// import = ident [":=" ident].
func _import(
	lexemes *[]lexer.Lexeme,
	position *int,
) *ParseNode {
	var importNode = new(ParseNode)
	importNode.Label = "import"

	// ident
	attempt_log("ident", lexemes, position)
	_identNode := matchtype(lexemes, position, lexer.IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil
	}
	matched_log("ident", lexemes, position)

	importNode.Children = append(importNode.Children, _identNode)

	// :=
	attempt_log(":=", lexemes, position)
	_assignmentOperator := matchOperator(lexemes, position, ":=")
	if _assignmentOperator != nil {
		// ident
		matched_log(":=", lexemes, position)
		attempt_log("ident", lexemes, position)
		_identNode := matchtype(lexemes, position, lexer.IDENT)
		if _identNode == nil {
			did_not_match_log("ident", lexemes, position)
			return nil
		}
		matched_log("ident", lexemes, position)
		// importNode.Children = append(importNode.Children, _assignmentOperator)
		importNode.Children = append(importNode.Children, _identNode)
	}
	return importNode
}

// ImportList = IMPORT import {"," import} ";".
func importList(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var importListNode = new(ParseNode)
	importListNode.Label = "importList"
	var positionCheckpoint = *position

	// IMPORT
	attempt_log("IMPORT", lexemes, position)
	_importReservedWordNode := matchReservedWord(lexemes, position, "IMPORT")
	if _importReservedWordNode == nil {
		did_not_match_log("IMPORT", lexemes, position)
		return nil, nil
	}
	matched_log("IMPORT", lexemes, position)
	importListNode.Children = append(importListNode.Children, _importReservedWordNode)

	// import
	attempt_log("import", lexemes, position)
	_importNode := _import(lexemes, position)
	if _importNode == nil {
		did_not_match_log("import", lexemes, position)
		return nil, nil
	}
	matched_log("import", lexemes, position)
	importListNode.Children = append(importListNode.Children, _importNode)

	// {"," import}
	for {
		attempt_log(",", lexemes, position)
		_commaNode := matchOperator(lexemes, position, ",")
		if _commaNode == nil {
			did_not_match_log(",", lexemes, position)
			break
		}
		matched_log(",", lexemes, position)
		_additionalImportNode := _import(lexemes, position)
		if _additionalImportNode == nil {
			debug("Did not find additionalImportNode after matching ','", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		debug("Found additionalImportNode", lexemes, position)
		// importListNode.Children = append(importListNode.Children, _commaNode)
		importListNode.Children = append(importListNode.Children, _additionalImportNode)
	}

	// ;
	debug("Attempting to match ';'", lexemes, position)
	_semicolonNode := matchOperator(lexemes, position, ";")
	if _semicolonNode == nil {
		debug("Did not match ';'", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	debug("Matched ';'", lexemes, position)
	// importListNode.Children = append(importListNode.Children, _semicolonNode)

	return importListNode, nil
}

// qualident = [ident "."] ident.
func qualident(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var qualidentNode = new(ParseNode)
	qualidentNode.Label = "qualident"
	var positionCheckpoint = *position

	// [ident "."]
	attempt_log("ident", lexemes, position)
	var _identNode = matchtype(lexemes, position, lexer.IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, nil
	}
	qualidentNode.Children = append(qualidentNode.Children, _identNode)
	matched_log("ident", lexemes, position)
	positionCheckpoint = *position

	attempt_log(".", lexemes, position)
	_dotOperatorNode := matchOperator(lexemes, position, ".")
	if _dotOperatorNode != nil {
		matched_log(".", lexemes, position)

		attempt_log("ident", lexemes, position)
		_identNode := matchtype(lexemes, position, lexer.IDENT)
		if _identNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		// qualidentNode.Children = append(qualidentNode.Children, _dotOperatorNode)
		qualidentNode.Children = append(qualidentNode.Children, _identNode)
	}

	return qualidentNode, nil
}

// ExpList = expression {"," expression}.
func expList(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var expListNode = new(ParseNode)
	expListNode.Label = "expList"
	var positionCheckpoint = *position

	// expression
	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if nil == _expressionNode {
		did_not_match_log("expression", lexemes, position)
		return nil, nil
	}
	matched_log("expression", lexemes, position)
	expListNode.Children = append(expListNode.Children, _expressionNode)

	// {, expList}
	for {
		attempt_optionally_log(",", lexemes, position)
		_commaOperatorNode := matchOperator(lexemes, position, ",")
		if nil == _commaOperatorNode {
			did_not_match_optionally_log(",", lexemes, position)
			break
		}
		optionally_matched_log(",", lexemes, position)

		attempt_log("expression", lexemes, position)
		_expressionNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _expressionNode {
			did_not_match_log("expression", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("expression", lexemes, position)

		// expListNode.Children = append(expListNode.Children, _commaOperatorNode)
		expListNode.Children = append(expListNode.Children, _expressionNode)
	}
	return expListNode, nil
}

// selector = "." ident | "[" ExpList "]" | "^" | "(" qualident ")".
func selector(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var selectorNode = new(ParseNode)
	selectorNode.Label = "selector"
	var positionCheckpoint = *position

	// "." ident
	attempt_optionally_log(".", lexemes, position)
	_dotOperatorNode := matchOperator(lexemes, position, ".")
	if _dotOperatorNode != nil {
		optionally_matched_log(".", lexemes, position)

		attempt_log("ident", lexemes, position)
		_identNode := matchtype(lexemes, position, lexer.IDENT)
		if _identNode == nil {
			did_not_match_log("ident", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		did_not_match_log("ident", lexemes, position)
		// selectorNode.Children = append(selectorNode.Children, _dotOperatorNode)
		selectorNode.Children = append(selectorNode.Children, _identNode)

		return selectorNode, nil
	} else {
		did_not_match_optionally_log(".", lexemes, position)
	}

	// "[" ExpList "]"
	attempt_optionally_log("[", lexemes, position)
	_leftBracketNode := matchOperator(lexemes, position, "[")
	if _leftBracketNode != nil {
		optionally_matched_log("[", lexemes, position)

		attempt_log("expList", lexemes, position)
		_expListNode, err := expList(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _expListNode == nil {
			did_not_match_log("expList", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("expList", lexemes, position)

		attempt_log("]", lexemes, position)
		_rightBracketNode := matchOperator(lexemes, position, "]")
		if _rightBracketNode == nil {
			did_not_match_log("]", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("]", lexemes, position)

		// selectorNode.Children = append(selectorNode.Children, _leftBracketNode)
		selectorNode.Children = append(selectorNode.Children, _expListNode)
		// selectorNode.Children = append(selectorNode.Children, _rightBracketNode)

		return selectorNode, nil
	} else {
		did_not_match_optionally_log("[", lexemes, position)
	}

	// ^
	attempt_optionally_log("^", lexemes, position)
	_caratOperatorNode := matchOperator(lexemes, position, "^")
	if _caratOperatorNode != nil {
		optionally_matched_log("^", lexemes, position)
		selectorNode.Children = append(selectorNode.Children, _caratOperatorNode)
		return selectorNode, nil
	}
	did_not_match_optionally_log("^", lexemes, position)

	// "(" qualident ")"
	attempt_optionally_log("(", lexemes, position)
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode != nil {
		optionally_matched_log("(", lexemes, position)

		attempt_log("qualident", lexemes, position)
		_qualidentNode, err := qualident(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _qualidentNode == nil {
			did_not_match_log("qualident", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("qualident", lexemes, position)

		attempt_log(")", lexemes, position)
		_rightParenNode := matchOperator(lexemes, position, ")")
		if _rightParenNode == nil {
			did_not_match_log(")", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log(")", lexemes, position)

		selectorNode.Children = append(selectorNode.Children, _leftParenNode)
		selectorNode.Children = append(selectorNode.Children, _qualidentNode)
		selectorNode.Children = append(selectorNode.Children, _rightParenNode)

		return selectorNode, nil
	} else {
		did_not_match_optionally_log("(", lexemes, position)
	}

	return nil, nil
}

// designator = qualident {selector}.
func designator(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var designatorNode = new(ParseNode)
	designatorNode.Label = "designator"

	// qualident
	attempt_log("qualident", lexemes, position)
	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode == nil {
		did_not_match_log("qualident", lexemes, position)
		debug("Did not match qualident", lexemes, position)
		return nil, nil
	}
	matched_log("qualident", lexemes, position)
	designatorNode.Children = append(designatorNode.Children, _qualidentNode)

	// {selector}
	for {
		optionally_matched_log("selector", lexemes, position)
		_selectorNode, err := selector(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _selectorNode == nil {
			did_not_match_optionally_log("selector", lexemes, position)
			break
		}
		optionally_matched_log("selector", lexemes, position)
		designatorNode.Children = append(designatorNode.Children, _selectorNode)
	}
	return designatorNode, nil
}

// element = expression [".." expression].
func element(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var elementNode = new(ParseNode)
	elementNode.Label = "element"

	// expression
	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		did_not_match_log("expression", lexemes, position)
		return nil, err
	}
	matched_log("expression", lexemes, position)
	elementNode.Children = append(elementNode.Children, _expressionNode)

	// [ .. expression ]
	attempt_optionally_log("..", lexemes, position)
	_doubleDotOperator := matchOperator(lexemes, position, "..")
	if _doubleDotOperator != nil {
		attempt_optionally_log("..", lexemes, position)
		_elementNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _elementNode == nil {
			return nil, nil
		}
		optionally_matched_log("..", lexemes, position)
		elementNode.Children = append(elementNode.Children, _doubleDotOperator)
		elementNode.Children = append(elementNode.Children, _elementNode)
	}

	return elementNode, nil
}

// set = "{" [element {"," element}] "}".
func set(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var setNode = new(ParseNode)
	setNode.Label = "set"
	var positionCheckpoint = *position

	// {
	attempt_log("{", lexemes, position)
	_leftbraceNode := matchOperator(lexemes, position, "{")
	if _leftbraceNode == nil {
		did_not_match_log("{", lexemes, position)
		return nil, nil
	}
	matched_log("{", lexemes, position)
	setNode.Children = append(setNode.Children, _leftbraceNode)

	// element
	attempt_optionally_log("element", lexemes, position)
	_elementNode, err := element(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _elementNode != nil {
		optionally_matched_log("element", lexemes, position)
		setNode.Children = append(setNode.Children, _elementNode)
		// {, element}
		for {
			_commaNode := matchOperator(lexemes, position, ",")
			if _commaNode == nil {
				did_not_match_optionally_log("element", lexemes, position)
				break
			}
			optionally_matched_log("element", lexemes, position)

			attempt_log("element", lexemes, position)
			_elementNode, err := element(lexemes, position)
			if err != nil {
				return nil, err
			}
			if _elementNode == nil {
				did_not_match_log("element", lexemes, position)
				*position = positionCheckpoint
				return nil, nil
			}
			matched_log("element", lexemes, position)
			setNode.Children = append(setNode.Children, _commaNode)
			setNode.Children = append(setNode.Children, _elementNode)
		}
	} else {
		did_not_match_log("element", lexemes, position)
	}

	// }
	attempt_log("}", lexemes, position)
	_rightBraceNode := matchOperator(lexemes, position, "}")
	if _rightBraceNode == nil {
		did_not_match_log("}", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("}", lexemes, position)
	setNode.Children = append(setNode.Children, _rightBraceNode)

	return setNode, nil
}

// "(" [ExpList] ")".
func actualParameters(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var actualParametersNode = new(ParseNode)
	actualParametersNode.Label = "actualParameters"
	var positionCheckpoint = *position

	// "("
	attempt_log("(", lexemes, position)
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode == nil {
		did_not_match_log("(", lexemes, position)
		return nil, nil
	}
	actualParametersNode.Children = append(actualParametersNode.Children, _leftParenNode)
	matched_log("(", lexemes, position)

	// [expList]
	attempt_optionally_log("expList", lexemes, position)
	_expListNode, err := expList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expListNode != nil {
		did_not_match_optionally_log("expList", lexemes, position)
		actualParametersNode.Children = append(actualParametersNode.Children, _expListNode)
	}
	optionally_matched_log("expList", lexemes, position)

	// )
	attempt_log(")", lexemes, position)
	_rightParenNode := matchOperator(lexemes, position, ")")
	if _rightParenNode == nil {
		did_not_match_log(")", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log(")", lexemes, position)

	actualParametersNode.Children = append(actualParametersNode.Children, _leftParenNode)
	return actualParametersNode, nil
}

// factor = number | string | NIL | TRUE | FALSE | set | designator [ActualParameters] | "(" expression ")" | "~" factor.
func factor(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var factorNode = new(ParseNode)
	factorNode.Label = "factor"
	var positionCheckpoint = *position

	// number
	{
		// BUG: given that a term in SimpleExpression may be preceded by a
		// + or a - already, this is unnecessary, but without this,
		// -10 and +10 are unrecognized.
		attempt_optionally_log("+", lexemes, position)
		_plusOperatorNode := matchOperator(lexemes, position, "+")
		if _plusOperatorNode == nil {
			did_not_match_optionally_log("+", lexemes, position)
			attempt_optionally_log("-", lexemes, position)
			_minusOperatorNode := matchOperator(lexemes, position, "-")
			if _minusOperatorNode != nil {
				optionally_matched_log("-", lexemes, position)
				factorNode.Children = append(factorNode.Children, _minusOperatorNode)
			} else {
				did_not_match_optionally_log("-", lexemes, position)
			}
		} else {
			optionally_matched_log("+", lexemes, position)
			factorNode.Children = append(factorNode.Children, _plusOperatorNode)
		}

		attempt_log("integer", lexemes, position)
		_integerNode := matchtype(lexemes, position, lexer.INTEGER)
		if _integerNode != nil {
			matched_log("integer", lexemes, position)
			factorNode.Children = append(factorNode.Children, _integerNode)
			debug("Matched integer", lexemes, position)
			return factorNode, nil
		}
		did_not_match_log("integer", lexemes, position)

		attempt_log("real", lexemes, position)
		_realNode := matchtype(lexemes, position, lexer.REAL)
		if _realNode != nil {
			matched_log("real", lexemes, position)
			factorNode.Children = append(factorNode.Children, _realNode)
			return factorNode, nil
		}
		did_not_match_log("real", lexemes, position)

		*position = positionCheckpoint
		// TODO: determine whether it is sage to return nil, nil here
	}

	// string
	attempt_log("string", lexemes, position)
	_stringNode := matchtype(lexemes, position, lexer.STRING)
	if _stringNode != nil {
		matched_log("string", lexemes, position)
		factorNode.Children = append(factorNode.Children, _stringNode)
		return factorNode, nil
	}
	did_not_match_log("string", lexemes, position)

	// NIL
	attempt_log("NIL", lexemes, position)
	_nilNode := matchReservedWord(lexemes, position, "NIL")
	if _nilNode != nil {
		matched_log("NIL", lexemes, position)
		factorNode.Children = append(factorNode.Children, _nilNode)
		return factorNode, nil
	}
	did_not_match_log("NIL", lexemes, position)

	// TRUE
	attempt_log("TRUE", lexemes, position)
	_trueNode := matchReservedWord(lexemes, position, "TRUE")
	if _trueNode != nil {
		matched_log("TRUE", lexemes, position)
		factorNode.Children = append(factorNode.Children, _trueNode)
		return factorNode, nil
	}
	did_not_match_log("TRUE", lexemes, position)

	// FALSE
	attempt_log("FALSE", lexemes, position)
	_falseNode := matchReservedWord(lexemes, position, "FALSE")
	if _falseNode != nil {
		matched_log("FALSE", lexemes, position)
		debug("Matched FALSE", lexemes, position)
		factorNode.Children = append(factorNode.Children, _falseNode)
		return factorNode, nil
	}
	did_not_match_log("FALSE", lexemes, position)

	// set
	attempt_log("set", lexemes, position)
	_setNode, err := set(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _setNode != nil {
		matched_log("set", lexemes, position)
		factorNode.Children = append(factorNode.Children, _setNode)
		return factorNode, nil
	}
	did_not_match_log("set", lexemes, position)

	// designator [ActualParameters]
	attempt_log("designator", lexemes, position)
	_designatorNode, err := designator(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _designatorNode != nil {
		matched_log("designator", lexemes, position)
		factorNode.Children = append(factorNode.Children, _designatorNode)

		optionally_matched_log("actualParameters", lexemes, position)
		_actualParametersNode, err := actualParameters(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil != _actualParametersNode {
			optionally_matched_log("actualParameters", lexemes, position)
			factorNode.Children = append(factorNode.Children, _actualParametersNode)
		} else {
			did_not_match_optionally_log("actualParameters", lexemes, position)
		}
		return factorNode, nil
	}
	did_not_match_log("designator", lexemes, position)

	// "(" expression ")"
	attempt_log("(", lexemes, position)
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode != nil {
		matched_log("(", lexemes, position)

		attempt_log("expression", lexemes, position)
		_expressionNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _expressionNode {
			did_not_match_log("expression", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("expression", lexemes, position)

		attempt_log(")", lexemes, position)
		_rightParenNode := matchOperator(lexemes, position, ")")
		if err != nil {
			return nil, err
		}
		if _rightParenNode == nil {
			did_not_match_log(")", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log(")", lexemes, position)

		factorNode.Children = append(factorNode.Children, _leftParenNode)
		factorNode.Children = append(factorNode.Children, _expressionNode)
		factorNode.Children = append(factorNode.Children, _rightParenNode)

		return factorNode, nil
	}

	// "~" factor
	attempt_log("~", lexemes, position)
	_tildeOperator := matchOperator(lexemes, position, "~")
	if _tildeOperator == nil {
		did_not_match_log("~", lexemes, position)
		return nil, nil
	}
	matched_log("~", lexemes, position)

	attempt_log("factor", lexemes, position)
	_factorNode, err := factor(lexemes, position)
	if err != nil {
		return nil, err
	}
	if nil == _factorNode {
		did_not_match_log("factor", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("factor", lexemes, position)

	factorNode.Children = append(factorNode.Children, _tildeOperator)
	factorNode.Children = append(factorNode.Children, _factorNode)

	return factorNode, nil
}

func mulOperator(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var mulOperatorNode = new(ParseNode)
	mulOperatorNode.Label = "mulOperator"

	attempt_log("*", lexemes, position)
	_asterixOperatorNode := matchOperator(lexemes, position, "*")
	if _asterixOperatorNode != nil {
		matched_log("*", lexemes, position)
		mulOperatorNode.Children = append(mulOperatorNode.Children, _asterixOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("*", lexemes, position)

	attempt_log("/", lexemes, position)
	_divOperatorNode := matchOperator(lexemes, position, "/")
	if _divOperatorNode != nil {
		matched_log("/", lexemes, position)
		mulOperatorNode.Children = append(mulOperatorNode.Children, _divOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("/", lexemes, position)

	attempt_log("DIV", lexemes, position)
	_divNode := matchReservedWord(lexemes, position, "DIV")
	if _divNode != nil {
		matched_log("DIV", lexemes, position)
		mulOperatorNode.Children = append(mulOperatorNode.Children, _divNode)
		return mulOperatorNode, nil
	}
	matched_log("DIV", lexemes, position)

	attempt_log("MOD", lexemes, position)
	_modeOperatorNode := matchReservedWord(lexemes, position, "MOD")
	if _modeOperatorNode != nil {
		matched_log("MOD", lexemes, position)
		mulOperatorNode.Children = append(mulOperatorNode.Children, _modeOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("MOD", lexemes, position)

	attempt_log("&", lexemes, position)
	_ampersandOperatorNode := matchOperator(lexemes, position, "&")
	if _ampersandOperatorNode != nil {
		matched_log("&", lexemes, position)
		mulOperatorNode.Children = append(mulOperatorNode.Children, _ampersandOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("&", lexemes, position)

	return nil, nil
}

func term(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var termNode = new(ParseNode)
	termNode.Label = "term"
	var positionCheckpoint = *position

	attempt_log("factor", lexemes, position)
	_factorNode, err := factor(lexemes, position)
	if err != nil {
		debug("Error matching factor", lexemes, position)
		return nil, err
	}
	if _factorNode == nil {
		did_not_match_log("factor", lexemes, position)
		return nil, nil
	}
	debug("Matched factor", lexemes, position)
	termNode.Children = append(termNode.Children, _factorNode)
	for {
		attempt_optionally_log("mulOperator", lexemes, position)
		_mulOperatorNode, err := mulOperator(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _mulOperatorNode {
			did_not_match_optionally_log("mulOperator", lexemes, position)
			break
		}

		attempt_log("factor", lexemes, position)
		_factorNode, err := factor(lexemes, position)
		if err != nil {
			*position = positionCheckpoint
			return nil, err
		}
		if nil == _factorNode {
			did_not_match_log("factor", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("factor", lexemes, position)

		termNode.Children = append(termNode.Children, _mulOperatorNode)
		termNode.Children = append(termNode.Children, _factorNode)
		positionCheckpoint = *position
	}
	return termNode, nil
}

func addOperator(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var addOperatorNode = new(ParseNode)
	addOperatorNode.Label = "addOperator"

	attempt_optionally_log("+", lexemes, position)
	_plusOperatorNode := matchOperator(lexemes, position, "+")
	if _plusOperatorNode != nil {
		matched_log("+", lexemes, position)
		addOperatorNode.Children = append(addOperatorNode.Children, _plusOperatorNode)
		return addOperatorNode, nil
	} else {
		did_not_match_optionally_log("+", lexemes, position)
	}

	attempt_optionally_log("-", lexemes, position)
	_minusOperatorNode := matchOperator(lexemes, position, "-")
	if _minusOperatorNode != nil {
		matched_log("-", lexemes, position)
		addOperatorNode.Children = append(addOperatorNode.Children, _minusOperatorNode)
		return addOperatorNode, nil
	} else {
		did_not_match_optionally_log("-", lexemes, position)
	}

	attempt_optionally_log("OR", lexemes, position)
	_orOperatorNode := matchReservedWord(lexemes, position, "OR")
	if _orOperatorNode != nil {
		matched_log("OR", lexemes, position)
		addOperatorNode.Children = append(addOperatorNode.Children, _orOperatorNode)
		return addOperatorNode, nil
	} else {
		did_not_match_optionally_log("OR", lexemes, position)
	}

	return nil, nil
}

func simpleExpression(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var simpleExpressionNode = new(ParseNode)
	simpleExpressionNode.Label = "simpleExpression"
	var positionCheckpoint = *position

	attempt_optionally_log("+", lexemes, position)
	_plusOperatorNode := matchOperator(lexemes, position, "+")
	if nil == _plusOperatorNode {
		did_not_match_optionally_log("+", lexemes, position)

		attempt_optionally_log("-", lexemes, position)
		_minusOperatorNode := matchOperator(lexemes, position, "-")
		if nil != _minusOperatorNode {
			optionally_matched_log("-", lexemes, position)
			simpleExpressionNode.Children = append(simpleExpressionNode.Children, _minusOperatorNode)
		} else {
			did_not_match_optionally_log("-", lexemes, position)
		}
	} else {
		optionally_matched_log("+", lexemes, position)
		simpleExpressionNode.Children = append(simpleExpressionNode.Children, _plusOperatorNode)
	}

	attempt_log("term", lexemes, position)
	_termNode, err := term(lexemes, position)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if nil == _termNode {
		did_not_match_log("term", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	simpleExpressionNode.Children = append(simpleExpressionNode.Children, _termNode)
	matched_log("term", lexemes, position)

	for {
		attempt_optionally_log("addOperator", lexemes, position)
		_addOperatorNode, err := addOperator(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _addOperatorNode {
			did_not_match_optionally_log("addOperator", lexemes, position)
			break
		}
		optionally_matched_log("addOperator", lexemes, position)

		attempt_log("term", lexemes, position)
		_termNode, err := term(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _termNode {
			did_not_match_log("term", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("term", lexemes, position)
		simpleExpressionNode.Children = append(simpleExpressionNode.Children, _addOperatorNode)
		simpleExpressionNode.Children = append(simpleExpressionNode.Children, _termNode)
	}
	return simpleExpressionNode, nil
}

func relation(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var relationOperatorNode = new(ParseNode)
	relationOperatorNode.Label = "relation"

	attempt_log("=", lexemes, position)
	_equalOperatorNode := matchOperator(lexemes, position, "=")
	if _equalOperatorNode != nil {
		matched_log("=", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _equalOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("=", lexemes, position)

	attempt_log("#", lexemes, position)
	_hashOperatorNode := matchOperator(lexemes, position, "#")
	if _hashOperatorNode != nil {
		matched_log("#", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _hashOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("#", lexemes, position)

	attempt_log("<", lexemes, position)
	_lessThanOperatorNode := matchOperator(lexemes, position, "<")
	if _lessThanOperatorNode != nil {
		matched_log("<", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _lessThanOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("<", lexemes, position)

	attempt_log("<=", lexemes, position)
	_lessThanEqualOperatorNode := matchOperator(lexemes, position, "<=")
	if _lessThanEqualOperatorNode != nil {
		matched_log("<=", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _lessThanEqualOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("<=", lexemes, position)

	attempt_log(">", lexemes, position)
	_greaterThanOperatorNode := matchOperator(lexemes, position, ">")
	if _greaterThanOperatorNode != nil {
		matched_log(">", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _greaterThanOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log(">", lexemes, position)

	attempt_log(">=", lexemes, position)
	_greaterThanEqualOperatorNode := matchOperator(lexemes, position, ">=")
	if _greaterThanEqualOperatorNode != nil {
		matched_log(">=", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _greaterThanEqualOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log(">=", lexemes, position)

	attempt_log("IN", lexemes, position)
	_inOperatorNode := matchReservedWord(lexemes, position, "IN")
	if _inOperatorNode != nil {
		matched_log("IN", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _inOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("IN", lexemes, position)

	attempt_log("IS", lexemes, position)
	_isOperatorNode := matchReservedWord(lexemes, position, "IS")
	if _isOperatorNode != nil {
		matched_log("IS", lexemes, position)
		relationOperatorNode.Children = append(relationOperatorNode.Children, _isOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("IS", lexemes, position)

	return nil, nil
}

func expression(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var expressionNode = new(ParseNode)
	expressionNode.Label = "expression"
	var positionCheckpoint = *position

	attempt_log("simpleExpression", lexemes, position)
	_simpleExpressionNode, err := simpleExpression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _simpleExpressionNode == nil {
		did_not_match_log("simpleExpression", lexemes, position)
		return nil, nil
	}
	expressionNode.Children = append(expressionNode.Children, _simpleExpressionNode)
	matched_log("simpleExpression", lexemes, position)

	attempt_optionally_log("relation", lexemes, position)
	_relationNode, err := relation(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _relationNode != nil {
		optionally_matched_log("relation", lexemes, position)

		attempt_log("simpleExpression", lexemes, position)
		_simpleExpressionNode, err := simpleExpression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _simpleExpressionNode == nil {
			did_not_match_log("simpleExpression", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("simpleExpression", lexemes, position)

		expressionNode.Children = append(expressionNode.Children, _relationNode)
		expressionNode.Children = append(expressionNode.Children, _simpleExpressionNode)
	}
	return expressionNode, nil
}

func length(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	return constExpression(lexemes, position)
}

func _type(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var _typeNode = new(ParseNode)
	_typeNode.Label = "type"

	attempt_log("qualident", lexemes, position)
	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode != nil {
		matched_log("qualident", lexemes, position)
		_typeNode.Children = append(_typeNode.Children, _qualidentNode)
		return _typeNode, nil
	}
	did_not_match_log("qualident", lexemes, position)

	attempt_log("structype", lexemes, position)
	_structypeNode, err := structype(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _structypeNode != nil {
		matched_log("structype", lexemes, position)
		_typeNode.Children = append(_typeNode.Children, _structypeNode)
		return _typeNode, nil
	}
	did_not_match_log("structype", lexemes, position)

	return nil, nil
}

func arraytype(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var arraytypeNode = new(ParseNode)
	arraytypeNode.Label = "arraytype"
	var positionCheckpoint = *position
	var positionCheckpoint1 = *position

	attempt_log("ARRAY", lexemes, position)
	_arrayReservedWord := matchReservedWord(lexemes, position, "ARRAY")
	if _arrayReservedWord == nil {
		did_not_match_log("ARRAY", lexemes, position)
		return nil, nil
	}
	matched_log("ARRAY", lexemes, position)

	attempt_log("length", lexemes, position)
	_lengthNode, err := length(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _lengthNode == nil {
		did_not_match_log("length", lexemes, position)
		return nil, nil
	}
	arraytypeNode.Children = append(arraytypeNode.Children, _arrayReservedWord)
	arraytypeNode.Children = append(arraytypeNode.Children, _lengthNode)
	matched_log("length", lexemes, position)

	for {
		attempt_optionally_log(",", lexemes, position)
		_commaOperatorNode := matchOperator(lexemes, position, ",")
		if err != nil {
			return nil, err
		}
		if _commaOperatorNode == nil {
			did_not_match_optionally_log(",", lexemes, position)
			break
		}
		optionally_matched_log(",", lexemes, position)

		attempt_log("length", lexemes, position)
		_lengthNode, err := length(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _lengthNode == nil {
			did_not_match_log("length", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		arraytypeNode.Children = append(arraytypeNode.Children, _commaOperatorNode)
		arraytypeNode.Children = append(arraytypeNode.Children, _lengthNode)
		matched_log("length", lexemes, position)
	}

	attempt_log("OF", lexemes, position)
	_ofReservedWordNode := matchReservedWord(lexemes, position, "OF")
	if _ofReservedWordNode == nil {
		did_not_match_log("OF", lexemes, position)
		*position = positionCheckpoint1
		return nil, nil
	}
	matched_log("OF", lexemes, position)

	matched_log("type", lexemes, position)
	_typeNode, err := _type(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _typeNode == nil {
		did_not_match_log("type", lexemes, position)
		*position = positionCheckpoint1
		return nil, nil
	}
	matched_log("type", lexemes, position)

	arraytypeNode.Children = append(arraytypeNode.Children, _ofReservedWordNode)
	arraytypeNode.Children = append(arraytypeNode.Children, _typeNode)

	return arraytypeNode, nil
}

func basetype(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	return qualident(lexemes, position)
}

func identList(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var identListNode = new(ParseNode)
	identListNode.Label = "identList"
	var positionCheckpoint = *position

	attempt_log("identdef", lexemes, position)
	_identDefNode, err := identdef(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identDefNode == nil {
		did_not_match_log("identdef", lexemes, position)
		return nil, nil
	}
	identListNode.Children = append(identListNode.Children, _identDefNode)
	matched_log("identdef", lexemes, position)

	for {
		attempt_optionally_log(",", lexemes, position)
		_commaOperatorNode := matchOperator(lexemes, position, ",")
		if _commaOperatorNode == nil {
			did_not_match_optionally_log(",", lexemes, position)
			break
		}
		optionally_matched_log(",", lexemes, position)

		attempt_log("identdef", lexemes, position)
		_identDefNode, err := identdef(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _identDefNode == nil {
			did_not_match_log("identdef", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}

		identListNode.Children = append(identListNode.Children, _commaOperatorNode)
		identListNode.Children = append(identListNode.Children, _identDefNode)
	}
	return identListNode, nil

}

func fieldList(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var fieldListNode = new(ParseNode)
	fieldListNode.Label = "fieldList"
	var positionCheckpoint = *position

	attempt_log("identList", lexemes, position)
	_identListNode, err := identList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identListNode == nil {
		did_not_match_log("identList", lexemes, position)
		return nil, nil
	}
	matched_log("identList", lexemes, position)

	attempt_log(":", lexemes, position)
	_colonNode := matchOperator(lexemes, position, ":")
	if _colonNode == nil {
		did_not_match_log(":", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log(":", lexemes, position)

	attempt_log("type", lexemes, position)
	_typeNode, err := _type(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _typeNode == nil {
		did_not_match_log("type", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("type", lexemes, position)

	fieldListNode.Children = append(fieldListNode.Children, _identListNode)
	fieldListNode.Children = append(fieldListNode.Children, _colonNode)
	fieldListNode.Children = append(fieldListNode.Children, _typeNode)

	return fieldListNode, nil
}

func fieldListSequence(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var fieldListSequenceNode = new(ParseNode)
	fieldListSequenceNode.Label = "fieldListSequence"

	attempt_log("fieldListSequence", lexemes, position)
	_fieldListNode, err := fieldList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _fieldListNode == nil {
		attempt_log("fieldListSequence", lexemes, position)
		return nil, nil
	}
	fieldListSequenceNode.Children = append(fieldListSequenceNode.Children, _fieldListNode)
	attempt_log("fieldListSequence", lexemes, position)

	for {
		attempt_optionally_log(";", lexemes, position)
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			did_not_match_optionally_log(";", lexemes, position)
			break
		}
		attempt_optionally_log(";", lexemes, position)

		attempt_log("fieldList", lexemes, position)
		_fieldListNode, err := fieldList(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _fieldListNode == nil {
			did_not_match_log("fieldList", lexemes, position)
			break
		}
		fieldListSequenceNode.Children = append(fieldListSequenceNode.Children, _semicolonNode)
		fieldListSequenceNode.Children = append(fieldListSequenceNode.Children, _fieldListNode)
		matched_log("fieldList", lexemes, position)
	}
	return fieldListSequenceNode, nil
}

func recordtype(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var recordtypeNode = new(ParseNode)
	recordtypeNode.Label = "recordtype"
	var positionCheckpoint = *position

	attempt_log("RECORD", lexemes, position)
	_recordReservedWordNode := matchReservedWord(lexemes, position, "RECORD")
	if _recordReservedWordNode == nil {
		did_not_match_log("RECORD", lexemes, position)
		return nil, nil
	}
	matched_log("RECORD", lexemes, position)

	attempt_optionally_log("(", lexemes, position)
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode != nil {
		matched_log("(", lexemes, position)
		attempt_log("basetype", lexemes, position)
		_basetypeNode, err := basetype(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _basetypeNode == nil {
			did_not_match_log("basetype", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("basetype", lexemes, position)

		attempt_log(")", lexemes, position)
		_rightParenNode := matchOperator(lexemes, position, ")")
		if _rightParenNode == nil {
			did_not_match_log(")", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log(")", lexemes, position)

		recordtypeNode.Children = append(recordtypeNode.Children, _leftParenNode)
		recordtypeNode.Children = append(recordtypeNode.Children, _basetypeNode)
		recordtypeNode.Children = append(recordtypeNode.Children, _rightParenNode)
	} else {
		did_not_match_log("(", lexemes, position)
	}

	attempt_optionally_log("fieldListSequence", lexemes, position)
	_fieldListSequenceNode, err := fieldListSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _fieldListSequenceNode != nil {
		optionally_matched_log("fieldListSequence", lexemes, position)
		recordtypeNode.Children = append(recordtypeNode.Children, _fieldListSequenceNode)
	} else {
		did_not_match_optionally_log("fieldListSequence", lexemes, position)
	}

	attempt_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		did_not_match_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("END", lexemes, position)
	recordtypeNode.Children = append(recordtypeNode.Children, _endReservedWordNode)

	return recordtypeNode, nil
}

func pointertype(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var pointertypeNode = new(ParseNode)
	pointertypeNode.Label = "pointertype"
	var positionCheckpoint = *position

	attempt_log("POINTER", lexemes, position)
	_pointertypeNode := matchReservedWord(lexemes, position, "POINTER")
	if _pointertypeNode == nil {
		did_not_match_log("POINTER", lexemes, position)
		return nil, nil
	}
	matched_log("POINTER", lexemes, position)

	attempt_log("TO", lexemes, position)
	_toNode := matchReservedWord(lexemes, position, "TO")
	if _toNode == nil {
		did_not_match_log("TO", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("TO", lexemes, position)

	attempt_log("type", lexemes, position)
	_typeNode, err := _type(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _typeNode == nil {
		did_not_match_log("type", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("type", lexemes, position)

	pointertypeNode.Children = append(pointertypeNode.Children, _pointertypeNode)
	pointertypeNode.Children = append(pointertypeNode.Children, _toNode)
	pointertypeNode.Children = append(pointertypeNode.Children, _typeNode)

	return pointertypeNode, nil
}

func formaltype(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var formaltypeNode = new(ParseNode)
	formaltypeNode.Label = "formaltype"
	var positionCheckpoint = *position

	attempt_log("ARRAY", lexemes, position)
	_arrayReservedNode := matchReservedWord(lexemes, position, "ARRAY")
	if _arrayReservedNode != nil {
		matched_log("ARRAY", lexemes, position)

		attempt_log("OF", lexemes, position)
		_ofReservedNode := matchReservedWord(lexemes, position, "OF")
		if _ofReservedNode == nil {
			did_not_match_log("OF", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("OF", lexemes, position)
		formaltypeNode.Children = append(formaltypeNode.Children, _arrayReservedNode)
		formaltypeNode.Children = append(formaltypeNode.Children, _ofReservedNode)
	} else {
		did_not_match_log("ARRAY", lexemes, position)
	}

	attempt_log("qualident", lexemes, position)
	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode == nil {
		did_not_match_log("qualident", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	formaltypeNode.Children = append(formaltypeNode.Children, _qualidentNode)
	matched_log("qualident", lexemes, position)

	return formaltypeNode, nil
}

func fpSection(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var fpSectionNode = new(ParseNode)
	fpSectionNode.Label = "fpSection"
	var positionCheckpoint = *position

	attempt_log("VAR", lexemes, position)
	_varNode := matchReservedWord(lexemes, position, "VAR")
	if _varNode != nil {
		matched_log("VAR", lexemes, position)
		fpSectionNode.Children = append(fpSectionNode.Children, _varNode)
	} else {
		did_not_match_log("VAR", lexemes, position)
	}

	attempt_log("ident", lexemes, position)
	_identNode := matchtype(lexemes, position, lexer.IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, nil
	}
	matched_log("ident", lexemes, position)
	fpSectionNode.Children = append(fpSectionNode.Children, _identNode)

	for {
		attempt_optionally_log(",", lexemes, position)
		_commaNode := matchOperator(lexemes, position, ",")
		if _commaNode == nil {
			did_not_match_optionally_log(",", lexemes, position)
			break
		}
		optionally_matched_log(",", lexemes, position)

		attempt_log("ident", lexemes, position)
		_identNode := matchtype(lexemes, position, lexer.IDENT)
		if _identNode == nil {
			did_not_match_log("ident", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("ident", lexemes, position)

		fpSectionNode.Children = append(fpSectionNode.Children, _commaNode)
		fpSectionNode.Children = append(fpSectionNode.Children, _identNode)
	}

	attempt_log(":", lexemes, position)
	_colonNode := matchOperator(lexemes, position, ":")
	if _colonNode == nil {
		did_not_match_log(":", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	fpSectionNode.Children = append(fpSectionNode.Children, _colonNode)
	matched_log(":", lexemes, position)

	attempt_log("formaltype", lexemes, position)
	_formaltypeNode, err := formaltype(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _formaltypeNode == nil {
		did_not_match_log("formaltype", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("formaltype", lexemes, position)
	fpSectionNode.Children = append(fpSectionNode.Children, _formaltypeNode)

	return fpSectionNode, nil
}

func formalParameters(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var formalParametersNode = new(ParseNode)
	formalParametersNode.Label = "formalParameters"
	var positionCheckpoint = *position

	attempt_log("(", lexemes, position)
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode == nil {
		did_not_match_log("(", lexemes, position)
		return nil, nil
	}
	matched_log("(", lexemes, position)
	formalParametersNode.Children = append(formalParametersNode.Children, _leftParenNode)

	attempt_optionally_log("fpSectionNode", lexemes, position)
	_fpSectionNode, err := fpSection(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _fpSectionNode != nil {
		optionally_matched_log("fpSectionNode", lexemes, position)
		formalParametersNode.Children = append(formalParametersNode.Children, _fpSectionNode)
		for {
			attempt_optionally_log(";", lexemes, position)
			_semicolonNode := matchOperator(lexemes, position, ";")
			if _semicolonNode == nil {
				did_not_match_optionally_log(";", lexemes, position)
				break
			}
			optionally_matched_log(";", lexemes, position)

			attempt_log("fpSection", lexemes, position)
			_fpSectionNode, err := fpSection(lexemes, position)
			if err != nil {
				return nil, err
			}
			if _fpSectionNode == nil {
				did_not_match_log("fpSection", lexemes, position)
				*position = positionCheckpoint
				return nil, nil
			}
			matched_log("fpSection", lexemes, position)

			formalParametersNode.Children = append(formalParametersNode.Children, _semicolonNode)
			formalParametersNode.Children = append(formalParametersNode.Children, _fpSectionNode)
		}
	}

	attempt_log(")", lexemes, position)
	_rightParenNode := matchOperator(lexemes, position, ")")
	if _rightParenNode == nil {
		did_not_match_log(")", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log(")", lexemes, position)
	formalParametersNode.Children = append(formalParametersNode.Children, _rightParenNode)

	attempt_log(":", lexemes, position)
	_colonNode := matchOperator(lexemes, position, ":")
	if _colonNode != nil {
		matched_log(":", lexemes, position)
		attempt_log("qualident", lexemes, position)
		_qualidentNode, err := qualident(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _qualidentNode == nil {
			did_not_match_log("qualident", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("qualident", lexemes, position)
		formalParametersNode.Children = append(formalParametersNode.Children, _colonNode)
		formalParametersNode.Children = append(formalParametersNode.Children, _qualidentNode)
	} else {
		did_not_match_log(":", lexemes, position)
	}
	return formalParametersNode, nil
}

func proceduretype(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var proceduretypeNode = new(ParseNode)
	proceduretypeNode.Label = "proceduretype"

	attempt_log("PROCEDURE", lexemes, position)
	_procedureNode := matchReservedWord(lexemes, position, "PROCEDURE")
	if _procedureNode == nil {
		did_not_match_log("PROCEDURE", lexemes, position)
		return nil, nil
	}
	proceduretypeNode.Children = append(proceduretypeNode.Children, _procedureNode)
	matched_log("PROCEDURE", lexemes, position)

	attempt_optionally_log("formalParameters", lexemes, position)
	_formalParametersNode, err := formalParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _formalParametersNode != nil {
		matched_log("formalParameters", lexemes, position)
		proceduretypeNode.Children = append(proceduretypeNode.Children, _formalParametersNode)
	} else {
		did_not_match_optionally_log("formalParameters", lexemes, position)
	}
	return proceduretypeNode, nil
}

func structype(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var structypeNode = new(ParseNode)
	structypeNode.Label = "structype"

	attempt_log("arraytype", lexemes, position)
	_arraytypeNode, err := arraytype(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _arraytypeNode != nil {
		matched_log("arraytype", lexemes, position)
		structypeNode.Children = append(structypeNode.Children, _arraytypeNode)
		return structypeNode, nil
	}
	did_not_match_log("arraytype", lexemes, position)

	attempt_log("recordtype", lexemes, position)
	_recordtypeNode, err := recordtype(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _recordtypeNode != nil {
		matched_log("recordtype", lexemes, position)
		structypeNode.Children = append(structypeNode.Children, _recordtypeNode)
		return structypeNode, nil
	}
	did_not_match_log("recordtype", lexemes, position)

	attempt_log("pointertype", lexemes, position)
	_pointertypeNode, err := pointertype(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _pointertypeNode != nil {
		matched_log("pointertype", lexemes, position)
		structypeNode.Children = append(structypeNode.Children, _pointertypeNode)
		return structypeNode, nil
	}
	did_not_match_log("pointertype", lexemes, position)

	attempt_log("proceduretype", lexemes, position)
	_proceduretypeNode, err := proceduretype(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _proceduretypeNode != nil {
		matched_log("proceduretype", lexemes, position)
		structypeNode.Children = append(structypeNode.Children, _proceduretypeNode)
		return structypeNode, nil
	}
	did_not_match_log("proceduretype", lexemes, position)

	return nil, nil
}

func typeDeclaration(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var typeDeclarationNode = new(ParseNode)
	typeDeclarationNode.Label = "typeDeclaration"
	var positionCheckpoint = *position

	attempt_log("identdef", lexemes, position)
	_identDefNode, err := identdef(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identDefNode == nil {
		did_not_match_log("identdef", lexemes, position)
		return nil, nil
	}
	matched_log("identdef", lexemes, position)

	attempt_log("=", lexemes, position)
	_equalOperatorNode := matchOperator(lexemes, position, "=")
	if _equalOperatorNode == nil {
		did_not_match_log("=", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("=", lexemes, position)

	attempt_log("structype", lexemes, position)
	_structypeNode, err := structype(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _structypeNode == nil {
		did_not_match_log("structype", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("structype", lexemes, position)

	typeDeclarationNode.Children = append(typeDeclarationNode.Children, _identDefNode)
	typeDeclarationNode.Children = append(typeDeclarationNode.Children, _equalOperatorNode)
	typeDeclarationNode.Children = append(typeDeclarationNode.Children, _structypeNode)

	return typeDeclarationNode, nil
}

func identdef(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var identdefNode = new(ParseNode)
	identdefNode.Label = "identdef"

	attempt_log("ident", lexemes, position)
	_identNode := matchtype(lexemes, position, lexer.IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, nil
	}
	identdefNode.Children = append(identdefNode.Children, _identNode)
	matched_log("ident", lexemes, position)

	attempt_log("*", lexemes, position)
	_asterixNode := matchOperator(lexemes, position, "*")
	if _asterixNode != nil {
		did_not_match_log("*", lexemes, position)
		identdefNode.Children = append(identdefNode.Children, _asterixNode)
	}
	matched_log("*", lexemes, position)

	return identdefNode, nil
}

func constExpression(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	return expression(lexemes, position)
}

func assignment(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var assignmentNode = new(ParseNode)
	assignmentNode.Label = "assignment"
	var positionCheckpoint = *position

	attempt_log("designator", lexemes, position)
	_designatorNode, err := designator(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _designatorNode == nil {
		did_not_match_log("designator", lexemes, position)
		return nil, nil
	}
	matched_log("designator", lexemes, position)

	attempt_log(":=", lexemes, position)
	_colonEqualOperatorNode := matchOperator(lexemes, position, ":=")
	if _colonEqualOperatorNode == nil {
		did_not_match_log(":=", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log(":=", lexemes, position)

	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		did_not_match_log("expression", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("expression", lexemes, position)

	assignmentNode.Children = append(assignmentNode.Children, _designatorNode)
	assignmentNode.Children = append(assignmentNode.Children, _colonEqualOperatorNode)
	assignmentNode.Children = append(assignmentNode.Children, _expressionNode)

	return assignmentNode, nil
}

func procedureCall(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureCallNode = new(ParseNode)
	procedureCallNode.Label = "procedureCall"

	attempt_log("designator", lexemes, position)
	_designatorNode, err := designator(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _designatorNode == nil {
		did_not_match_log("designator", lexemes, position)
		return nil, nil
	}
	procedureCallNode.Children = append(procedureCallNode.Children, _designatorNode)
	matched_log("designator", lexemes, position)

	attempt_log("actualParameters", lexemes, position)
	_actualParametersNode, err := actualParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _actualParametersNode != nil {
		did_not_match_log("actualParameters", lexemes, position)
		procedureCallNode.Children = append(procedureCallNode.Children, _actualParametersNode)
	}
	matched_log("actualParameters", lexemes, position)

	return procedureCallNode, nil
}

func ifStatement(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var ifStatementNode = new(ParseNode)
	ifStatementNode.Label = "ifStatement"
	var positionCheckpoint = *position

	attempt_log("IF", lexemes, position)
	_ifReservedWordNode := matchReservedWord(lexemes, position, "IF")
	if _ifReservedWordNode == nil {
		did_not_match_log("IF", lexemes, position)
		return nil, nil
	}
	matched_log("IF", lexemes, position)

	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		attempt_log("expression", lexemes, position)
		return nil, err
	}
	if _expressionNode == nil {
		did_not_match_log("expression", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("expression", lexemes, position)

	attempt_log("THEN", lexemes, position)
	_thenReservedWordNode := matchReservedWord(lexemes, position, "THEN")
	if _thenReservedWordNode == nil {
		did_not_match_log("THEN", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("THEN", lexemes, position)

	attempt_log("statementSequence", lexemes, position)
	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		did_not_match_log("statementSequence", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	ifStatementNode.Children = append(ifStatementNode.Children, _ifReservedWordNode)
	ifStatementNode.Children = append(ifStatementNode.Children, _expressionNode)
	ifStatementNode.Children = append(ifStatementNode.Children, _thenReservedWordNode)
	ifStatementNode.Children = append(ifStatementNode.Children, _statementSequenceNode)
	matched_log("statementSequence", lexemes, position)

	for {
		attempt_optionally_log("ELSIF", lexemes, position)
		_elsifReservedWordNode := matchReservedWord(lexemes, position, "ELSIF")
		if _elsifReservedWordNode == nil {
			did_not_match_optionally_log("ELSIF", lexemes, position)
			break
		}
		optionally_matched_log("ELSIF", lexemes, position)

		attempt_log("expression", lexemes, position)
		_expressionNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _expressionNode == nil {
			did_not_match_log("expression", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("expression", lexemes, position)

		attempt_log("THEN", lexemes, position)
		_thenReservedWordNode := matchReservedWord(lexemes, position, "THEN")
		if _thenReservedWordNode == nil {
			did_not_match_optionally_log("THEN", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("THEN", lexemes, position)

		attempt_log("statementSequence", lexemes, position)
		_statementSequenceNode, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode == nil {
			did_not_match_log("statementSequence", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("statementSequence", lexemes, position)

		ifStatementNode.Children = append(ifStatementNode.Children, _elsifReservedWordNode)
		ifStatementNode.Children = append(ifStatementNode.Children, _expressionNode)
		ifStatementNode.Children = append(ifStatementNode.Children, _thenReservedWordNode)
		ifStatementNode.Children = append(ifStatementNode.Children, _statementSequenceNode)
	}

	attempt_log("ELSE", lexemes, position)
	_elseReservedWordNode := matchReservedWord(lexemes, position, "ELSE")
	if _elseReservedWordNode != nil {
		matched_log("ELSE", lexemes, position)

		attempt_log("statementSequence", lexemes, position)
		_statementSequenceNode1, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode1 == nil {
			did_not_match_log("statementSequence", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		ifStatementNode.Children = append(ifStatementNode.Children, _elseReservedWordNode)
		ifStatementNode.Children = append(ifStatementNode.Children, _statementSequenceNode1)
		matched_log("statementSequence", lexemes, position)
	} else {
		did_not_match_log("ELSE", lexemes, position)
	}

	attempt_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		did_not_match_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	ifStatementNode.Children = append(ifStatementNode.Children, _endReservedWordNode)
	matched_log("END", lexemes, position)

	return ifStatementNode, nil
}

func label(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var labelNode = new(ParseNode)
	labelNode.Label = "label"

	attempt_log("INTEGER", lexemes, position)
	_integerNode := matchtype(lexemes, position, lexer.INTEGER)
	if _integerNode != nil {
		matched_log("INTEGER", lexemes, position)
		labelNode.Children = append(labelNode.Children, _integerNode)
		return labelNode, nil
	}
	did_not_match_log("INTEGER", lexemes, position)

	attempt_log("STRING", lexemes, position)
	_stringNode := matchtype(lexemes, position, lexer.STRING)
	if _stringNode != nil {
		matched_log("STRING", lexemes, position)
		labelNode.Children = append(labelNode.Children, _stringNode)
		return labelNode, nil
	}
	did_not_match_log("STRING", lexemes, position)

	attempt_log("qualident", lexemes, position)
	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode != nil {
		matched_log("qualident", lexemes, position)
		labelNode.Children = append(labelNode.Children, _qualidentNode)
		return labelNode, nil
	}
	did_not_match_log("qualident", lexemes, position)

	return nil, nil
}

func labelRange(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var labelRangeNode = new(ParseNode)
	labelRangeNode.Label = "labelRange"
	var positionCheckpoint = *position

	attempt_log("label", lexemes, position)
	_labelNode, err := label(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _labelNode == nil {
		did_not_match_log("label", lexemes, position)
		return nil, nil
	}
	labelRangeNode.Children = append(labelRangeNode.Children, _labelNode)
	matched_log("label", lexemes, position)

	attempt_optionally_log("..", lexemes, position)
	_doubleDotOperatorNode := matchOperator(lexemes, position, "..")
	if _doubleDotOperatorNode != nil {
		optionally_matched_log("..", lexemes, position)

		attempt_log("label", lexemes, position)
		_labelNode, err := label(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _labelNode == nil {
			did_not_match_log("label", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("label", lexemes, position)

		labelRangeNode.Children = append(labelRangeNode.Children, _doubleDotOperatorNode)
		labelRangeNode.Children = append(labelRangeNode.Children, _labelNode)
	} else {
		did_not_match_optionally_log("..", lexemes, position)
	}

	return labelRangeNode, nil
}

func caseLabelList(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var caseLabelListNode = new(ParseNode)
	caseLabelListNode.Label = "caseLabelList"
	var positionCheckpoint = *position

	attempt_log("labelRange", lexemes, position)
	_labelRangeNode, err := labelRange(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _labelRangeNode == nil {
		did_not_match_log("labelRange", lexemes, position)
		return nil, nil
	}
	caseLabelListNode.Children = append(caseLabelListNode.Children, _labelRangeNode)
	matched_log("labelRange", lexemes, position)

	for {
		attempt_optionally_log(",", lexemes, position)
		_commaOperatorNode := matchOperator(lexemes, position, ",")
		if _commaOperatorNode == nil {
			did_not_match_optionally_log(",", lexemes, position)
			break
		}
		optionally_matched_log(",", lexemes, position)

		attempt_log("labelRange", lexemes, position)
		_labelRangeNode, err := labelRange(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _labelRangeNode == nil {
			did_not_match_log("labelRange", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		caseLabelListNode.Children = append(caseLabelListNode.Children, _commaOperatorNode)
		caseLabelListNode.Children = append(caseLabelListNode.Children, _labelRangeNode)
		matched_log("labelRange", lexemes, position)
	}

	return caseLabelListNode, nil
}

func _case(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var _caseNode = new(ParseNode)
	_caseNode.Label = "case"

	attempt_log("caseLabelList", lexemes, position)
	_caseLabelListNode, err := caseLabelList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _caseLabelListNode == nil {
		did_not_match_log("caseLabelList", lexemes, position)
		return _caseNode, nil
	}
	matched_log("caseLabelList", lexemes, position)

	attempt_log(":", lexemes, position)
	_colonOperatorNode := matchOperator(lexemes, position, ":")
	if _colonOperatorNode == nil {
		did_not_match_log(":", lexemes, position)
		return nil, nil
	}
	matched_log(":", lexemes, position)

	attempt_log("statementSequence", lexemes, position)
	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		did_not_match_log("statementSequence", lexemes, position)
		return nil, nil
	}
	matched_log("statementSequence", lexemes, position)

	_caseNode.Children = append(_caseNode.Children, _caseLabelListNode)
	_caseNode.Children = append(_caseNode.Children, _colonOperatorNode)
	_caseNode.Children = append(_caseNode.Children, _statementSequenceNode)

	return _caseNode, nil
}

func caseStatement(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var caseStatementNode = new(ParseNode)
	caseStatementNode.Label = "caseStatement"
	var positionCheckpoint = *position

	attempt_log("CASE", lexemes, position)
	_caseReservedWordNode := matchReservedWord(lexemes, position, "CASE")
	if _caseReservedWordNode == nil {
		did_not_match_log("CASE", lexemes, position)
		return nil, nil
	}
	matched_log("CASE", lexemes, position)

	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		did_not_match_log("expression", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("expression", lexemes, position)

	attempt_log("OF", lexemes, position)
	_ofReservedWordNode := matchReservedWord(lexemes, position, "OF")
	if _ofReservedWordNode == nil {
		did_not_match_log("OF", lexemes, position)
		return nil, nil
	}

	attempt_log("case", lexemes, position)
	_caseNode, err := _case(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _caseNode == nil {
		did_not_match_log("case", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("case", lexemes, position)

	caseStatementNode.Children = append(caseStatementNode.Children, _caseReservedWordNode)
	caseStatementNode.Children = append(caseStatementNode.Children, _expressionNode)
	caseStatementNode.Children = append(caseStatementNode.Children, _ofReservedWordNode)
	caseStatementNode.Children = append(caseStatementNode.Children, _caseNode)

	for {
		attempt_optionally_log("|", lexemes, position)
		_verticalBarReservedWordNode := matchOperator(lexemes, position, "|")
		if _verticalBarReservedWordNode == nil {
			did_not_match_optionally_log("|", lexemes, position)
			break
		}
		attempt_log("|", lexemes, position)

		matched_log("case", lexemes, position)
		_caseNode, err := _case(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _caseNode == nil {
			did_not_match_log("case", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("case", lexemes, position)

		caseStatementNode.Children = append(caseStatementNode.Children, _verticalBarReservedWordNode)
		caseStatementNode.Children = append(caseStatementNode.Children, _caseNode)
	}

	matched_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		did_not_match_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	caseStatementNode.Children = append(caseStatementNode.Children, _endReservedWordNode)
	matched_log("END", lexemes, position)

	return caseStatementNode, nil

}

func repeatStatement(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var repeatStatementNode = new(ParseNode)
	repeatStatementNode.Label = "repeatStatement"
	var positionCheckpoint = *position

	attempt_log("REPEAT", lexemes, position)
	_repeatReservedWordNode := matchReservedWord(lexemes, position, "REPEAT")
	if _repeatReservedWordNode == nil {
		did_not_match_log("REPEAT", lexemes, position)
		return nil, nil
	}
	matched_log("REPEAT", lexemes, position)

	attempt_log("statementSequence", lexemes, position)
	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		did_not_match_log("statementSequence", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("statementSequence", lexemes, position)

	attempt_log("UNTIL", lexemes, position)
	_untilReservedWordNode := matchReservedWord(lexemes, position, "UNTIL")
	if _untilReservedWordNode == nil {
		did_not_match_log("UNTIL", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("UNTIL", lexemes, position)

	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		did_not_match_log("expression", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("expression", lexemes, position)

	repeatStatementNode.Children = append(repeatStatementNode.Children, _repeatReservedWordNode)
	repeatStatementNode.Children = append(repeatStatementNode.Children, _statementSequenceNode)
	repeatStatementNode.Children = append(repeatStatementNode.Children, _untilReservedWordNode)
	repeatStatementNode.Children = append(repeatStatementNode.Children, _expressionNode)

	return repeatStatementNode, nil
}

func forStatement(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var forStatementNode = new(ParseNode)
	forStatementNode.Label = "forStatement"
	var positionCheckpoint = *position

	attempt_log("FOR", lexemes, position)
	_forReservedWordNode := matchReservedWord(lexemes, position, "FOR")
	if _forReservedWordNode == nil {
		did_not_match_log("FOR", lexemes, position)
		return nil, nil
	}
	matched_log("FOR", lexemes, position)

	attempt_log("ident", lexemes, position)
	_identNode := matchtype(lexemes, position, lexer.IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("ident", lexemes, position)

	attempt_log(":=", lexemes, position)
	_colonEqualOperatorNode := matchOperator(lexemes, position, ":=")
	if _colonEqualOperatorNode == nil {
		did_not_match_log(":=", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log(":=", lexemes, position)

	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		did_not_match_log("expression", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("expression", lexemes, position)

	attempt_log("TO", lexemes, position)
	_toReservedWordNode := matchReservedWord(lexemes, position, "TO")
	if _toReservedWordNode == nil {
		did_not_match_log("TO", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("TO", lexemes, position)

	attempt_log("expression", lexemes, position)
	_expressionNode1, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode1 == nil {
		did_not_match_log("expression", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("expression", lexemes, position)

	forStatementNode.Children = append(forStatementNode.Children, _forReservedWordNode)
	forStatementNode.Children = append(forStatementNode.Children, _identNode)
	forStatementNode.Children = append(forStatementNode.Children, _colonEqualOperatorNode)
	forStatementNode.Children = append(forStatementNode.Children, _expressionNode)
	forStatementNode.Children = append(forStatementNode.Children, _toReservedWordNode)
	forStatementNode.Children = append(forStatementNode.Children, _expressionNode1)

	attempt_optionally_log("BY", lexemes, position)
	_byReservedWordNode := matchReservedWord(lexemes, position, "BY")
	if _byReservedWordNode != nil {
		optionally_matched_log("BY", lexemes, position)

		attempt_log("constExpression", lexemes, position)
		_constExpressionNode, err := constExpression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _constExpressionNode == nil {
			did_not_match_log("constExpression", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		forStatementNode.Children = append(forStatementNode.Children, _byReservedWordNode)
		forStatementNode.Children = append(forStatementNode.Children, _constExpressionNode)
		matched_log("constExpression", lexemes, position)
	} else {
		did_not_match_optionally_log("BY", lexemes, position)
	}

	attempt_log("DO", lexemes, position)
	_doReservedWordNode := matchReservedWord(lexemes, position, "DO")
	if _doReservedWordNode == nil {
		did_not_match_log("DO", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("DO", lexemes, position)

	attempt_log("statementSequence", lexemes, position)
	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		did_not_match_log("statementSequence", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("statementSequence", lexemes, position)

	attempt_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		did_not_match_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("END", lexemes, position)

	forStatementNode.Children = append(forStatementNode.Children, _doReservedWordNode)
	forStatementNode.Children = append(forStatementNode.Children, _statementSequenceNode)
	forStatementNode.Children = append(forStatementNode.Children, _endReservedWordNode)

	return forStatementNode, nil
}

func whileStatement(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var whileStatementNode = new(ParseNode)
	whileStatementNode.Label = "whileStatement"
	var positionCheckpoint = *position

	attempt_log("WHILE", lexemes, position)
	_whileReservedWordNode := matchReservedWord(lexemes, position, "WHILE")
	if _whileReservedWordNode == nil {
		did_not_match_log("WHILE", lexemes, position)
		return nil, nil
	}
	matched_log("WHILE", lexemes, position)

	attempt_log("expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		did_not_match_log("expression", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("expression", lexemes, position)

	attempt_log("DO", lexemes, position)
	_doReservedWordNode := matchReservedWord(lexemes, position, "DO")
	if _doReservedWordNode == nil {
		did_not_match_log("DO", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("DO", lexemes, position)

	attempt_log("statementSequence", lexemes, position)
	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		did_not_match_log("statementSequence", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	whileStatementNode.Children = append(whileStatementNode.Children, _whileReservedWordNode)
	whileStatementNode.Children = append(whileStatementNode.Children, _expressionNode)
	whileStatementNode.Children = append(whileStatementNode.Children, _doReservedWordNode)
	whileStatementNode.Children = append(whileStatementNode.Children, _statementSequenceNode)
	matched_log("statementSequence", lexemes, position)

	for {
		attempt_log("ELSIF", lexemes, position)
		_elsifReservedWordNode := matchReservedWord(lexemes, position, "ELSIF")
		if _elsifReservedWordNode == nil {
			did_not_match_log("ELSIF", lexemes, position)
			break
		}
		matched_log("ELSIF", lexemes, position)

		attempt_log("expression", lexemes, position)
		_expressionNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _expressionNode == nil {
			did_not_match_log("expression", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("expression", lexemes, position)

		attempt_log("DO", lexemes, position)
		_doReservedWordNode := matchReservedWord(lexemes, position, "DO")
		if _doReservedWordNode == nil {
			did_not_match_log("DO", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("DO", lexemes, position)

		attempt_log("statementSequence", lexemes, position)
		_statementSequenceNode, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode == nil {
			did_not_match_log("statementSequence", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("statementSequence", lexemes, position)

		whileStatementNode.Children = append(whileStatementNode.Children, _elsifReservedWordNode)
		whileStatementNode.Children = append(whileStatementNode.Children, _expressionNode)
		whileStatementNode.Children = append(whileStatementNode.Children, _doReservedWordNode)
		whileStatementNode.Children = append(whileStatementNode.Children, _statementSequenceNode)
	}

	attempt_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		attempt_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	whileStatementNode.Children = append(whileStatementNode.Children, _endReservedWordNode)
	attempt_log("END", lexemes, position)

	return whileStatementNode, nil
}

func statement(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var statementNode = new(ParseNode)
	statementNode.Label = "statement"

	attempt_log("assignment", lexemes, position)
	_assignmentNode, err := assignment(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _assignmentNode != nil {
		matched_log("assignment", lexemes, position)
		statementNode.Children = append(statementNode.Children, _assignmentNode)
		return statementNode, nil
	}
	did_not_match_log("assignment", lexemes, position)

	attempt_log("procedureCall", lexemes, position)
	_procedureCallNode, err := procedureCall(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _procedureCallNode != nil {
		matched_log("procedureCall", lexemes, position)
		statementNode.Children = append(statementNode.Children, _procedureCallNode)
		return statementNode, nil
	}
	did_not_match_log("procedureCall", lexemes, position)

	attempt_log("ifStatement", lexemes, position)
	_ifStatementNode, err := ifStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _ifStatementNode != nil {
		matched_log("ifStatement", lexemes, position)
		statementNode.Children = append(statementNode.Children, _ifStatementNode)
		return statementNode, nil
	}
	did_not_match_log("ifStatement", lexemes, position)

	attempt_log("caseStatement", lexemes, position)
	_caseStatementNode, err := caseStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _caseStatementNode != nil {
		matched_log("caseStatement", lexemes, position)
		statementNode.Children = append(statementNode.Children, _caseStatementNode)
		return statementNode, nil
	}
	did_not_match_log("caseStatement", lexemes, position)

	attempt_log("whileStatement", lexemes, position)
	_whiteStatementNode, err := whileStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _whiteStatementNode != nil {
		matched_log("whileStatement", lexemes, position)
		statementNode.Children = append(statementNode.Children, _whiteStatementNode)
		return statementNode, nil
	}
	did_not_match_log("whileStatement", lexemes, position)

	attempt_log("repeatStatement", lexemes, position)
	_repeatStatementNode, err := repeatStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _repeatStatementNode != nil {
		matched_log("repeatStatement", lexemes, position)
		statementNode.Children = append(statementNode.Children, _repeatStatementNode)
		return statementNode, nil
	}
	did_not_match_log("repeatStatement", lexemes, position)

	attempt_log("forStatement", lexemes, position)
	_forStatementNode, err := forStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _forStatementNode != nil {
		matched_log("forStatement", lexemes, position)
		statementNode.Children = append(statementNode.Children, _forStatementNode)
		return statementNode, nil
	}
	did_not_match_log("forStatement", lexemes, position)

	return statementNode, nil
}

func statementSequence(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var statementSequenceNode = new(ParseNode)
	statementSequenceNode.Label = "statementSequence"

	attempt_log("statement", lexemes, position)
	_statementNode, err := statement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementNode == nil {
		attempt_log("statement", lexemes, position)
		return nil, nil
	}
	statementSequenceNode.Children = append(statementSequenceNode.Children, _statementNode)
	matched_log("statement", lexemes, position)

	// we deviate here from the spec to allow the final statement in
	// a statementSequence to be followed by an optional semi-colon
	for {
		attempt_optionally_log(";", lexemes, position)
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			did_not_match_optionally_log(";", lexemes, position)
			break
		}
		statementSequenceNode.Children = append(statementSequenceNode.Children, _semicolonNode)
		optionally_matched_log(";", lexemes, position)

		attempt_log("statement", lexemes, position)
		_statementNode, err := statement(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementNode == nil {
			did_not_match_log("statement", lexemes, position)
			break
		}
		statementSequenceNode.Children = append(statementSequenceNode.Children, _statementNode)
		matched_log("statement", lexemes, position)
	}

	return statementSequenceNode, nil
}

func procedureBody(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureBodyNode = new(ParseNode)
	procedureBodyNode.Label = "procedureBody"
	var positionCheckpoint = *position

	attempt_log("declarationSequence", lexemes, position)
	_declarationSequenceNode, err := declarationSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _declarationSequenceNode == nil {
		did_not_match_log("declarationSequence", lexemes, position)
		return nil, nil
	}
	matched_log("declarationSequence", lexemes, position)
	procedureBodyNode.Children = append(procedureBodyNode.Children, _declarationSequenceNode)

	attempt_optionally_log("BEGIN", lexemes, position)
	_beginReservedWordNode := matchReservedWord(lexemes, position, "BEGIN")
	if _beginReservedWordNode != nil {
		optionally_matched_log("BEGIN", lexemes, position)
		attempt_log("statementSequence", lexemes, position)
		_statementSequenceNode, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode == nil {
			did_not_match_log("statementSequence", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("statementSequence", lexemes, position)
		procedureBodyNode.Children = append(procedureBodyNode.Children, _beginReservedWordNode)
		procedureBodyNode.Children = append(procedureBodyNode.Children, _statementSequenceNode)
	} else {
		did_not_match_optionally_log("BEGIN", lexemes, position)
	}

	attempt_optionally_log("RETURN", lexemes, position)
	_returnReservedWordNode := matchReservedWord(lexemes, position, "RETURN")
	if _returnReservedWordNode != nil {
		optionally_matched_log("RETURN", lexemes, position)
		attempt_log("expression", lexemes, position)
		_expression, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _expression == nil {
			did_not_match_log("expression", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("expression", lexemes, position)
		procedureBodyNode.Children = append(procedureBodyNode.Children, _returnReservedWordNode)
		procedureBodyNode.Children = append(procedureBodyNode.Children, _expression)
	}

	attempt_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		did_not_match_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("END", lexemes, position)

	return procedureBodyNode, nil
}

func procedureHeading(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureHeadingNode = new(ParseNode)
	procedureHeadingNode.Label = "procedureHeading"
	var positionCheckpoint = *position

	attempt_log("PROCEDURE", lexemes, position)
	_procedureReservedWordNode := matchReservedWord(lexemes, position, "PROCEDURE")
	if _procedureReservedWordNode == nil {
		did_not_match_log("PROCEDURE", lexemes, position)
		return nil, nil
	}
	matched_log("END", lexemes, position)

	attempt_log("identdef", lexemes, position)
	_identDefNode, err := identdef(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identDefNode == nil {
		did_not_match_log("identdef", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	procedureHeadingNode.Children = append(procedureHeadingNode.Children, _procedureReservedWordNode)
	procedureHeadingNode.Children = append(procedureHeadingNode.Children, _identDefNode)
	matched_log("identdef", lexemes, position)

	attempt_log("formalParameters", lexemes, position)
	_formalParametersNode, err := formalParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _formalParametersNode != nil {
		did_not_match_log("formalParameters", lexemes, position)
		procedureHeadingNode.Children = append(procedureHeadingNode.Children, _formalParametersNode)
	}
	matched_log("formalParameters", lexemes, position)

	return procedureHeadingNode, nil
}

func procedureDeclaration(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureDeclarationNode = new(ParseNode)
	procedureDeclarationNode.Label = "procedureDeclaration"
	var positionCheckpoint = *position

	attempt_log("procedureHeading", lexemes, position)
	_pocedureHeadingNode, err := procedureHeading(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _pocedureHeadingNode == nil {
		did_not_match_log("procedureHeading", lexemes, position)
		return nil, nil
	}
	matched_log("procedureHeading", lexemes, position)

	attempt_log(";", lexemes, position)
	_semicolonNode := matchOperator(lexemes, position, ";")
	if _semicolonNode == nil {
		did_not_match_log(";", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log(";", lexemes, position)

	attempt_log("procedureBody", lexemes, position)
	_procedureBodyNode, err := procedureBody(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _procedureBodyNode == nil {
		did_not_match_log("procedureBody", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("procedureBody", lexemes, position)

	attempt_log("ident", lexemes, position)
	_identNode := matchtype(lexemes, position, lexer.IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("ident", lexemes, position)

	procedureDeclarationNode.Children = append(procedureDeclarationNode.Children, _pocedureHeadingNode)
	procedureDeclarationNode.Children = append(procedureDeclarationNode.Children, _semicolonNode)
	procedureDeclarationNode.Children = append(procedureDeclarationNode.Children, _procedureBodyNode)
	procedureDeclarationNode.Children = append(procedureDeclarationNode.Children, _identNode)

	return procedureDeclarationNode, nil
}

func varDeclaration(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var varDeclarationNode = new(ParseNode)
	varDeclarationNode.Label = "varDeclaration"
	var positionCheckpoint = *position

	attempt_log("ident", lexemes, position)
	_identListNode, err := identList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identListNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, nil
	}
	matched_log("ident", lexemes, position)

	attempt_log(":", lexemes, position)
	_colonNode := matchOperator(lexemes, position, ":")
	if _colonNode == nil {
		did_not_match_log(":", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log(":", lexemes, position)

	attempt_log("type", lexemes, position)
	_typeNode, err := _type(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _typeNode == nil {
		did_not_match_log("type", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("type", lexemes, position)

	varDeclarationNode.Children = append(varDeclarationNode.Children, _identListNode)
	varDeclarationNode.Children = append(varDeclarationNode.Children, _colonNode)
	varDeclarationNode.Children = append(varDeclarationNode.Children, _typeNode)

	return varDeclarationNode, nil
}

func constDeclaration(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var constDeclarationNode = new(ParseNode)
	constDeclarationNode.Label = "constDeclaration"

	matched_log("identdef", lexemes, position)
	_identDefNode, err := identdef(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identDefNode == nil {
		did_not_match_log("identdef", lexemes, position)
		return nil, nil
	}
	matched_log("identdef", lexemes, position)

	attempt_log("=", lexemes, position)
	_assignmentNode := matchOperator(lexemes, position, "=")
	if _assignmentNode == nil {
		did_not_match_log("=", lexemes, position)
		return nil, nil
	}
	matched_log("=", lexemes, position)

	attempt_log("constExpression", lexemes, position)
	_constExpressionNode, err := constExpression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _constExpressionNode == nil {
		did_not_match_log("constExpression", lexemes, position)
		return nil, nil
	}
	matched_log("constExpression", lexemes, position)

	constDeclarationNode.Children =
		append(constDeclarationNode.Children, _identDefNode)
	constDeclarationNode.Children =
		append(constDeclarationNode.Children, _assignmentNode)
	constDeclarationNode.Children =
		append(constDeclarationNode.Children, _constExpressionNode)

	return constDeclarationNode, nil
}

func declarationSequence_constSequence(
	lexemes *[]lexer.Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var declarationSequence_constSequenceNode = new(ParseNode)
	declarationSequence_constSequenceNode.Label = "declarationSequence_constSequence"

	// [CONST {ConstDeclaration ";"}]
	attempt_log("CONST", lexemes, position)
	_constReservedWordNode := matchReservedWord(lexemes, position, "CONST")
	if _constReservedWordNode == nil {
		did_not_match_log("CONST", lexemes, position)
		return nil, nil
	}
	matched_log("CONST", lexemes, position)
	declarationSequence_constSequenceNode.Children = append(declarationSequence_constSequenceNode.Children, _constReservedWordNode)

	for {
		attempt_log("constDeclaration", lexemes, position)
		_constDeclarationNode, err := constDeclaration(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _constDeclarationNode == nil {
			did_not_match_optionally_log("constDeclaration", lexemes, position)
			break
		}
		optionally_matched_log("constDeclaration", lexemes, position)

		attempt_log(";", lexemes, position)
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			did_not_match_optionally_log(";", lexemes, position)
			return nil, nil
		}
		matched_log(";", lexemes, position)

		declarationSequence_constSequenceNode.Children =
			append(declarationSequence_constSequenceNode.Children, _constDeclarationNode)
		declarationSequence_constSequenceNode.Children =
			append(declarationSequence_constSequenceNode.Children, _semicolonNode)
	}
	return declarationSequence_constSequenceNode, nil
}

func declarationSequence_typeDeclaration(
	lexemes *[]lexer.Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var _typeDeclarationSequenceNode = new(ParseNode)
	_typeDeclarationSequenceNode.Label = "declarationSequence_typeDeclaration"

	// [TYPE {typeDeclaration ";"}]
	attempt_log("TYPE", lexemes, position)
	_typeReservedWordNode := matchReservedWord(lexemes, position, "TYPE")
	if _typeReservedWordNode == nil {
		did_not_match_log("TYPE", lexemes, position)
		return nil, nil
	}
	_typeDeclarationSequenceNode.Children = append(_typeDeclarationSequenceNode.Children, _typeReservedWordNode)
	matched_log("TYPE", lexemes, position)

	for {
		attempt_log("typeDeclaration", lexemes, position)
		_typeDeclarationNode, err := typeDeclaration(lexemes, position)
		if err != nil {
			return _typeDeclarationSequenceNode, err
		}
		if _typeDeclarationNode == nil {
			did_not_match_log("typeDeclaration", lexemes, position)
			break
		}
		matched_log("typeDeclaration", lexemes, position)

		attempt_log(";", lexemes, position)
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			did_not_match_log(";", lexemes, position)
			return nil, nil
		}
		matched_log(";", lexemes, position)
		_typeDeclarationSequenceNode.Children =
			append(_typeDeclarationSequenceNode.Children, _typeDeclarationNode)
		_typeDeclarationSequenceNode.Children =
			append(_typeDeclarationSequenceNode.Children, _semicolonNode)
	}

	return _typeDeclarationSequenceNode, nil
}

func declarationSequence_varDeclaration(
	lexemes *[]lexer.Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var _varDeclarationSequenceNode = new(ParseNode)
	_varDeclarationSequenceNode.Label = "declarationSequence_varDeclaration"
	var positionCheckpoint = *position

	attempt_log("VAR", lexemes, position)
	_varReservedWordNode := matchReservedWord(lexemes, position, "VAR")
	if _varReservedWordNode == nil {
		did_not_match_log("VAR", lexemes, position)
		return nil, nil
	}
	_varDeclarationSequenceNode.Children = append(_varDeclarationSequenceNode.Children, _varReservedWordNode)
	matched_log("VAR", lexemes, position)

	for {
		attempt_log("varDeclaration", lexemes, position)
		_varDeclarationNode, err := varDeclaration(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _varDeclarationNode == nil {
			did_not_match_log("varDeclaration", lexemes, position)
			break
		}
		matched_log("varDeclaration", lexemes, position)

		attempt_log(";", lexemes, position)
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			did_not_match_log(";", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("varDeclaration", lexemes, position)

		_varDeclarationSequenceNode.Children =
			append(_varDeclarationSequenceNode.Children, _varDeclarationNode)
		_varDeclarationSequenceNode.Children =
			append(_varDeclarationSequenceNode.Children, _semicolonNode)
	}
	return _varDeclarationSequenceNode, nil
}

func declarationSequence_procedureDeclaration(
	lexemes *[]lexer.Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var _procedureDeclarationSequenceNode = new(ParseNode)
	_procedureDeclarationSequenceNode.Label = "declarationSequence_procedureDeclaration"

	for {
		attempt_optionally_log("procedureDeclaration", lexemes, position)
		_procedureDeclarationNode, err := procedureDeclaration(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _procedureDeclarationNode == nil {
			did_not_match_optionally_log("procedureDeclaration", lexemes, position)
			break
		}
		optionally_matched_log("procedureDeclaration", lexemes, position)

		attempt_log(";", lexemes, position)
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			did_not_match_log(";", lexemes, position)
			return nil, nil
		}
		matched_log(";", lexemes, position)

		_procedureDeclarationSequenceNode.Children =
			append(_procedureDeclarationSequenceNode.Children, _procedureDeclarationNode)
		_procedureDeclarationSequenceNode.Children =
			append(_procedureDeclarationSequenceNode.Children, _semicolonNode)
	}
	return _procedureDeclarationSequenceNode, nil
}

func declarationSequence(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var declarationSequenceNode = new(ParseNode)
	declarationSequenceNode.Label = "declarationSequence"
	var positionCheckpoint = *position

	// [CONST {ConstDeclaration ";"}]
	attempt_optionally_log("declarationSequence_constSequence", lexemes, position)
	_constDeclarationSequenceNode, err := declarationSequence_constSequence(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _constDeclarationSequenceNode != nil {
		optionally_matched_log("declarationSequence_constSequence", lexemes, position)
		declarationSequenceNode.Children =
			append(declarationSequenceNode.Children, _constDeclarationSequenceNode)
	} else {
		did_not_match_optionally_log("declarationSequence_constSequence", lexemes, position)
	}

	// [TYPE {typeDeclaration ";"}]
	attempt_optionally_log("declarationSequence_typeDeclaration", lexemes, position)
	_typeDeclarationSequenceNode, err := declarationSequence_typeDeclaration(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _typeDeclarationSequenceNode != nil {
		optionally_matched_log("declarationSequence_typeDeclaration", lexemes, position)
		declarationSequenceNode.Children =
			append(declarationSequenceNode.Children, _typeDeclarationSequenceNode)
	} else {
		did_not_match_optionally_log("declarationSequence_typeDeclaration", lexemes, position)
	}

	// [VAR {VarDeclaration ";"}]
	attempt_optionally_log("declarationSequence_varDeclaration", lexemes, position)
	_varDeclarationSequenceNode, err := declarationSequence_varDeclaration(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _varDeclarationSequenceNode != nil {
		optionally_matched_log("declarationSequence_varDeclaration", lexemes, position)
		declarationSequenceNode.Children =
			append(declarationSequenceNode.Children, _varDeclarationSequenceNode)
	} else {
		did_not_match_optionally_log("declarationSequence_varDeclaration", lexemes, position)
	}

	// {ProcedureDeclaration ";"}
	attempt_optionally_log("declarationSequence_procedureDeclaration", lexemes, position)
	_procedureDeclarationSequenceNode, err := declarationSequence_procedureDeclaration(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _procedureDeclarationSequenceNode != nil {
		optionally_matched_log("declarationSequence_procedureDeclaration", lexemes, position)
		declarationSequenceNode.Children =
			append(declarationSequenceNode.Children, _procedureDeclarationSequenceNode)
	} else {
		did_not_match_optionally_log("declarationSequence_procedureDeclaration", lexemes, position)
	}

	return declarationSequenceNode, nil
}

func module(
	lexemes *[]lexer.Lexeme,
	position *int,
) (*ParseNode, error) {
	var moduleNode = new(ParseNode)
	moduleNode.Label = "module"

	// MODULE
	attempt_log("MODULE", lexemes, position)
	_moduleNode := matchReservedWord(lexemes, position, "MODULE")
	if _moduleNode == nil {
		did_not_match_log("MODULE", lexemes, position)
		return nil, fmt.Errorf("parse error: expected 'MODULE', found %v", (*lexemes)[*position])
	}
	matched_log("MODULE", lexemes, position)
	moduleNode.Children = append(moduleNode.Children, _moduleNode)

	// ident
	attempt_log("ident", lexemes, position)
	_identNode := matchtype(lexemes, position, lexer.IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, fmt.Errorf("parse error: expected identNode, found %v", (*lexemes)[*position])
	}
	matched_log("ident", lexemes, position)
	moduleNode.Children = append(moduleNode.Children, _identNode)

	// ;
	attempt_log(";", lexemes, position)
	_semicolonNode := matchOperator(lexemes, position, ";")
	if _semicolonNode == nil {
		did_not_match_log(";", lexemes, position)
		return nil, fmt.Errorf("parse error: expected ';', found %v", (*lexemes)[*position])
	}
	matched_log(";", lexemes, position)
	moduleNode.Children = append(moduleNode.Children, _identNode)

	// [ImportList]
	attempt_optionally_log("importList", lexemes, position)
	_importListNode, err := importList(lexemes, position)
	if err != nil {
		did_not_match_optionally_log("importList", lexemes, position)
		return nil, err
	}
	if _importListNode != nil {
		optionally_matched_log("importList", lexemes, position)
		moduleNode.Children = append(moduleNode.Children, _importListNode)
	}

	// DeclarationSequence
	attempt_log("declarationSequence", lexemes, position)
	_declarationSequenceNode, err := declarationSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _declarationSequenceNode == nil {
		did_not_match_log("declarationSequence", lexemes, position)
		return nil, fmt.Errorf("parse error: expected declarationSequence, found %v", (*lexemes)[*position])
	}
	matched_log("declarationSequence", lexemes, position)
	moduleNode.Children = append(moduleNode.Children, _declarationSequenceNode)

	// [BEGIN StatementSequence]
	attempt_optionally_log("BEGIN", lexemes, position)
	_beginNode := matchReservedWord(lexemes, position, "BEGIN")
	if _beginNode != nil {
		optionally_matched_log("BEGIN", lexemes, position)
		moduleNode.Children = append(moduleNode.Children, _beginNode)
		attempt_log("statementSequence", lexemes, position)
		_statementSequenceNode, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode == nil {
			did_not_match_log("statementSequence", lexemes, position)
			return nil, parse_error("statementSequence", lexemes, position)
		}
		matched_log("statementSequence", lexemes, position)
		moduleNode.Children = append(moduleNode.Children, _statementSequenceNode)
	} else {
		did_not_match_optionally_log("BEGIN", lexemes, position)
	}

	attempt_log("END", lexemes, position)
	_endNode := matchReservedWord(lexemes, position, "END")
	if _endNode == nil {
		did_not_match_log("END", lexemes, position)
		return nil, parse_error("END", lexemes, position)
	}
	matched_log("END", lexemes, position)
	moduleNode.Children = append(moduleNode.Children, _endNode)

	attempt_log("ident", lexemes, position)
	_identNode1 := matchtype(lexemes, position, lexer.IDENT)
	if _identNode1 == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, parse_error("ident", lexemes, position)
	}
	matched_log("ident", lexemes, position)
	moduleNode.Children = append(moduleNode.Children, _identNode1)

	attempt_log(".", lexemes, position)
	_dotOperatorNode := matchOperator(lexemes, position, ".")
	if _dotOperatorNode == nil {
		did_not_match_log(".", lexemes, position)
		return nil, parse_error(".", lexemes, position)
	}
	matched_log(".", lexemes, position)
	moduleNode.Children = append(moduleNode.Children, _dotOperatorNode)

	return moduleNode, nil
}

func parser(lexemes *[]lexer.Lexeme, debug bool) (*ParseNode, error) {
	logging.SetBackend(parser_log_backend_formatter)
	parserDebug = debug
	var position = 0
	tree, err := module(lexemes, &position)
	if err != nil {
		return nil, err
	}
	if position < len(*lexemes) {
		unparsedToken := (*lexemes)[position]
		return nil, fmt.Errorf("parse error: unparsed token: %v at (line: %d, column: %d), token number: %d", unparsedToken.Label, unparsedToken.Line, unparsedToken.Column, position)
	}
	return tree, err
}
