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
		LOG.Debug(terminalNode)
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
		LOG.Debug(terminalNode)
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
		LOG.Debug(terminalNode)
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
	attempt_log("ident", lexemes, position)

	// ident
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
	debug("Attempting to match ident", lexemes, position)
	var _identNode = matchType(lexemes, position, IDENT)
	if _identNode == nil {
		debug("Attempting to match predefined identifier", lexemes, position)
		_identNode = matchType(lexemes, position, PREDEFINED)
	}
	if _identNode != nil {
		qualidentNode.children = append(qualidentNode.children, _identNode)
		debug("Attempting to match ';'", lexemes, position)
		_dotOperatorNode := matchOperator(lexemes, position, ".")
		if _dotOperatorNode == nil {
			return qualidentNode, nil
		}
		qualidentNode.children = append(qualidentNode.children, _dotOperatorNode)
	}

	// ident
	debug("Attempting to match ident", lexemes, position)
	var _identNode1 = matchType(lexemes, position, IDENT)
	if _identNode1 == nil {
		debug("Attempting to match predefined identifie", lexemes, position)
		_identNode1 = matchType(lexemes, position, PREDEFINED)
	}
	if _identNode1 == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	qualidentNode.children = append(qualidentNode.children, _identNode1)

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
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if nil == _expressionNode {
		return nil, nil
	}
	expListNode.children = append(expListNode.children, _expressionNode)

	// {, expList}
	for {
		_commaOperatorNode := matchOperator(lexemes, position, ",")
		if nil == _commaOperatorNode {
			break
		}
		_expressionNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _expressionNode {
			*position = positionCheckpoint
			return nil, nil
		}
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
	_dotOperatorNode := matchOperator(lexemes, position, ".")
	if _dotOperatorNode != nil {
		_identNode := matchType(lexemes, position, IDENT)
		if _identNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		selectorNode.children = append(selectorNode.children, _dotOperatorNode)
		selectorNode.children = append(selectorNode.children, _identNode)

		return selectorNode, nil
	}

	// "[" ExpList "]"
	_leftBracketNode := matchOperator(lexemes, position, "[")
	if _leftBracketNode != nil {
		_expListNode, err := expList(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _expListNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		_rightBracketNode := matchOperator(lexemes, position, "]")
		if _rightBracketNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		selectorNode.children = append(selectorNode.children, _leftBracketNode)
		selectorNode.children = append(selectorNode.children, _expListNode)
		selectorNode.children = append(selectorNode.children, _rightBracketNode)

		return selectorNode, nil
	}

	// ^
	_caratOperatorNode := matchOperator(lexemes, position, "^")
	if _caratOperatorNode != nil {
		selectorNode.children = append(selectorNode.children, _caratOperatorNode)
		return selectorNode, nil
	}

	// "(" qualident ")"
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode != nil {
		_qualidentNode, err := qualident(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _qualidentNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		_rightParenNode := matchOperator(lexemes, position, ")")
		if _rightParenNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		selectorNode.children = append(selectorNode.children, _leftParenNode)
		selectorNode.children = append(selectorNode.children, _qualidentNode)
		selectorNode.children = append(selectorNode.children, _rightParenNode)

		return selectorNode, nil
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
	debug("Attempting to match qualident", lexemes, position)
	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode == nil {
		debug("Did not match qualident", lexemes, position)
		return nil, nil
	}
	designatorNode.children = append(designatorNode.children, _qualidentNode)

	// {selector}
	for {
		debug("Attempting to match optional selector", lexemes, position)
		_selectorNode, err := selector(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _selectorNode == nil {
			debug("Did not match optional selector", lexemes, position)
			break
		}
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
	_expressionNode, err := element(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		return nil, err
	}
	elementNode.children = append(elementNode.children, _expressionNode)

	// [ .. expression ]
	_doubleDotOperator := matchOperator(lexemes, position, "..")
	if _doubleDotOperator != nil {
		_elementNode, err := element(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _elementNode == nil {
			return nil, nil
		}
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
	_leftbraceNode := matchOperator(lexemes, position, "{")
	if _leftbraceNode == nil {
		return nil, nil
	}
	setNode.children = append(setNode.children, _leftbraceNode)

	// element
	_elementNode, err := element(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _elementNode != nil {
		setNode.children = append(setNode.children, _elementNode)
		// {, element}
		for {
			_commaNode := matchOperator(lexemes, position, ",")
			if _commaNode == nil {
				break
			}
			_elementNode, err := element(lexemes, position)
			if err != nil {
				return nil, err
			}
			if _elementNode == nil {
				*position = positionCheckpoint
				return nil, nil
			}
			setNode.children = append(setNode.children, _commaNode)
			setNode.children = append(setNode.children, _elementNode)
		}
	}

	// }
	_rightBraceNode := matchOperator(lexemes, position, "}")
	if _rightBraceNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
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
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode == nil {
		return nil, nil
	}
	actualParametersNode.children = append(actualParametersNode.children, _leftParenNode)

	// [expList]
	_expListNode, err := expList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expListNode != nil {
		actualParametersNode.children = append(actualParametersNode.children, _expListNode)
	}

	// )
	_rightParenNode := matchOperator(lexemes, position, ")")
	if _rightParenNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
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

		debug("Attempting to match integer", lexemes, position)
		_integerNode := matchType(lexemes, position, INTEGER)
		if _integerNode != nil {
			factorNode.children = append(factorNode.children, _integerNode)
			debug("Matched integer", lexemes, position)
			return factorNode, nil
		}

		debug("Attempting to match real", lexemes, position)
		_realNode := matchType(lexemes, position, REAL)
		if _realNode != nil {
			debug("Matched real", lexemes, position)
			factorNode.children = append(factorNode.children, _realNode)
			return factorNode, nil
		}
	}

	// string
	debug("Attempting to match string", lexemes, position)
	_stringNode := matchType(lexemes, position, STRING)
	if _stringNode != nil {
		debug("Matched string", lexemes, position)
		factorNode.children = append(factorNode.children, _stringNode)
		return factorNode, nil
	}

	// NIL
	debug("Attempting to match NIL", lexemes, position)
	_nilNode := matchReservedWord(lexemes, position, "NIL")
	if _nilNode != nil {
		debug("Matched NIL", lexemes, position)
		factorNode.children = append(factorNode.children, _nilNode)
		return factorNode, nil
	}

	// TRUE
	debug("Attempting to match TRUE", lexemes, position)
	_trueNode := matchReservedWord(lexemes, position, "TRUE")
	if _trueNode != nil {
		debug("Matched TRUE", lexemes, position)
		factorNode.children = append(factorNode.children, _trueNode)
		return factorNode, nil
	}

	// FALSE
	debug("Attempting to match FALSE", lexemes, position)
	_falseNode := matchReservedWord(lexemes, position, "FALSE")
	if _falseNode != nil {
		debug("Matched FALSE", lexemes, position)
		factorNode.children = append(factorNode.children, _falseNode)
		return factorNode, nil
	}

	// set
	debug("Attempting to match set", lexemes, position)
	_setNode, err := set(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _setNode != nil {
		debug("Matched set", lexemes, position)
		factorNode.children = append(factorNode.children, _setNode)
		return factorNode, nil
	}

	// designator [ActualParameters]
	debug("Attempting to match designator", lexemes, position)
	_designatorNode, err := designator(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _designatorNode != nil {
		debug("Matched designator", lexemes, position)
		factorNode.children = append(factorNode.children, _designatorNode)
		debug("Attempting to match optional actualParameters", lexemes, position)
		_actualParametersNode, err := actualParameters(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil != _actualParametersNode {
			debug("Matched optional actualParameters", lexemes, position)
			factorNode.children = append(factorNode.children, _actualParametersNode)
		} else {
			debug("Did not match optional actualParameters", lexemes, position)
		}
		return factorNode, nil
	}

	// "(" expression ")"
	_leftParenNode := matchOperator(lexemes, position, "(")
	if _leftParenNode == nil {
		return nil, nil
	}
	debug("Attempting to match expression", lexemes, position)
	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if nil == _expressionNode {
		*position = positionCheckpoint
		return nil, nil
	}
	_rightParenNode := matchOperator(lexemes, position, ")")
	if err != nil {
		return nil, err
	}
	if _rightParenNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	factorNode.children = append(factorNode.children, _leftParenNode)
	factorNode.children = append(factorNode.children, _expressionNode)
	factorNode.children = append(factorNode.children, _rightParenNode)
	return factorNode, nil
}

func mulOperator(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var mulOperatorNode = new(ParseNode)

	_asterixOperatorNode := matchOperator(lexemes, position, "*")
	if _asterixOperatorNode != nil {
		mulOperatorNode.children = append(mulOperatorNode.children, _asterixOperatorNode)
		return mulOperatorNode, nil
	}

	_divOperatorNode := matchOperator(lexemes, position, "/")
	if _divOperatorNode != nil {
		mulOperatorNode.children = append(mulOperatorNode.children, _divOperatorNode)
		return mulOperatorNode, nil
	}

	_divNode := matchOperator(lexemes, position, "DIV")
	if _divNode != nil {
		mulOperatorNode.children = append(mulOperatorNode.children, _divNode)
		return mulOperatorNode, nil
	}

	_modeOperatorNode := matchOperator(lexemes, position, "MOD")
	if _modeOperatorNode != nil {
		mulOperatorNode.children = append(mulOperatorNode.children, _modeOperatorNode)
		return mulOperatorNode, nil
	}

	_ampersandOperatorNode := matchOperator(lexemes, position, "&")
	if _ampersandOperatorNode != nil {
		mulOperatorNode.children = append(mulOperatorNode.children, _ampersandOperatorNode)
		return mulOperatorNode, nil
	}

	return nil, nil
}

func term(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var termNode = new(ParseNode)
	var positionCheckpoint = *position

	debug("Attempting to match factor", lexemes, position)
	_factorNode, err := factor(lexemes, position)
	if err != nil {
		debug("Error matching factor", lexemes, position)
		return nil, err
	}
	if _factorNode == nil {
		debug("Did not match factor", lexemes, position)
		return nil, nil
	}
	debug("Matched factor", lexemes, position)
	termNode.children = append(termNode.children, _factorNode)
	for {
		_mulOperatorNode, err := mulOperator(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _mulOperatorNode {
			break
		}
		_factorNode, err := factor(lexemes, position)
		if err != nil {
			*position = positionCheckpoint
			return nil, err
		}
		if nil == _factorNode {
			*position = positionCheckpoint
			return nil, nil
		}
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

	_plusOperatorNode := matchOperator(lexemes, position, "+")
	if _plusOperatorNode != nil {
		addOperatorNode.children = append(addOperatorNode.children, _plusOperatorNode)
		return addOperatorNode, nil
	}

	_minusOperatorNode := matchOperator(lexemes, position, "-")
	if _minusOperatorNode != nil {
		addOperatorNode.children = append(addOperatorNode.children, _minusOperatorNode)
		return addOperatorNode, nil
	}

	_orOperatorNode := matchOperator(lexemes, position, "&")
	if _orOperatorNode != nil {
		addOperatorNode.children = append(addOperatorNode.children, _orOperatorNode)
		return addOperatorNode, nil
	}

	return nil, nil

}

func simpleExpression(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var simpleExpressionNode = new(ParseNode)
	var positionCheckpoint = *position
	_plusOperatorNode := matchOperator(lexemes, position, "+")
	if nil == _plusOperatorNode {
		_minusOperatorNode := matchOperator(lexemes, position, "+")
		if nil != _minusOperatorNode {
			simpleExpressionNode.children = append(simpleExpressionNode.children, _minusOperatorNode)
		}
	} else {
		simpleExpressionNode.children = append(simpleExpressionNode.children, _plusOperatorNode)
	}

	debug("Attempting to match term", lexemes, position)
	_termNode, err := term(lexemes, position)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if nil == _termNode {
		*position = positionCheckpoint
		return nil, nil
	}
	debug("Matched term", lexemes, position)

	simpleExpressionNode.children = append(simpleExpressionNode.children, _termNode)
	for {
		_addOperatorNode, err := addOperator(lexemes, position)
		if err != nil {
			return nil, err
		}
		if nil == _addOperatorNode {
			break
		}

		_termNode, err := term(lexemes, position)
		if err != nil {
			*position = positionCheckpoint
			return nil, err
		}
		if nil == _termNode {
			*position = positionCheckpoint
			return nil, nil
		}
		simpleExpressionNode.children = append(simpleExpressionNode.children, _addOperatorNode)
		simpleExpressionNode.children = append(simpleExpressionNode.children, _termNode)
		positionCheckpoint = *position
	}
	return simpleExpressionNode, nil
}

func relation(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var relationOperatorNode = new(ParseNode)

	_equalOperatorNode := matchOperator(lexemes, position, "=")
	if _equalOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _equalOperatorNode)
		return relationOperatorNode, nil
	}

	_hashOperatorNode := matchOperator(lexemes, position, "#")
	if _hashOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _hashOperatorNode)
		return relationOperatorNode, nil
	}

	_lessThanOperatorNode := matchOperator(lexemes, position, "<")
	if _lessThanOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _lessThanOperatorNode)
		return relationOperatorNode, nil
	}

	_lessThanEqualOperatorNode := matchOperator(lexemes, position, "<=")
	if _lessThanEqualOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _lessThanEqualOperatorNode)
		return relationOperatorNode, nil
	}

	_greaterThanOperatorNode := matchOperator(lexemes, position, ">")
	if _greaterThanOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _greaterThanOperatorNode)
		return relationOperatorNode, nil
	}

	_greaterThanEqualOperatorNode := matchOperator(lexemes, position, ">=")
	if _greaterThanEqualOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _greaterThanEqualOperatorNode)
		return relationOperatorNode, nil
	}

	_inOperatorNode := matchOperator(lexemes, position, "IN")
	if _inOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _inOperatorNode)
		return relationOperatorNode, nil
	}

	_isOperatorNode := matchOperator(lexemes, position, "IS")
	if _isOperatorNode != nil {
		relationOperatorNode.children = append(relationOperatorNode.children, _isOperatorNode)
		return relationOperatorNode, nil
	}

	return nil, nil

}

