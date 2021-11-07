package main

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
)

var LOG = logging.MustGetLogger("parser")
var backend = logging.NewLogBackend(os.Stdout, "", 0)
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{longfile}#%{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
)

//var backendLeveled = logging.AddModuleLevel(backend)
var backendFormatter = logging.NewBackendFormatter(backend, format)
var parserDebug = false

type ParseNode struct {
	label    string
	children []*ParseNode
}

func parse_error(
	message string,
	lexemes *[]Lexeme,
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
	lexemes *[]Lexeme,
	position *int,
) {
	if !parserDebug {
		return
	}
	if *position < len(*lexemes) {
		LOG.Debug(fmt.Sprintf("%s (current_token: %v, position: %d)", message, ((*lexemes)[*position]), *position))
	}
}

func attempt_log(
	message string,
	lexemes *[]Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Attempting to match %s", message), lexemes, position)
}

func attempt_optionally_log(
	message string,
	lexemes *[]Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Attempting to optionally match %s", message), lexemes, position)
}

func did_not_match_log(
	message string,
	lexemes *[]Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Did not match %s", message), lexemes, position)
}

func did_not_match_optionally_log(
	message string,
	lexemes *[]Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Did not optionally match %s", message), lexemes, position)
}

func matched_log(
	message string,
	lexemes *[]Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Matched %s", message), lexemes, position)
}

func optionally_matched_log(
	message string,
	lexemes *[]Lexeme,
	position *int,
) {
	debug(fmt.Sprintf("Optionally matched %s", message), lexemes, position)
}

func matchReservedWord(
	lexemes *[]Lexeme,
	position *int,
	terminal string,
) *ParseNode {
	if *position >= len(*lexemes) {
		return nil
	}
	lexeme := (*lexemes)[*position]
	if lexeme.label == terminal {
		var terminalNode = new(ParseNode)
		(*terminalNode).label = lexeme.label
		(*position)++
		return terminalNode
	}
	return nil
}

func matchOperator(
	lexemes *[]Lexeme,
	position *int,
	operator string,
) *ParseNode {
	if *position >= len(*lexemes) {
		return nil
	}
	lexeme := (*lexemes)[*position]
	if lexeme.typ == OP_OR_DELIM && lexeme.label == operator {
		var terminalNode = new(ParseNode)
		(*terminalNode).label = lexeme.label
		(*position)++
		return terminalNode
	}
	return nil
}

func matchType(
	lexemes *[]Lexeme,
	position *int,
	lexemeType LexemeType,
) *ParseNode {
	if *position >= len(*lexemes) {
		return nil
	}
	lexeme := (*lexemes)[*position]
	if lexeme.typ == lexemeType {
		var terminalNode = new(ParseNode)
		(*terminalNode).label = lexeme.label
		(*position)++
		return terminalNode
	}
	return nil
}

// import = ident [":=" ident].
func _import(
	lexemes *[]Lexeme,
	position *int,
) *ParseNode {
	var importNode = new(ParseNode)

	// ident
	attempt_log("ident", lexemes, position)
	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil
	}
	matched_log("ident", lexemes, position)

	importNode.children = append(importNode.children, _identNode)

	// :=
	attempt_log(":=", lexemes, position)
	_assignmentOperator := matchOperator(lexemes, position, ":=")
	if _assignmentOperator != nil {
		// ident
		matched_log(":=", lexemes, position)
		attempt_log("ident", lexemes, position)
		_identNode := matchType(lexemes, position, IDENT)
		if _identNode == nil {
			did_not_match_log("ident", lexemes, position)
			return nil
		}
		matched_log("ident", lexemes, position)
		importNode.children = append(importNode.children, _assignmentOperator)
		importNode.children = append(importNode.children, _identNode)
	}
	return importNode
}

// ImportList = IMPORT import {"," import} ";".
func importList(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var importListNode = new(ParseNode)
	var positionCheckpoint = *position

	// IMPORT
	attempt_log("IMPORT", lexemes, position)
	_importReservedWordNode := matchReservedWord(lexemes, position, "IMPORT")
	if _importReservedWordNode == nil {
		did_not_match_log("IMPORT", lexemes, position)
		return nil, nil
	}
	matched_log("IMPORT", lexemes, position)
	importListNode.children = append(importListNode.children, _importReservedWordNode)

	// import
	attempt_log("import", lexemes, position)
	_importNode := _import(lexemes, position)
	if _importNode == nil {
		did_not_match_log("import", lexemes, position)
		return nil, nil
	}
	matched_log("import", lexemes, position)
	importListNode.children = append(importListNode.children, _importNode)

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
		importListNode.children = append(importListNode.children, _commaNode)
		importListNode.children = append(importListNode.children, _additionalImportNode)
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
	importListNode.children = append(importListNode.children, _semicolonNode)

	return importListNode, nil
}

// qualident = [ident "."] ident.
func qualident(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var qualidentNode = new(ParseNode)
	var positionCheckpoint = *position

	// [ident "."]
	attempt_log("ident", lexemes, position)
	var _identNode = matchType(lexemes, position, IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, nil
	}
	qualidentNode.children = append(qualidentNode.children, _identNode)
	matched_log("ident", lexemes, position)
	positionCheckpoint = *position

	attempt_log(".", lexemes, position)
	_dotOperatorNode := matchOperator(lexemes, position, ".")
	if _dotOperatorNode != nil {
		matched_log(".", lexemes, position)

		attempt_log("ident", lexemes, position)
		_identNode := matchType(lexemes, position, IDENT)
		if _identNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		qualidentNode.children = append(qualidentNode.children, _dotOperatorNode)
		qualidentNode.children = append(qualidentNode.children, _identNode)
	}

	return qualidentNode, nil
}

// ExpList = expression {"," expression}.
func expList(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var expListNode = new(ParseNode)
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
	expListNode.children = append(expListNode.children, _expressionNode)

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

		expListNode.children = append(expListNode.children, _commaOperatorNode)
		expListNode.children = append(expListNode.children, _expressionNode)
	}
	return expListNode, nil
}

