package main

import (
	"fmt"
)

type LexemeType int

const (
	CHAR          int = 1
	NUMBER            = 2
	STRING            = 3
	COMMENT           = 4
	KEYWORD           = 5
	PREDEFINED        = 6
	IDENT             = 7
	OP_OR_DELIM       = 8
	RESERVED_WORD     = 9
)

var OPERATORS = map[string]bool{
	"+":  true,
	"-":  true,
	"*":  true,
	"/":  true,
	"~":  true,
	"&":  true,
	".":  true,
	",":  true,
	";":  true,
	"|":  true,
	"(":  true,
	"[":  true,
	"{":  true,
	":=": true,
	"^":  true,
	"=":  true,
	"#":  true,
	"<":  true,
	">":  true,
	">=": true,
	"..": true,
	":":  true,
	")":  true,
	"]":  true,
	"}":  true,
}

var RESERVED_WORDS = map[string]bool{
	"ARRAY":     true,
	"BEGIN":     true,
	"BY":        true,
	"CASE":      true,
	"CONST":     true,
	"DIV":       true,
	"DO":        true,
	"ELSE":      true,
	"ELSIF":     true,
	"END":       true,
	"FALSE":     true,
	"FOR":       true,
	"IF":        true,
	"IMPORT":    true,
	"IN":        true,
	"IS":        true,
	"MOD":       true,
	"MODULE":    true,
	"NIL":       true,
	"OF":        true,
	"OR":        true,
	"POINTER":   true,
	"PROCEDURE": true,
	"RECORD":    true,
	"REPEAT":    true,
	"RETURN":    true,
	"THEN":      true,
	"TO":        true,
	"TRUE":      true,
	"TYPE":      true,
	"UNTIL":     true,
	"VAR":       true,
	"WHILE":     true,
}

var PREDEFINED_IDENTIFIERS = map[string]bool{
	"ABS":      true,
	"ASH":      true,
	"BOOLEAN":  true,
	"CAP":      true,
	"CHAR":     true,
	"CHR":      true,
	"COPY":     true,
	"DEC":      true,
	"ENTIER":   true,
	"EXCL":     true,
	"FALSE":    true,
	"HALT":     true,
	"INC":      true,
	"INCL":     true,
	"INTEGER":  true,
	"LEN":      true,
	"LONG":     true,
	"LONGINT":  true,
	"LONGREAL": true,
	"MAX":      true,
	"MIN":      true,
	"NEW":      true,
	"ODD":      true,
	"ORD":      true,
	"REAL":     true,
	"SET":      true,
	"SHORT":    true,
	"SHORTINT": true,
	"SIZE":     true,
	"TRUE":     true,
}

type Lexeme struct {
	label string
	typ   LexemeType
}

type LexerResult struct {
	lexemes *[]Lexeme
	err     error
}

func isDigit(b byte) bool {
	return b >= 48 && b <= 57
}

func isHexDigit(b byte) bool {
	return b >= 65 && b <= 70
}

func isLetter(b byte) bool {
	return (b >= 65 && b <= 90) || (b >= 97 && b <= 122)
}

func isAlphaNumeric(b byte) bool {
	return isDigit(b) || isLetter(b)
}

func isUnicode(b byte) bool {
	return b > 127
}

func isWhitespace(b byte) bool {
	return (b >= 9 && b <= 13) || b == 32
}

func isReservedWord(lexeme string) bool {
	return RESERVED_WORDS[lexeme]
}

func isPredefinedIdentifier(lexeme string) bool {
	return PREDEFINED_IDENTIFIERS[lexeme]
}

func isOperator(lexeme string) bool {
	return OPERATORS[lexeme]
}

func isString(lexeme string) bool {
	if len(lexeme) < 2 {
		return false
	}
	if lexeme[0] == '"' && lexeme[len(lexeme)-1] == '"' {
		for i := 1; i < len(lexeme)-1; i++ {
			if !isAlphaNumeric(lexeme[i]) {
				return false
			}
		}
		return true
	}
	for i := 0; i < len(lexeme)-1; i++ {
		if !isDigit(lexeme[i]) {
			return false
		}
	}
	return lexeme[len(lexeme)-1] == 'X'
}

func isInteger(lexeme string) bool {
	var foundHex = false
	if len(lexeme) == 0 || (!isDigit(byte(lexeme[0])) && !isHexDigit(byte(lexeme[0]))) {
		return false
	}
	for i := 1; i < len(lexeme)-1; i++ {
		if isDigit(byte(lexeme[i])) {
			continue
		} else if isHexDigit(byte(lexeme[i])) {
			foundHex = true
			continue
		} else {
			return false
		}
	}
	if foundHex {
		return lexeme[len(lexeme)-1] == 'H'
	}
	return lexeme[len(lexeme)-1] == 'H' || isDigit(byte(lexeme[len(lexeme)-1]))
}