func expression(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var expressionNode = new(ParseNode)
	var positionCheckpoint = *position

	debug("Attempting to match simpleExpression", lexemes, position)
	_simpleExpressionNode, err := simpleExpression(lexemes, position)
	if err != nil {
		return nil, err
	}
	expressionNode.children = append(expressionNode.children, _simpleExpressionNode)
	_relationNode, err := relation(lexemes, position)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _relationNode != nil {
		_simpleExpressionNode, err := simpleExpression(lexemes, position)
		if err != nil {
			*position = positionCheckpoint
			return nil, err
		}
		if _simpleExpressionNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
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

	debug("Attempting to match qualident", lexemes, position)
	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode != nil {
		debug("Matched qualident", lexemes, position)
		_typeNode.children = append(_typeNode.children, _qualidentNode)
		return _typeNode, nil
	}
	debug("Did not match qualident", lexemes, position)

	_strucTypeNode, err := strucType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _strucTypeNode != nil {
		_typeNode.children = append(_typeNode.children, _strucTypeNode)
		return _typeNode, nil
	}

	return nil, nil
}

func arrayType(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var arrayTypeNode = new(ParseNode)
	var positionCheckpoint = *position
	var positionCheckpoint1 = *position

	debug("Attempting to match ARRAY", lexemes, position)
	_arrayReservedWord := matchReservedWord(lexemes, position, "ARRAY")
	if _arrayReservedWord == nil {
		return nil, nil
	}
	debug("Matched ARRAY", lexemes, position)

	debug("Attempting to match length", lexemes, position)
	_lengthNode, err := length(lexemes, position)
	if err != nil {
		debug("Error matching length", lexemes, position)
		return nil, err
	}
	if _lengthNode == nil {
		debug("Did not match length", lexemes, position)
		return nil, nil
	}
	debug("Matched length", lexemes, position)
	arrayTypeNode.children = append(arrayTypeNode.children, _arrayReservedWord)
	arrayTypeNode.children = append(arrayTypeNode.children, _lengthNode)
	for {
		debug("Attempting to match optional ','", lexemes, position)
		_commaOperatorNode := matchOperator(lexemes, position, ",")
		if err != nil {
			return nil, err
		}
		if _commaOperatorNode == nil {
			debug("Did not match optional ','", lexemes, position)
			break
		}
		debug("Matched optional ','", lexemes, position)
		debug("Attempting to match length", lexemes, position)
		_lengthNode, err := length(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _lengthNode == nil {
			debug("Did not match length", lexemes, position)
			*position = positionCheckpoint
			return nil, nil
		}
		debug("Matched length", lexemes, position)
		arrayTypeNode.children = append(arrayTypeNode.children, _commaOperatorNode)
		arrayTypeNode.children = append(arrayTypeNode.children, _lengthNode)
	}
	debug("Attempting to match OF", lexemes, position)
	_ofReservedWordNode := matchReservedWord(lexemes, position, "OF")
	if _ofReservedWordNode == nil {
		debug("Did not match OF", lexemes, position)
		*position = positionCheckpoint1
		return nil, nil
	}
	debug("Matched OF", lexemes, position)

	debug("Attempting to match type", lexemes, position)
	_typeNode, err := _type(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _typeNode == nil {
		debug("Did not match type", lexemes, position)
		*position = positionCheckpoint1
		return nil, nil
	}
	debug("Matched type", lexemes, position)
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

	_identListNode, err := identList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identListNode == nil {
		return nil, nil
	}

	_colonNode := matchOperator(lexemes, position, ":")
	if _colonNode == nil {
		*position = positionCheckpoint
	}

	_typeNode, err := _type(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _typeNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
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

	_fieldListNode, err := fieldList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _fieldListNode == nil {
		return nil, nil
	}
	fieldListSequenceNode.children = append(fieldListSequenceNode.children, _fieldListNode)
	for {
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			break
		}
		fieldListSequenceNode.children = append(fieldListSequenceNode.children, _semicolonNode)

		_fieldListNode, err := fieldList(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _fieldListNode == nil {
			break
		}
		fieldListSequenceNode.children = append(fieldListSequenceNode.children, _fieldListNode)
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

	debug("Attempting to match POINTER", lexemes, position)
	_pointerTypeNode := matchReservedWord(lexemes, position, "POINTER")
	if _pointerTypeNode == nil {
		debug("Did not match POINTER", lexemes, position)
		return nil, nil
	}
	debug("Matched POINTER", lexemes, position)

	debug("Attempting to match TO", lexemes, position)
	_toNode := matchReservedWord(lexemes, position, "TO")
	if _toNode == nil {
		debug("Did not match TO", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	debug("Matched TO", lexemes, position)

	_typeNode, err := _type(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _typeNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
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

	_arrayReservedNode := matchReservedWord(lexemes, position, "ARRAY")
	if _arrayReservedNode != nil {
		_ofReservedNode := matchReservedWord(lexemes, position, "OF")
		if _ofReservedNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		formalTypeNode.children = append(formalTypeNode.children, _arrayReservedNode)
		formalTypeNode.children = append(formalTypeNode.children, _ofReservedNode)
	}

	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	formalTypeNode.children = append(formalTypeNode.children, _qualidentNode)

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

	debug("Attempting to match arrayType", lexemes, position)
	_arrayTypeNode, err := arrayType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _arrayTypeNode != nil {
		debug("Matched arrayType", lexemes, position)
		strucTypeNode.children = append(strucTypeNode.children, _arrayTypeNode)
		return strucTypeNode, nil
	}

	_recordTypeNode, err := recordType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _recordTypeNode != nil {
		strucTypeNode.children = append(strucTypeNode.children, _recordTypeNode)
		return strucTypeNode, nil
	}

	_pointerTypeNode, err := pointerType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _pointerTypeNode != nil {
		strucTypeNode.children = append(strucTypeNode.children, _pointerTypeNode)
		return strucTypeNode, nil
	}

	_procedureTypeNode, err := procedureType(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _procedureTypeNode != nil {
		strucTypeNode.children = append(strucTypeNode.children, _procedureTypeNode)
		return strucTypeNode, nil
	}

	return nil, nil
}

func typeDeclaration(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var typeDeclarationNode = new(ParseNode)
	var positionCheckpoint = *position

	debug("Attempting to match identdef", lexemes, position)
	_identDefNode, err := identdef(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identDefNode == nil {
		debug("Could not find identdef", lexemes, position)
		return nil, nil
	}

	debug("Attempting to match '='", lexemes, position)
	_equalOperatorNode := matchOperator(lexemes, position, "=")
	if _equalOperatorNode == nil {
		debug("Did not match '='", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}

	debug("Attempting to match strucType", lexemes, position)
	_strucTypeNode, err := strucType(lexemes, position)
	if err != nil {
		debug(fmt.Sprintf("Error matching strucType %s", err.Error()), lexemes, position)
		return nil, err
	}
	if _strucTypeNode == nil {
		debug("Did not match strucType", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	debug("Matched strucType", lexemes, position)

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
	debug("Attempting to match ident", lexemes, position)
	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		debug("Did not match ident", lexemes, position)
		return nil, nil
	}
	debug("Matched ident", lexemes, position)
	identdefNode.children =
		append(identdefNode.children, _identNode)
	debug("Attempting to match optional '*'", lexemes, position)
	_asterixNode := matchOperator(lexemes, position, "*")
	if _asterixNode != nil {
		debug("Matched optional '*'", lexemes, position)
		identdefNode.children =
			append(identdefNode.children, _asterixNode)
	}

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

	_designatorNode, err := designator(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _designatorNode == nil {
		return nil, nil
	}

	_colonEqualOperatorNode := matchOperator(lexemes, position, ":=")
	if _colonEqualOperatorNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
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

	_designatorNode, err := designator(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _designatorNode == nil {
		return nil, nil
	}
	procedureCallNode.children = append(procedureCallNode.children, _designatorNode)

	_actualParametersNode, err := actualParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _actualParametersNode != nil {
		procedureCallNode.children = append(procedureCallNode.children, _actualParametersNode)
	}

	return procedureCallNode, nil
}

func ifStatement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var ifStatementNode = new(ParseNode)
	var positionCheckpoint = *position

	_ifReservedWordNode := matchReservedWord(lexemes, position, "IF")
	if _ifReservedWordNode == nil {
		return nil, nil
	}

	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_thenReservedWordNode := matchReservedWord(lexemes, position, "THEN")
	if _thenReservedWordNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	ifStatementNode.children = append(ifStatementNode.children, _ifReservedWordNode)
	ifStatementNode.children = append(ifStatementNode.children, _expressionNode)
	ifStatementNode.children = append(ifStatementNode.children, _thenReservedWordNode)
	ifStatementNode.children = append(ifStatementNode.children, _statementSequenceNode)

	for {
		_elsifReservedWordNode := matchReservedWord(lexemes, position, "ELSIF")
		if _elsifReservedWordNode == nil {
			break
		}

		_expressionNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _expressionNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}

		_thenReservedWordNode := matchReservedWord(lexemes, position, "THEN")
		if _thenReservedWordNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}

		_statementSequenceNode, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}

		ifStatementNode.children = append(ifStatementNode.children, _elsifReservedWordNode)
		ifStatementNode.children = append(ifStatementNode.children, _expressionNode)
		ifStatementNode.children = append(ifStatementNode.children, _thenReservedWordNode)
		ifStatementNode.children = append(ifStatementNode.children, _statementSequenceNode)
	}

	_elseReservedWordNode := matchReservedWord(lexemes, position, "ELSE")
	if _elseReservedWordNode != nil {
		_statementSequenceNode1, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode1 == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		ifStatementNode.children = append(ifStatementNode.children, _elseReservedWordNode)
		ifStatementNode.children = append(ifStatementNode.children, _statementSequenceNode1)
	}

	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	ifStatementNode.children = append(ifStatementNode.children, _endReservedWordNode)

	return ifStatementNode, nil
}

func label(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var labelNode = new(ParseNode)

	_integerNode := matchType(lexemes, position, INTEGER)
	if _integerNode != nil {
		labelNode.children = append(labelNode.children, _integerNode)
		return labelNode, nil
	}

	_stringNode := matchType(lexemes, position, STRING)
	if _stringNode != nil {
		labelNode.children = append(labelNode.children, _stringNode)
		return labelNode, nil
	}

	_qualidentNode, err := qualident(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _qualidentNode != nil {
		labelNode.children = append(labelNode.children, _qualidentNode)
		return labelNode, nil
	}

	return nil, nil
}

func labelRange(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var labelRangeNode = new(ParseNode)
	var positionCheckpoint = *position

	_labelNode, err := label(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _labelNode == nil {
		return nil, nil
	}
	labelRangeNode.children = append(labelRangeNode.children, _labelNode)

	_doubleDotOperatorNode := matchOperator(lexemes, position, "..")
	if _doubleDotOperatorNode != nil {
		_labelNode, err := label(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _labelNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		labelRangeNode.children = append(labelRangeNode.children, _doubleDotOperatorNode)
		labelRangeNode.children = append(labelRangeNode.children, _labelNode)
	}

	return labelRangeNode, nil
}

func caseLabelList(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var caseLabelListNode = new(ParseNode)
	var positionCheckpoint = *position

	_labelRangeNode, err := labelRange(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _labelRangeNode == nil {
		return nil, nil
	}
	caseLabelListNode.children = append(caseLabelListNode.children, _labelRangeNode)

	for {
		_commaOperatorNode := matchOperator(lexemes, position, ",")
		if _commaOperatorNode == nil {
			break
		}

		_labelRangeNode, err := labelRange(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _labelRangeNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		caseLabelListNode.children = append(caseLabelListNode.children, _commaOperatorNode)
		caseLabelListNode.children = append(caseLabelListNode.children, _labelRangeNode)
	}

	return caseLabelListNode, nil
}

func _case(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var _caseNode = new(ParseNode)

	_caseLabelListNode, err := caseLabelList(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _caseLabelListNode == nil {
		return _caseNode, nil
	}

	_colonOperatorNode := matchOperator(lexemes, position, ":")
	if _colonOperatorNode == nil {
		return nil, nil
	}

	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		return nil, nil
	}

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

	_caseReservedWordNode := matchReservedWord(lexemes, position, "CASE")
	if _caseReservedWordNode == nil {
		return nil, nil
	}

	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_ofReservedWordNode := matchReservedWord(lexemes, position, "OF")
	if _ofReservedWordNode == nil {
		return nil, nil
	}

	_caseNode, err := _case(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _caseNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	caseStatementNode.children = append(caseStatementNode.children, _caseReservedWordNode)
	caseStatementNode.children = append(caseStatementNode.children, _expressionNode)
	caseStatementNode.children = append(caseStatementNode.children, _ofReservedWordNode)
	caseStatementNode.children = append(caseStatementNode.children, _caseNode)

	for {
		_verticalBarReservedWordNode := matchOperator(lexemes, position, "|")
		if _verticalBarReservedWordNode == nil {
			break
		}

		_caseNode, err := _case(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _caseNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}

		caseStatementNode.children = append(caseStatementNode.children, _verticalBarReservedWordNode)
		caseStatementNode.children = append(caseStatementNode.children, _caseNode)
	}

	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	caseStatementNode.children = append(caseStatementNode.children, _endReservedWordNode)

	return caseStatementNode, nil

}

func repeatStatement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var repeatStatementNode = new(ParseNode)
	var positionCheckpoint = *position

	_repeatReservedWordNode := matchReservedWord(lexemes, position, "REPEAT")
	if _repeatReservedWordNode == nil {
		return nil, nil
	}

	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_untilReservedWordNode := matchReservedWord(lexemes, position, "UNTIL")
	if _untilReservedWordNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
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

	_forReservedWordNode := matchReservedWord(lexemes, position, "FOR")
	if _forReservedWordNode == nil {
		return nil, nil
	}

	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_colonEqualOperatorNode := matchOperator(lexemes, position, ":=")
	if _colonEqualOperatorNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_toReservedWordNode := matchReservedWord(lexemes, position, "TO")
	if _toReservedWordNode == nil {
		return nil, nil
	}

	_expressionNode1, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode1 == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	forStatementNode.children = append(forStatementNode.children, _forReservedWordNode)
	forStatementNode.children = append(forStatementNode.children, _identNode)
	forStatementNode.children = append(forStatementNode.children, _colonEqualOperatorNode)
	forStatementNode.children = append(forStatementNode.children, _expressionNode)
	forStatementNode.children = append(forStatementNode.children, _toReservedWordNode)
	forStatementNode.children = append(forStatementNode.children, _expressionNode1)

	_byReservedWordNode := matchReservedWord(lexemes, position, "BY")
	if _byReservedWordNode != nil {
		_constExpressionNode, err := constExpression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _constExpressionNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}
		forStatementNode.children = append(forStatementNode.children, _byReservedWordNode)
		forStatementNode.children = append(forStatementNode.children, _constExpressionNode)
	}

	_doReservedWordNode := matchReservedWord(lexemes, position, "DO")
	if _doReservedWordNode == nil {
		return nil, nil
	}

	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}

	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		return nil, nil
	}

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

	_whileReservedWordNode := matchReservedWord(lexemes, position, "WHILE")
	if _whileReservedWordNode == nil {
		return nil, nil
	}

	_expressionNode, err := expression(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _expressionNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_doReservedWordNode := matchReservedWord(lexemes, position, "DO")
	if _doReservedWordNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

	_statementSequenceNode, err := statementSequence(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementSequenceNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	whileStatementNode.children = append(whileStatementNode.children, _whileReservedWordNode)
	whileStatementNode.children = append(whileStatementNode.children, _expressionNode)
	whileStatementNode.children = append(whileStatementNode.children, _doReservedWordNode)
	whileStatementNode.children = append(whileStatementNode.children, _statementSequenceNode)

	for {
		_elsifReservedWordNode := matchReservedWord(lexemes, position, "ELSIF")
		if _elsifReservedWordNode == nil {
			break
		}

		_expressionNode, err := expression(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _expressionNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}

		_doReservedWordNode := matchReservedWord(lexemes, position, "DO")
		if _doReservedWordNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}

		_statementSequenceNode, err := statementSequence(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementSequenceNode == nil {
			*position = positionCheckpoint
			return nil, nil
		}

		whileStatementNode.children = append(whileStatementNode.children, _elsifReservedWordNode)
		whileStatementNode.children = append(whileStatementNode.children, _expressionNode)
		whileStatementNode.children = append(whileStatementNode.children, _doReservedWordNode)
		whileStatementNode.children = append(whileStatementNode.children, _statementSequenceNode)
	}

	_endReservedWordNode := matchReservedWord(lexemes, position, "END")
	if _endReservedWordNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}
	whileStatementNode.children = append(whileStatementNode.children, _endReservedWordNode)

	return whileStatementNode, nil
}

func statement(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var statementNode = new(ParseNode)

	_assignmentNode, err := assignment(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _assignmentNode != nil {
		statementNode.children = append(statementNode.children, _assignmentNode)
		return statementNode, nil
	}

	_procedureCallNode, err := procedureCall(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _procedureCallNode != nil {
		statementNode.children = append(statementNode.children, _procedureCallNode)
		return statementNode, nil
	}

	_ifStatementNode, err := ifStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _ifStatementNode != nil {
		statementNode.children = append(statementNode.children, _ifStatementNode)
		return statementNode, nil
	}

	_caseStatementNode, err := caseStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _caseStatementNode != nil {
		statementNode.children = append(statementNode.children, _caseStatementNode)
		return statementNode, nil
	}

	_whiteStatementNode, err := whileStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _whiteStatementNode != nil {
		statementNode.children = append(statementNode.children, _whiteStatementNode)
		return statementNode, nil
	}

	_repeatStatementNode, err := repeatStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _repeatStatementNode != nil {
		statementNode.children = append(statementNode.children, _repeatStatementNode)
		return statementNode, nil
	}

	_forStatementNode, err := forStatement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _forStatementNode != nil {
		statementNode.children = append(statementNode.children, _forStatementNode)
		return statementNode, nil
	}

	return nil, nil
}

func statementSequence(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var statementSequenceNode = new(ParseNode)

	_statementNode, err := statement(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _statementNode == nil {
		return nil, nil
	}
	statementSequenceNode.children = append(statementSequenceNode.children, _statementNode)

	for {
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			break
		}
		_statementNode, err := statement(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _statementNode == nil {
			return nil, nil
		}
		statementSequenceNode.children = append(statementSequenceNode.children, _semicolonNode)
		statementSequenceNode.children = append(statementSequenceNode.children, _statementNode)
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

	debug("Attempting to match PROCEDURE reserved word", lexemes, position)
	_procedureReservedWordNode := matchReservedWord(lexemes, position, "PROCEDURE")
	if _procedureReservedWordNode == nil {
		debug("Did not match PROCEDURE reserved word", lexemes, position)
		return nil, nil
	}
	debug("Matched PROCEDURE reserved word", lexemes, position)

	debug("Attempting to match identdef", lexemes, position)
	_identDefNode, err := identdef(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _identDefNode == nil {
		debug("Did not match identdef", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	debug("Matched identdef", lexemes, position)

	procedureHeadingNode.children = append(procedureHeadingNode.children, _procedureReservedWordNode)
	procedureHeadingNode.children = append(procedureHeadingNode.children, _identDefNode)

	_formalParametersNode, err := formalParameters(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _formalParametersNode != nil {
		procedureHeadingNode.children = append(procedureHeadingNode.children, _formalParametersNode)
	}

	return procedureHeadingNode, nil
}

func procedureDeclaration(
	lexemes *[]Lexeme,
	position *int,
) (*ParseNode, error) {
	var procedureDeclarationNode = new(ParseNode)
	var positionCheckpoint = *position

	debug("Attempting to match procedureHeading", lexemes, position)
	_pocedureHeadingNode, err := procedureHeading(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _pocedureHeadingNode == nil {
		debug("Did not find procedureHeading", lexemes, position)
		return nil, nil
	}
	debug("Matched procedureHeading", lexemes, position)

	debug("Attempting to match procedureHeading", lexemes, position)
	_semicolonNode := matchOperator(lexemes, position, ";")
	if _semicolonNode == nil {
		debug("Did not find ;", lexemes, position)
		*position = positionCheckpoint
		return nil, nil
	}
	debug("Matched ;", lexemes, position)

	debug("Attempting to match procedureBody", lexemes, position)
	_procedureBodyNode, err := procedureBody(lexemes, position)
	if err != nil {
		return nil, err
	}
	if _procedureBodyNode == nil {
		*position = positionCheckpoint
		debug("Did not find procedureBody", lexemes, position)
		return nil, nil
	}
	debug("Matched procedureBody", lexemes, position)

	_identNode := matchType(lexemes, position, IDENT)
	if _identNode == nil {
		*position = positionCheckpoint
		return nil, nil
	}

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
	_typeReservedWordNode := matchReservedWord(lexemes, position, "TYPE")

	if _typeReservedWordNode == nil {
		return nil, nil
	}
	_typeDeclarationSequenceNode.children =
		append(_typeDeclarationSequenceNode.children, _typeReservedWordNode)
	for {
		_typeDeclarationNode, err := typeDeclaration(lexemes, position)
		if err != nil {
			return _typeDeclarationSequenceNode, err
		}
		if _typeDeclarationNode == nil {
			break
		}
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			return nil, nil
		}
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
		_procedureDeclarationNode, err := procedureDeclaration(lexemes, position)
		if err != nil {
			return nil, err
		}
		if _procedureDeclarationNode == nil {
			break
		}
		_semicolonNode := matchOperator(lexemes, position, ";")
		if _semicolonNode == nil {
			return nil, nil
		}
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
	attempt_log("constSequence", lexemes, position)
	_constDeclarationSequenceNode, err := declarationSequence_constSequence(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _constDeclarationSequenceNode != nil {
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _constDeclarationSequenceNode)
	}

	// [TYPE {TypeDeclaration ";"}]
	_typeDeclarationSequenceNode, err := declarationSequence_typeDeclaration(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _typeDeclarationSequenceNode != nil {
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _typeDeclarationSequenceNode)
	}

	// [VAR {VarDeclaration ";"}]
	_varDeclarationSequenceNode, err := declarationSequence_varDeclaration(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _varDeclarationSequenceNode != nil {
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _varDeclarationSequenceNode)
	}

	// {ProcedureDeclaration ";"}
	_procedureDeclarationSequenceNode, err := declarationSequence_procedureDeclaration(lexemes, position, declarationSequenceNode)
	if err != nil {
		*position = positionCheckpoint
		return nil, err
	}
	if _procedureDeclarationSequenceNode != nil {
		declarationSequenceNode.children =
			append(declarationSequenceNode.children, _procedureDeclarationSequenceNode)
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

func parser(lexemes *[]Lexeme) (*ParseNode, error) {
	logging.SetBackend(backendFormatter)
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