// selector = "." ident | "[" ExpList "]" | "^" | "(" qualident ")".
func selector(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var selectorNode = new(ParseNode)
	var positionCheckpoint = *position

	// "." ident
	attempt_optionally_log(".", lexemes, position)
	_dotOperatorNode := matchOperator(lexemes, position, ".")
	if _dotOperatorNode != nil {
		optionally_matched_log(".", lexemes, position)

		attempt_log("ident", lexemes, position)
		_identNode := matchType(lexemes, position, IDENT)
		if _identNode == nil {
			did_not_match_log("ident", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		did_not_match_log("ident", lexemes, position)
		selectorNode.children = append(selectorNode.children, _dotOperatorNode)
		selectorNode.children = append(selectorNode.children, _identNode)

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

		selectorNode.children = append(selectorNode.children, _leftBracketNode)
		selectorNode.children = append(selectorNode.children, _expListNode)
		selectorNode.children = append(selectorNode.children, _rightBracketNode)

		return selectorNode, nil
	} else {
		did_not_match_optionally_log("[", lexemes, position)
	}

	// ^
	attempt_optionally_log("^", lexemes, position)
	_caratOperatorNode := matchOperator(lexemes, position, "^")
	if _caratOperatorNode != nil {
		optionally_matched_log("^", lexemes, position)
		selectorNode.children = append(selectorNode.children, _caratOperatorNode)
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

		selectorNode.children = append(selectorNode.children, _leftParenNode)
		selectorNode.children = append(selectorNode.children, _qualidentNode)
		selectorNode.children = append(selectorNode.children, _rightParenNode)

		return selectorNode, nil
	} else {
		did_not_match_optionally_log("(", lexemes, position)
	}

	return nil, nil
}

// designator = qualident {selector}.
func designator(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var designatorNode = new(ParseNode)

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
	designatorNode.children = append(designatorNode.children, _qualidentNode)

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
		designatorNode.children = append(designatorNode.children, _selectorNode)
	}
	return designatorNode, nil
}

// element = expression [".." expression].
func element(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var elementNode = new(ParseNode)

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
	elementNode.children = append(elementNode.children, _expressionNode)

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
		elementNode.children = append(elementNode.children, _doubleDotOperator)
		elementNode.children = append(elementNode.children, _elementNode)
	}

	return elementNode, nil
}