func isReal(lexeme string) bool {
	if len(lexeme) == 0 {
		return false
	}
	if !isDigit(byte(lexeme[0])) {
		return false
	}
	var i = 1
	for i < len(lexeme) {
		if !isDigit(byte(lexeme[i])) {
			break
		}
		i += 1
	}
	if i == len(lexeme) {
		return false
	}
	if lexeme[i] != '.' {
		return false
	}
	i += 1
	for i < len(lexeme) {
		if !isDigit(byte(lexeme[i])) {
			break
		}
		i += 1
	}
	if i == len(lexeme) {
		return true
	}
	if len(lexeme)-i < 2 {
		return false
	}
	if lexeme[i] == 'E' || lexeme[i] == 'D' {
		i += 1
	}
	if lexeme[i] == '+' || lexeme[i] == '-' {
		i += 1
	}
	if i == len(lexeme) {
		return false
	}
	for i < len(lexeme) {
		if !isDigit(byte(lexeme[i])) {
			return false
		}
		i += 1
	}
	return true
}

func isNumber(lexeme string) bool {
	return isInteger(lexeme) || isReal(lexeme)
}

func isIdent(lexeme string) bool {
	var _isIdent = len(lexeme) > 0 && isLetter(lexeme[0])
	if !_isIdent {
		return false
	}
	for i := 1; i < len(lexeme); i++ {
		if !isLetter(lexeme[i]) && !isDigit(lexeme[i]) {
			return false
		}
	}
	return true
}

func lexer(contents []byte) LexerResult {
	//fmt.Println(fmt.Printf("Total number of characters: %d", len(contents)))
	var i = 0
	var lineNo = 1
	var columnNo = 1
	var currentLexeme = ""
	var inComment = false
	var lexemes = new([]Lexeme)
	var inIdent = false
	var inNumber = false
	var inString = false
	var err = false
	for i < len(contents) {
		if i < len(contents)-1 && string(contents[i]) == "(" && string(contents[i+1]) == "*" {
			inComment = true
			i += 2
		} else if i < len(contents)-1 && string(contents[i]) == "*" && string(contents[i+1]) == ")" {
			inComment = false
			i += 2
		} else if !inComment && inNumber && (rune(contents[i]) == '.' || rune(contents[i]) == '+' || rune(contents[i]) == '-') {
			currentLexeme += string(contents[i])
			i += 1
		} else if !inComment && (inIdent || inNumber || inString) && (isOperator(string(contents[i])) || isWhitespace(contents[i])) {
			if isReservedWord(currentLexeme) {
				*lexemes = append(*lexemes, Lexeme{label: currentLexeme, typ: RESERVED_WORD})
			} else if isPredefinedIdentifier(currentLexeme) {
				*lexemes = append(*lexemes, Lexeme{label: currentLexeme, typ: PREDEFINED})
			} else if isString(currentLexeme) {
				*lexemes = append(*lexemes, Lexeme{label: currentLexeme, typ: STRING})
			} else if isNumber(currentLexeme) {
				*lexemes = append(*lexemes, Lexeme{label: currentLexeme, typ: NUMBER})
			} else if isIdent(currentLexeme) {
				*lexemes = append(*lexemes, Lexeme{label: currentLexeme, typ: IDENT})
			} else {
				fmt.Println(fmt.Sprintf("unrecognized token at line %d, column %d", lineNo, columnNo))
				err = true
				break
			}
			inIdent = false
			inNumber = false
			inString = false
			currentLexeme = ""
		} else if !inComment && i < len(contents)-1 && isOperator(string(contents[i])) && isOperator(string(contents[i+1])) {
			if isOperator(string(contents[i]) + string(contents[i+1])) {
				*lexemes = append(*lexemes, Lexeme{label: string(contents[i]) + string(contents[i+1]), typ: OP_OR_DELIM})
			} else {
				*lexemes = append(*lexemes, Lexeme{label: string(contents[i]), typ: OP_OR_DELIM})
				*lexemes = append(*lexemes, Lexeme{label: string(contents[i+1]), typ: OP_OR_DELIM})
			}
			i += 2
		} else if !inComment && !inString && isDigit(contents[i]) {
			currentLexeme += string(contents[i])
			inNumber = true
			i += 1
		} else if !inComment && isOperator(string(contents[i])) {
			*lexemes = append(*lexemes, Lexeme{label: string(contents[i]), typ: OP_OR_DELIM})
			i += 1
		} else if isWhitespace(contents[i]) {
			if contents[i] == 10 {
				columnNo = 1
				lineNo += 1
			}
			i += 1
		} else if !inComment && !inString && rune(contents[i]) == '"' {
			currentLexeme += string(contents[i])
			inString = true
			i += 1
		} else if inString && rune(contents[i]) == '"' {
			currentLexeme += string(contents[i])
			inString = false
			i += 1
		} else {
			if !inComment {
				currentLexeme += string(contents[i])
				inIdent = true
			}
			i += 1
		}
	}
	/* for _, ch := range *lexemes {*/
	//fmt.Println(ch)
	/*}*/
	if err {
		return LexerResult{err: fmt.Errorf("error while processing source"), lexemes: lexemes}
	}
	return LexerResult{lexemes: lexemes}
}