// set = "{" [element {"," element}] "}".
func set(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var setNode = new(ParseNode)
	var positionCheckpoint = *position

	// {
	attempt_log("{", lexemes, position)
	_leftbraceNode := matchOperator(lexemes, position, "{")
	if _leftbraceNode == nil {
		did_not_match_log("{", lexemes, position)
		return nil, nil
	}
	matched_log("{", lexemes, position)
	setNode.children = append(setNode.children, _leftbraceNode)

	// element
	attempt_optionally_log("element", lexemes, position)
	_elementNode, err := element(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _elementNode != nil {
		optionally_matched_log("element", lexemes, position)
		setNode.children = append(setNode.children, _elementNode)
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
			setNode.children = append(setNode.children, _commaNode)
			setNode.children = append(setNode.children, _elementNode)
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
	setNode.children = append(setNode.children, _rightBraceNode)

	return setNode, nil
}

// "(" [ExpList] ")".
func actualParameters(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var actualParametersNode = new(ParseNode)
	var positionCheckpoint = *position

	// "("
	attempt_log("(", lexemes, position)
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode == nil {
		did_not_match_log("(", lexemes, position)
		return nil, nil
	}
	actualParametersNode.children = append(actualParametersNode.children, _leftParenNode)
	matched_log("(", lexemes, position)

	// [expList]
	attempt_optionally_log("expList", lexemes, position)
	_expListNode, err := expList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expListNode != nil {
		did_not_match_optionally_log("expList", lexemes, position)
		actualParametersNode.children = append(actualParametersNode.children, _expListNode)
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

	actualParametersNode.children = append(actualParametersNode.children, _leftParenNode)
	return actualParametersNode, nil
}

// factor = number | string | NIL | TRUE | FALSE | set | designator [ActualParameters] | "(" expression ")" | "~" factor.
func factor(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var factorNode = new(ParseNode)
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
				factorNode.children = append(factorNode.children, _minusOperatorNode)
			} else {
				did_not_match_optionally_log("-", lexemes, position)
			}
		} else {
			optionally_matched_log("+", lexemes, position)
			factorNode.children = append(factorNode.children, _plusOperatorNode)
		}

		attempt_log("integer", lexemes, position)
		_integerNode := matchType(lexemes, position, INTEGER)
		if _integerNode != nil {
			matched_log("integer", lexemes, position)
			factorNode.children = append(factorNode.children, _integerNode)
			debug("Matched integer", lexemes, position)
			return factorNode, nil
		}
		did_not_match_log("integer", lexemes, position)

		attempt_log("real", lexemes, position)
		_realNode := matchType(lexemes, position, REAL)
		if _realNode != nil {
			matched_log("real", lexemes, position)
			factorNode.children = append(factorNode.children, _realNode)
			return factorNode, nil
		}
		did_not_match_log("real", lexemes, position)

		*position = positionCheckpoint
		// TODO: determine whether it is sage to return nil, nil here
	}

	// string
	attempt_log("string", lexemes, position)
	_stringNode := matchType(lexemes, position, STRING)
	if _stringNode != nil {
		matched_log("string", lexemes, position)
		factorNode.children = append(factorNode.children, _stringNode)
		return factorNode, nil
	}
	did_not_match_log("string", lexemes, position)

	// NIL
	attempt_log("NIL", lexemes, position)
	_nilNode := matchReservedWord(lexemes, position, "NIL")
	if _nilNode != nil {
		matched_log("NIL", lexemes, position)
		factorNode.children = append(factorNode.children, _nilNode)
		return factorNode, nil
	}
	did_not_match_log("NIL", lexemes, position)

	// TRUE
	attempt_log("TRUE", lexemes, position)
	_trueNode := matchReservedWord(lexemes, position, "TRUE")
	if _trueNode != nil {
		matched_log("TRUE", lexemes, position)
		factorNode.children = append(factorNode.children, _trueNode)
		return factorNode, nil
	}
	did_not_match_log("TRUE", lexemes, position)

	// FALSE
	attempt_log("FALSE", lexemes, position)
	_falseNode := matchReservedWord(lexemes, position, "FALSE")
	if _falseNode != nil {
		matched_log("FALSE", lexemes, position)
		debug("Matched FALSE", lexemes, position)
		factorNode.children = append(factorNode.children, _falseNode)
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
		factorNode.children = append(factorNode.children, _setNode)
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
		factorNode.children = append(factorNode.children, _designatorNode)

		optionally_matched_log("actualParameters", lexemes, position)
		_actualParametersNode, err := actualParameters(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil != _actualParametersNode {
			optionally_matched_log("actualParameters", lexemes, position)
			factorNode.children = append(factorNode.children, _actualParametersNode)
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

		factorNode.children = append(factorNode.children, _leftParenNode)
		factorNode.children = append(factorNode.children, _expressionNode)
		factorNode.children = append(factorNode.children, _rightParenNode)

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

	factorNode.children = append(factorNode.children, _tildeOperator)
	factorNode.children = append(factorNode.children, _factorNode)

	return factorNode, nil
}

func mulOperator(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var mulOperatorNode = new(ParseNode)

	attempt_log("*", lexemes, position)
	_asterixOperatorNode := matchOperator(lexemes, position, "*")
	if _asterixOperatorNode != nil {
		matched_log("*", lexemes, position)
		mulOperatorNode.children = append(mulOperatorNode.children, _asterixOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("*", lexemes, position)

	attempt_log("/", lexemes, position)
	_divOperatorNode := matchOperator(lexemes, position, "/")
	if _divOperatorNode != nil {
		matched_log("/", lexemes, position)
		mulOperatorNode.children = append(mulOperatorNode.children, _divOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("/", lexemes, position)

	attempt_log("DIV", lexemes, position)
	_divNode := matchReservedWord(lexemes, position, "DIV")
	if _divNode != nil {
		matched_log("DIV", lexemes, position)
		mulOperatorNode.children = append(mulOperatorNode.children, _divNode)
		return mulOperatorNode, nil
	}
	matched_log("DIV", lexemes, position)

	attempt_log("MOD", lexemes, position)
	_modeOperatorNode := matchReservedWord(lexemes, position, "MOD")
	if _modeOperatorNode != nil {
		matched_log("MOD", lexemes, position)
		mulOperatorNode.children = append(mulOperatorNode.children, _modeOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("MOD", lexemes, position)

	attempt_log("&", lexemes, position)
	_ampersandOperatorNode := matchOperator(lexemes, position, "&")
	if _ampersandOperatorNode != nil {
		matched_log("&", lexemes, position)
		mulOperatorNode.children = append(mulOperatorNode.children, _ampersandOperatorNode)
		return mulOperatorNode, nil
	}
	matched_log("&", lexemes, position)

	return nil, nil
}

func term(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var termNode = new(ParseNode)
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
	termNode.children = append(termNode.children, _factorNode)
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

		termNode.children = append(termNode.children, _mulOperatorNode)
		termNode.children = append(termNode.children, _factorNode)
		positionCheckpoint = *position
	}
	return termNode, nil
}

func addOperator(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var addOperatorNode = new(ParseNode)

	attempt_optionally_log("+", lexemes, position)
	_plusOperatorNode := matchOperator(lexemes, position, "+")
	if _plusOperatorNode != nil {
		matched_log("+", lexemes, position)
		addOperatorNode.children = append(addOperatorNode.children, _plusOperatorNode)
		return addOperatorNode, nil
	} else {
		did_not_match_optionally_log("+", lexemes, position)
	}

	attempt_optionally_log("-", lexemes, position)
	_minusOperatorNode := matchOperator(lexemes, position, "-")
	if _minusOperatorNode != nil {
		matched_log("-", lexemes, position)
		addOperatorNode.children = append(addOperatorNode.children, _minusOperatorNode)
		return addOperatorNode, nil
	} else {
		did_not_match_optionally_log("-", lexemes, position)
	}

	attempt_optionally_log("OR", lexemes, position)
	_orOperatorNode := matchReservedWord(lexemes, position, "OR")
	if _orOperatorNode != nil {
		matched_log("OR", lexemes, position)
		addOperatorNode.children = append(addOperatorNode.children, _orOperatorNode)
		return addOperatorNode, nil
	} else {
		did_not_match_optionally_log("OR", lexemes, position)
	}

	return nil, nil
}

func simpleExpression(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var simpleExpressionNode = new(ParseNode)
	var positionCheckpoint = *position

	attempt_optionally_log("+", lexemes, position)
	_plusOperatorNode := matchOperator(lexemes, position, "+")
	if nil == _plusOperatorNode {
		did_not_match_optionally_log("+", lexemes, position)

		attempt_optionally_log("-", lexemes, position)
		_minusOperatorNode := matchOperator(lexemes, position, "-")
		if nil != _minusOperatorNode {
			optionally_matched_log("-", lexemes, position)
			simpleExpressionNode.children = append(simpleExpressionNode.children, _minusOperatorNode)
		} else {
			did_not_match_optionally_log("-", lexemes, position)
		}
	} else {
		optionally_matched_log("+", lexemes, position)
		simpleExpressionNode.children = append(simpleExpressionNode.children, _plusOperatorNode)
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
	simpleExpressionNode.children = append(simpleExpressionNode.children, _termNode)
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
		simpleExpressionNode.children = append(simpleExpressionNode.children, _addOperatorNode)
		simpleExpressionNode.children = append(simpleExpressionNode.children, _termNode)
	}
	return simpleExpressionNode, nil
}

func relation(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var relationOperatorNode = new(ParseNode)

	attempt_log("=", lexemes, position)
	_equalOperatorNode := matchOperator(lexemes, position, "=")
	if _equalOperatorNode != nil {
		matched_log("=", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _equalOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("=", lexemes, position)

	attempt_log("#", lexemes, position)
	_hashOperatorNode := matchOperator(lexemes, position, "#")
	if _hashOperatorNode != nil {
		matched_log("#", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _hashOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("#", lexemes, position)

	attempt_log("<", lexemes, position)
	_lessThanOperatorNode := matchOperator(lexemes, position, "<")
	if _lessThanOperatorNode != nil {
		matched_log("<", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _lessThanOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("<", lexemes, position)

	attempt_log("<=", lexemes, position)
	_lessThanEqualOperatorNode := matchOperator(lexemes, position, "<=")
	if _lessThanEqualOperatorNode != nil {
		matched_log("<=", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _lessThanEqualOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("<=", lexemes, position)

	attempt_log(">", lexemes, position)
	_greaterThanOperatorNode := matchOperator(lexemes, position, ">")
	if _greaterThanOperatorNode != nil {
		matched_log(">", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _greaterThanOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log(">", lexemes, position)

	attempt_log(">=", lexemes, position)
	_greaterThanEqualOperatorNode := matchOperator(lexemes, position, ">=")
	if _greaterThanEqualOperatorNode != nil {
		matched_log(">=", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _greaterThanEqualOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log(">=", lexemes, position)

	attempt_log("IN", lexemes, position)
	_inOperatorNode := matchReservedWord(lexemes, position, "IN")
	if _inOperatorNode != nil {
		matched_log("IN", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _inOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("IN", lexemes, position)

	attempt_log("IS", lexemes, position)
	_isOperatorNode := matchReservedWord(lexemes, position, "IS")
	if _isOperatorNode != nil {
		matched_log("IS", lexemes, position)
		relationOperatorNode.children = append(relationOperatorNode.children, _isOperatorNode)
		return relationOperatorNode, nil
	}
	did_not_match_log("IS", lexemes, position)

	return nil, nil
}

func expression(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var expressionNode = new(ParseNode)
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
	expressionNode.children = append(expressionNode.children, _simpleExpressionNode)
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

		expressionNode.children = append(expressionNode.children, _relationNode)
		expressionNode.children = append(expressionNode.children, _simpleExpressionNode)
	}
	return expressionNode, nil
}

func length(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	return constExpression(lexemes, position)
}

func _type(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var _typeNode = new(ParseNode)

	attempt_log("qualident", lexemes, position)
	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode != nil {
		matched_log("qualident", lexemes, position)
		_typeNode.children = append(_typeNode.children, _qualidentNode)
		return _typeNode, nil
	}
	did_not_match_log("qualident", lexemes, position)

	attempt_log("strucType", lexemes, position)
	_strucTypeNode, err := strucType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _strucTypeNode != nil {
		matched_log("strucType", lexemes, position)
		_typeNode.children = append(_typeNode.children, _strucTypeNode)
		return _typeNode, nil
	}
	did_not_match_log("strucType", lexemes, position)

	return nil, nil
}

func arrayType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var arrayTypeNode = new(ParseNode)
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
	arrayTypeNode.children = append(arrayTypeNode.children, _arrayReservedWord)
	arrayTypeNode.children = append(arrayTypeNode.children, _lengthNode)
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
		arrayTypeNode.children = append(arrayTypeNode.children, _commaOperatorNode)
		arrayTypeNode.children = append(arrayTypeNode.children, _lengthNode)
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

	arrayTypeNode.children = append(arrayTypeNode.children, _ofReservedWordNode)
	arrayTypeNode.children = append(arrayTypeNode.children, _typeNode)

	return arrayTypeNode, nil
}

func baseType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	return qualident(lexemes, position)
}

func identList(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var identListNode = new(ParseNode)
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
	identListNode.children = append(identListNode.children, _identDefNode)
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

		identListNode.children = append(identListNode.children, _commaOperatorNode)
		identListNode.children = append(identListNode.children, _identDefNode)
	}
	return identListNode, nil

}

func fieldList(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var fieldListNode = new(ParseNode)
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

	fieldListNode.children = append(fieldListNode.children, _identListNode)
	fieldListNode.children = append(fieldListNode.children, _colonNode)
	fieldListNode.children = append(fieldListNode.children, _typeNode)

	return fieldListNode, nil
}

func fieldListSequence(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var fieldListSequenceNode = new(ParseNode)

	attempt_log("fieldListSequence", lexemes, position)
	_fieldListNode, err := fieldList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _fieldListNode == nil {
		attempt_log("fieldListSequence", lexemes, position)
		return nil, nil
	}
	fieldListSequenceNode.children = append(fieldListSequenceNode.children, _fieldListNode)
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
		fieldListSequenceNode.children = append(fieldListSequenceNode.children, _semicolonNode)
		fieldListSequenceNode.children = append(fieldListSequenceNode.children, _fieldListNode)
		matched_log("fieldList", lexemes, position)
	}
	return fieldListSequenceNode, nil
}

func recordType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var recordTypeNode = new(ParseNode)
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
		attempt_log("baseType", lexemes, position)
		_baseTypeNode, err := baseType(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _baseTypeNode == nil {
			did_not_match_log("baseType", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("baseType", lexemes, position)

		attempt_log(")", lexemes, position)
		_rightParenNode := matchOperator(lexemes, position, ")")
		if _rightParenNode == nil {
			did_not_match_log(")", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log(")", lexemes, position)

		recordTypeNode.children = append(recordTypeNode.children, _leftParenNode)
		recordTypeNode.children = append(recordTypeNode.children, _baseTypeNode)
		recordTypeNode.children = append(recordTypeNode.children, _rightParenNode)
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
		recordTypeNode.children = append(recordTypeNode.children, _fieldListSequenceNode)
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
	recordTypeNode.children = append(recordTypeNode.children, _endReservedWordNode)

	return recordTypeNode, nil
}

func pointerType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var pointerTypeNode = new(ParseNode)
	var positionCheckpoint = *position

	attempt_log("POINTER", lexemes, position)
	_pointerTypeNode := matchReservedWord(lexemes, position, "POINTER")
	if _pointerTypeNode == nil {
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

	pointerTypeNode.children = append(pointerTypeNode.children, _pointerTypeNode)
	pointerTypeNode.children = append(pointerTypeNode.children, _toNode)
	pointerTypeNode.children = append(pointerTypeNode.children, _typeNode)

	return pointerTypeNode, nil
}

func formalType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var formalTypeNode = new(ParseNode)
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
		formalTypeNode.children = append(formalTypeNode.children, _arrayReservedNode)
		formalTypeNode.children = append(formalTypeNode.children, _ofReservedNode)
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
	formalTypeNode.children = append(formalTypeNode.children, _qualidentNode)
	matched_log("qualident", lexemes, position)

	return formalTypeNode, nil
}

func fpSection(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var fpSectionNode = new(ParseNode)
	var positionCheckpoint = *position

	attempt_log("VAR", lexemes, position)
	_varNode := matchReservedWord(lexemes, position, "VAR")
	if _varNode != nil {
		matched_log("VAR", lexemes, position)
		fpSectionNode.children = append(fpSectionNode.children, _varNode)
	} else {
		did_not_match_log("VAR", lexemes, position)
	}

	attempt_log("ident", lexemes, position)
	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, nil
	}
	matched_log("ident", lexemes, position)
	fpSectionNode.children = append(fpSectionNode.children, _identNode)

	for {
		attempt_optionally_log(",", lexemes, position)
		_commaNode := matchOperator(lexemes, position, ",")
		if _commaNode == nil {
			did_not_match_optionally_log(",", lexemes, position)
			break
		}
		optionally_matched_log(",", lexemes, position)

		attempt_log("ident", lexemes, position)
		_identNode := matchType(lexemes, position, IDENT)
		if _identNode == nil {
			did_not_match_log("ident", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		matched_log("ident", lexemes, position)

		fpSectionNode.children = append(fpSectionNode.children, _commaNode)
		fpSectionNode.children = append(fpSectionNode.children, _identNode)
	}

	attempt_log(":", lexemes, position)
	_colonNode := matchOperator(lexemes, position, ":")
	if _colonNode == nil {
		did_not_match_log(":", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	fpSectionNode.children = append(fpSectionNode.children, _colonNode)
	matched_log(":", lexemes, position)

	attempt_log("formalType", lexemes, position)
	_formalTypeNode, err := formalType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _formalTypeNode == nil {
		did_not_match_log("formalType", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("formalType", lexemes, position)
	fpSectionNode.children = append(fpSectionNode.children, _formalTypeNode)

	return fpSectionNode, nil
}

func formalParameters(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var formalParametersNode = new(ParseNode)
	var positionCheckpoint = *position

	attempt_log("(", lexemes, position)
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode == nil {
		did_not_match_log("(", lexemes, position)
		return nil, nil
	}
	matched_log("(", lexemes, position)
	formalParametersNode.children = append(formalParametersNode.children, _leftParenNode)

	attempt_optionally_log("fpSectionNode", lexemes, position)
	_fpSectionNode, err := fpSection(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _fpSectionNode != nil {
		optionally_matched_log("fpSectionNode", lexemes, position)
		formalParametersNode.children = append(formalParametersNode.children, _fpSectionNode)
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

			formalParametersNode.children = append(formalParametersNode.children, _semicolonNode)
			formalParametersNode.children = append(formalParametersNode.children, _fpSectionNode)
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
	formalParametersNode.children = append(formalParametersNode.children, _rightParenNode)

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
		formalParametersNode.children = append(formalParametersNode.children, _colonNode)
		formalParametersNode.children = append(formalParametersNode.children, _qualidentNode)
	} else {
		did_not_match_log(":", lexemes, position)
	}
	return formalParametersNode, nil
}

func procedureType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureTypeNode = new(ParseNode)

	attempt_log("PROCEDURE", lexemes, position)
	_procedureNode := matchReservedWord(lexemes, position, "PROCEDURE")
	if _procedureNode == nil {
		did_not_match_log("PROCEDURE", lexemes, position)
		return nil, nil
	}
	procedureTypeNode.children = append(procedureTypeNode.children, _procedureNode)
	matched_log("PROCEDURE", lexemes, position)

	attempt_optionally_log("formalParameters", lexemes, position)
	_formalParametersNode, err := formalParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _formalParametersNode != nil {
		matched_log("formalParameters", lexemes, position)
		procedureTypeNode.children = append(procedureTypeNode.children, _formalParametersNode)
	} else {
		did_not_match_optionally_log("formalParameters", lexemes, position)
	}
	return procedureTypeNode, nil
}

func strucType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var strucTypeNode = new(ParseNode)

	attempt_log("arrayType", lexemes, position)
	_arrayTypeNode, err := arrayType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _arrayTypeNode != nil {
		matched_log("arrayType", lexemes, position)
		strucTypeNode.children = append(strucTypeNode.children, _arrayTypeNode)
		return strucTypeNode, nil
	}
	did_not_match_log("arrayType", lexemes, position)

	attempt_log("recordType", lexemes, position)
	_recordTypeNode, err := recordType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _recordTypeNode != nil {
		matched_log("recordType", lexemes, position)
		strucTypeNode.children = append(strucTypeNode.children, _recordTypeNode)
		return strucTypeNode, nil
	}
	did_not_match_log("recordType", lexemes, position)

	attempt_log("pointerType", lexemes, position)
	_pointerTypeNode, err := pointerType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _pointerTypeNode != nil {
		matched_log("pointerType", lexemes, position)
		strucTypeNode.children = append(strucTypeNode.children, _pointerTypeNode)
		return strucTypeNode, nil
	}
	did_not_match_log("pointerType", lexemes, position)

	attempt_log("procedureType", lexemes, position)
	_procedureTypeNode, err := procedureType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _procedureTypeNode != nil {
		matched_log("procedureType", lexemes, position)
		strucTypeNode.children = append(strucTypeNode.children, _procedureTypeNode)
		return strucTypeNode, nil
	}
	did_not_match_log("procedureType", lexemes, position)

	return nil, nil
}

func typeDeclaration(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var typeDeclarationNode = new(ParseNode)
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

	attempt_log("strucType", lexemes, position)
	_strucTypeNode, err := strucType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _strucTypeNode == nil {
		did_not_match_log("strucType", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("strucType", lexemes, position)

	typeDeclarationNode.children = append(typeDeclarationNode.children, _identDefNode)
	typeDeclarationNode.children = append(typeDeclarationNode.children, _equalOperatorNode)
	typeDeclarationNode.children = append(typeDeclarationNode.children, _strucTypeNode)

	return typeDeclarationNode, nil
}

func identdef(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var identdefNode = new(ParseNode)

	attempt_log("ident", lexemes, position)
	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, nil
	}
	identdefNode.children = append(identdefNode.children, _identNode)
	matched_log("ident", lexemes, position)

	attempt_log("*", lexemes, position)
	_asterixNode := matchOperator(lexemes, position, "*")
	if _asterixNode != nil {
		did_not_match_log("*", lexemes, position)
		identdefNode.children = append(identdefNode.children, _asterixNode)
	}
	matched_log("*", lexemes, position)

	return identdefNode, nil
}

func constExpression(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	return expression(lexemes, position)
}

func assignment(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var assignmentNode = new(ParseNode)
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

	assignmentNode.children = append(assignmentNode.children, _designatorNode)
	assignmentNode.children = append(assignmentNode.children, _colonEqualOperatorNode)
	assignmentNode.children = append(assignmentNode.children, _expressionNode)

	return assignmentNode, nil
}

func procedureCall(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureCallNode = new(ParseNode)

	attempt_log("designator", lexemes, position)
	_designatorNode, err := designator(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _designatorNode == nil {
		did_not_match_log("designator", lexemes, position)
		return nil, nil
	}
	procedureCallNode.children = append(procedureCallNode.children, _designatorNode)
	matched_log("designator", lexemes, position)

	attempt_log("actualParameters", lexemes, position)
	_actualParametersNode, err := actualParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _actualParametersNode != nil {
		did_not_match_log("actualParameters", lexemes, position)
		procedureCallNode.children = append(procedureCallNode.children, _actualParametersNode)
	}
	matched_log("actualParameters", lexemes, position)

	return procedureCallNode, nil
}

func ifStatement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var ifStatementNode = new(ParseNode)
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
	ifStatementNode.children = append(ifStatementNode.children, _ifReservedWordNode)
	ifStatementNode.children = append(ifStatementNode.children, _expressionNode)
	ifStatementNode.children = append(ifStatementNode.children, _thenReservedWordNode)
	ifStatementNode.children = append(ifStatementNode.children, _statementSequenceNode)
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

		ifStatementNode.children = append(ifStatementNode.children, _elsifReservedWordNode)
		ifStatementNode.children = append(ifStatementNode.children, _expressionNode)
		ifStatementNode.children = append(ifStatementNode.children, _thenReservedWordNode)
		ifStatementNode.children = append(ifStatementNode.children, _statementSequenceNode)
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
		ifStatementNode.children = append(ifStatementNode.children, _elseReservedWordNode)
		ifStatementNode.children = append(ifStatementNode.children, _statementSequenceNode1)
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
	ifStatementNode.children = append(ifStatementNode.children, _endReservedWordNode)
	matched_log("END", lexemes, position)

	return ifStatementNode, nil
}

func label(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var labelNode = new(ParseNode)

	attempt_log("INTEGER", lexemes, position)
	_integerNode := matchType(lexemes, position, INTEGER)
	if _integerNode != nil {
		matched_log("INTEGER", lexemes, position)
		labelNode.children = append(labelNode.children, _integerNode)
		return labelNode, nil
	}
	did_not_match_log("INTEGER", lexemes, position)

	attempt_log("STRING", lexemes, position)
	_stringNode := matchType(lexemes, position, STRING)
	if _stringNode != nil {
		matched_log("STRING", lexemes, position)
		labelNode.children = append(labelNode.children, _stringNode)
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
		labelNode.children = append(labelNode.children, _qualidentNode)
		return labelNode, nil
	}
	did_not_match_log("qualident", lexemes, position)

	return nil, nil
}

func labelRange(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var labelRangeNode = new(ParseNode)
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
	labelRangeNode.children = append(labelRangeNode.children, _labelNode)
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

		labelRangeNode.children = append(labelRangeNode.children, _doubleDotOperatorNode)
		labelRangeNode.children = append(labelRangeNode.children, _labelNode)
	} else {
		did_not_match_optionally_log("..", lexemes, position)
	}

	return labelRangeNode, nil
}

func caseLabelList(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var caseLabelListNode = new(ParseNode)
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
	caseLabelListNode.children = append(caseLabelListNode.children, _labelRangeNode)
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
		caseLabelListNode.children = append(caseLabelListNode.children, _commaOperatorNode)
		caseLabelListNode.children = append(caseLabelListNode.children, _labelRangeNode)
		matched_log("labelRange", lexemes, position)
	}

	return caseLabelListNode, nil
}

func _case(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var _caseNode = new(ParseNode)

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

	_caseNode.children = append(_caseNode.children, _caseLabelListNode)
	_caseNode.children = append(_caseNode.children, _colonOperatorNode)
	_caseNode.children = append(_caseNode.children, _statementSequenceNode)

	return _caseNode, nil
}

func caseStatement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var caseStatementNode = new(ParseNode)
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

	caseStatementNode.children = append(caseStatementNode.children, _caseReservedWordNode)
	caseStatementNode.children = append(caseStatementNode.children, _expressionNode)
	caseStatementNode.children = append(caseStatementNode.children, _ofReservedWordNode)
	caseStatementNode.children = append(caseStatementNode.children, _caseNode)

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

		caseStatementNode.children = append(caseStatementNode.children, _verticalBarReservedWordNode)
		caseStatementNode.children = append(caseStatementNode.children, _caseNode)
	}

	matched_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		did_not_match_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	caseStatementNode.children = append(caseStatementNode.children, _endReservedWordNode)
	matched_log("END", lexemes, position)

	return caseStatementNode, nil

}

func repeatStatement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var repeatStatementNode = new(ParseNode)
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

	repeatStatementNode.children = append(repeatStatementNode.children, _repeatReservedWordNode)
	repeatStatementNode.children = append(repeatStatementNode.children, _statementSequenceNode)
	repeatStatementNode.children = append(repeatStatementNode.children, _untilReservedWordNode)
	repeatStatementNode.children = append(repeatStatementNode.children, _expressionNode)

	return repeatStatementNode, nil
}

func forStatement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var forStatementNode = new(ParseNode)
	var positionCheckpoint = *position

	attempt_log("FOR", lexemes, position)
	_forReservedWordNode := matchReservedWord(lexemes, position, "FOR")
	if _forReservedWordNode == nil {
		did_not_match_log("FOR", lexemes, position)
		return nil, nil
	}
	matched_log("FOR", lexemes, position)

	attempt_log("ident", lexemes, position)
	_identNode := matchType(lexemes, position, IDENT)
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

	forStatementNode.children = append(forStatementNode.children, _forReservedWordNode)
	forStatementNode.children = append(forStatementNode.children, _identNode)
	forStatementNode.children = append(forStatementNode.children, _colonEqualOperatorNode)
	forStatementNode.children = append(forStatementNode.children, _expressionNode)
	forStatementNode.children = append(forStatementNode.children, _toReservedWordNode)
	forStatementNode.children = append(forStatementNode.children, _expressionNode1)

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
		forStatementNode.children = append(forStatementNode.children, _byReservedWordNode)
		forStatementNode.children = append(forStatementNode.children, _constExpressionNode)
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

	forStatementNode.children = append(forStatementNode.children, _doReservedWordNode)
	forStatementNode.children = append(forStatementNode.children, _statementSequenceNode)
	forStatementNode.children = append(forStatementNode.children, _endReservedWordNode)

	return forStatementNode, nil
}

func whileStatement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var whileStatementNode = new(ParseNode)
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
	whileStatementNode.children = append(whileStatementNode.children, _whileReservedWordNode)
	whileStatementNode.children = append(whileStatementNode.children, _expressionNode)
	whileStatementNode.children = append(whileStatementNode.children, _doReservedWordNode)
	whileStatementNode.children = append(whileStatementNode.children, _statementSequenceNode)
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

		whileStatementNode.children = append(whileStatementNode.children, _elsifReservedWordNode)
		whileStatementNode.children = append(whileStatementNode.children, _expressionNode)
		whileStatementNode.children = append(whileStatementNode.children, _doReservedWordNode)
		whileStatementNode.children = append(whileStatementNode.children, _statementSequenceNode)
	}

	attempt_log("END", lexemes, position)
	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		attempt_log("END", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	whileStatementNode.children = append(whileStatementNode.children, _endReservedWordNode)
	attempt_log("END", lexemes, position)

	return whileStatementNode, nil
}

func statement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var statementNode = new(ParseNode)

	attempt_log("assignment", lexemes, position)
	_assignmentNode, err := assignment(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _assignmentNode != nil {
		matched_log("assignment", lexemes, position)
		statementNode.children = append(statementNode.children, _assignmentNode)
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
		statementNode.children = append(statementNode.children, _procedureCallNode)
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
		statementNode.children = append(statementNode.children, _ifStatementNode)
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
		statementNode.children = append(statementNode.children, _caseStatementNode)
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
		statementNode.children = append(statementNode.children, _whiteStatementNode)
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
		statementNode.children = append(statementNode.children, _repeatStatementNode)
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
		statementNode.children = append(statementNode.children, _forStatementNode)
		return statementNode, nil
	}
	did_not_match_log("forStatement", lexemes, position)

	return statementNode, nil
}

func statementSequence(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var statementSequenceNode = new(ParseNode)

	attempt_log("statement", lexemes, position)
	_statementNode, err := statement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementNode == nil {
		attempt_log("statement", lexemes, position)
		return nil, nil
	}
	statementSequenceNode.children = append(statementSequenceNode.children, _statementNode)
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
		statementSequenceNode.children = append(statementSequenceNode.children, _semicolonNode)
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
		statementSequenceNode.children = append(statementSequenceNode.children, _statementNode)
		matched_log("statement", lexemes, position)
	}

	return statementSequenceNode, nil
}

func procedureBody(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureBodyNode = new(ParseNode)
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
	procedureBodyNode.children = append(procedureBodyNode.children, _declarationSequenceNode)

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
		procedureBodyNode.children = append(procedureBodyNode.children, _beginReservedWordNode)
		procedureBodyNode.children = append(procedureBodyNode.children, _statementSequenceNode)
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
		procedureBodyNode.children = append(procedureBodyNode.children, _returnReservedWordNode)
		procedureBodyNode.children = append(procedureBodyNode.children, _expression)
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
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureHeadingNode = new(ParseNode)
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
	procedureHeadingNode.children = append(procedureHeadingNode.children, _procedureReservedWordNode)
	procedureHeadingNode.children = append(procedureHeadingNode.children, _identDefNode)
	matched_log("identdef", lexemes, position)

	attempt_log("formalParameters", lexemes, position)
	_formalParametersNode, err := formalParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _formalParametersNode != nil {
		did_not_match_log("formalParameters", lexemes, position)
		procedureHeadingNode.children = append(procedureHeadingNode.children, _formalParametersNode)
	}
	matched_log("formalParameters", lexemes, position)

	return procedureHeadingNode, nil
}

func procedureDeclaration(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureDeclarationNode = new(ParseNode)
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
	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	matched_log("ident", lexemes, position)

	procedureDeclarationNode.children = append(procedureDeclarationNode.children, _pocedureHeadingNode)
	procedureDeclarationNode.children = append(procedureDeclarationNode.children, _semicolonNode)
	procedureDeclarationNode.children = append(procedureDeclarationNode.children, _procedureBodyNode)
	procedureDeclarationNode.children = append(procedureDeclarationNode.children, _identNode)

	return procedureDeclarationNode, nil
}

func varDeclaration(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var varDeclarationNode = new(ParseNode)
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

	varDeclarationNode.children = append(varDeclarationNode.children, _identListNode)
	varDeclarationNode.children = append(varDeclarationNode.children, _colonNode)
	varDeclarationNode.children = append(varDeclarationNode.children, _typeNode)

	return varDeclarationNode, nil
}

func constDeclaration(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var constDeclarationNode = new(ParseNode)

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

	constDeclarationNode.children =
		append(constDeclarationNode.children, _identDefNode)
	constDeclarationNode.children =
		append(constDeclarationNode.children, _assignmentNode)
	constDeclarationNode.children =
		append(constDeclarationNode.children, _constExpressionNode)

	return constDeclarationNode, nil
}

func declarationSequence_constSequence(
	lexemes *[]Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var declarationSequence_constSequenceNode = new(ParseNode)

	// [CONST {ConstDeclaration ";"}]
	attempt_log("CONST", lexemes, position)
	_constReservedWordNode := matchReservedWord(lexemes, position, "CONST")
	if _constReservedWordNode == nil {
		did_not_match_log("CONST", lexemes, position)
		return nil, nil
	}
	matched_log("CONST", lexemes, position)
	declarationSequence_constSequenceNode.children = append(declarationSequence_constSequenceNode.children, _constReservedWordNode)

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

		declarationSequence_constSequenceNode.children =
			append(declarationSequence_constSequenceNode.children, _constDeclarationNode)
		declarationSequence_constSequenceNode.children =
			append(declarationSequence_constSequenceNode.children, _semicolonNode)
	}
	return declarationSequence_constSequenceNode, nil
}

func declarationSequence_typeDeclaration(
	lexemes *[]Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var _typeDeclarationSequenceNode = new(ParseNode)

	// [TYPE {TypeDeclaration ";"}]
	attempt_log("TYPE", lexemes, position)
	_typeReservedWordNode := matchReservedWord(lexemes, position, "TYPE")
	if _typeReservedWordNode == nil {
		did_not_match_log("TYPE", lexemes, position)
		return nil, nil
	}
	_typeDeclarationSequenceNode.children = append(_typeDeclarationSequenceNode.children, _typeReservedWordNode)
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
		_typeDeclarationSequenceNode.children =
			append(_typeDeclarationSequenceNode.children, _typeDeclarationNode)
		_typeDeclarationSequenceNode.children =
			append(_typeDeclarationSequenceNode.children, _semicolonNode)
	}

	return _typeDeclarationSequenceNode, nil
}

func declarationSequence_varDeclaration(
	lexemes *[]Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var _varDeclarationSequenceNode = new(ParseNode)
	var positionCheckpoint = *position

	attempt_log("VAR", lexemes, position)
	_varReservedWordNode := matchReservedWord(lexemes, position, "VAR")
	if _varReservedWordNode == nil {
		did_not_match_log("VAR", lexemes, position)
		return nil, nil
	}
	_varDeclarationSequenceNode.children = append(_varDeclarationSequenceNode.children, _varReservedWordNode)
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

		_varDeclarationSequenceNode.children =
			append(_varDeclarationSequenceNode.children, _varDeclarationNode)
		_varDeclarationSequenceNode.children =
			append(_varDeclarationSequenceNode.children, _semicolonNode)
	}
	return _varDeclarationSequenceNode, nil
}

func declarationSequence_procedureDeclaration(
	lexemes *[]Lexeme,
	position *int,
	declarationSequenceNode *ParseNode,
) (*ParseNode, error) {
	var _procedureDeclarationSequenceNode = new(ParseNode)

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

		_procedureDeclarationSequenceNode.children =
			append(_procedureDeclarationSequenceNode.children, _procedureDeclarationNode)
		_procedureDeclarationSequenceNode.children =
			append(_procedureDeclarationSequenceNode.children, _semicolonNode)
	}
	return _procedureDeclarationSequenceNode, nil
}

func declarationSequence(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var declarationSequenceNode = new(ParseNode)
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
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _constDeclarationSequenceNode)
	} else {
		did_not_match_optionally_log("declarationSequence_constSequence", lexemes, position)
	}

	// [TYPE {TypeDeclaration ";"}]
	attempt_optionally_log("declarationSequence_typeDeclaration", lexemes, position)
	_typeDeclarationSequenceNode, err := declarationSequence_typeDeclaration(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _typeDeclarationSequenceNode != nil {
		optionally_matched_log("declarationSequence_typeDeclaration", lexemes, position)
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _typeDeclarationSequenceNode)
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
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _varDeclarationSequenceNode)
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
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _procedureDeclarationSequenceNode)
	} else {
		did_not_match_optionally_log("declarationSequence_procedureDeclaration", lexemes, position)
	}

	return declarationSequenceNode, nil
}

func module(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var moduleNode = new(ParseNode)

	// MODULE
	attempt_log("MODULE", lexemes, position)
	_moduleNode := matchReservedWord(lexemes, position, "MODULE")
	if _moduleNode == nil {
		did_not_match_log("MODULE", lexemes, position)
		return nil, fmt.Errorf("parse error: expected 'MODULE', found %v", (*lexemes)[*position])
	}
	matched_log("MODULE", lexemes, position)
	moduleNode.children = append(moduleNode.children, _moduleNode)

	// ident
	attempt_log("ident", lexemes, position)
	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, fmt.Errorf("parse error: expected identNode, found %v", (*lexemes)[*position])
	}
	matched_log("ident", lexemes, position)
	moduleNode.children = append(moduleNode.children, _identNode)

	// ;
	attempt_log(";", lexemes, position)
	_semicolonNode := matchOperator(lexemes, position, ";")
	if _semicolonNode == nil {
		did_not_match_log(";", lexemes, position)
		return nil, fmt.Errorf("parse error: expected ';', found %v", (*lexemes)[*position])
	}
	matched_log(";", lexemes, position)
	moduleNode.children = append(moduleNode.children, _identNode)

	// [ImportList]
	attempt_optionally_log("importList", lexemes, position)
	_importListNode, err := importList(lexemes, position)
	if err != nil {
		did_not_match_optionally_log("importList", lexemes, position)
		return nil, err
	}
	if _importListNode != nil {
		optionally_matched_log("importList", lexemes, position)
		moduleNode.children = append(moduleNode.children, _importListNode)
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
	moduleNode.children = append(moduleNode.children, _declarationSequenceNode)

	// [BEGIN StatementSequence]
	attempt_optionally_log("BEGIN", lexemes, position)
	_beginNode := matchReservedWord(lexemes, position, "BEGIN")
	if _beginNode != nil {
		optionally_matched_log("BEGIN", lexemes, position)
		moduleNode.children = append(moduleNode.children, _beginNode)
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
		moduleNode.children = append(moduleNode.children, _statementSequenceNode)
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
	moduleNode.children = append(moduleNode.children, _endNode)

	attempt_log("ident", lexemes, position)
	_identNode1 := matchType(lexemes, position, IDENT)
	if _identNode1 == nil {
		did_not_match_log("ident", lexemes, position)
		return nil, parse_error("ident", lexemes, position)
	}
	matched_log("ident", lexemes, position)
	moduleNode.children = append(moduleNode.children, _identNode1)

	attempt_log(".", lexemes, position)
	_dotOperatorNode := matchOperator(lexemes, position, ".")
	if _dotOperatorNode == nil {
		did_not_match_log(".", lexemes, position)
		return nil, parse_error(".", lexemes, position)
	}
	matched_log(".", lexemes, position)
	moduleNode.children = append(moduleNode.children, _dotOperatorNode)

	return moduleNode, nil
}

func parser(lexemes *[]Lexeme, debug bool) (*ParseNode, error) {
	logging.SetBackend(backendFormatter)
	parserDebug = debug
	var position = 0
	tree, err := module(lexemes, &position)
	if err == nil {
		return nil, err
	}
	if position < len(*lexemes) {
		unparsedToken := (*lexemes)[position]
		return nil, fmt.Errorf("parse error: unparsed token: %v at (line: %d, column: %d), token number: %d", unparsedToken.label, unparsedToken.line, unparsedToken.column, position)
	}
	return tree, err
}
