package parser

import (
	"github.com/yassinebenaid/bunny/ast"
	"github.com/yassinebenaid/bunny/token"
)

func (p *Parser) parseCommandSubstitution() ast.Expression {
	var cmds ast.CommandSubstitution
	p.proceed()
	for p.curr.Type == token.BLANK || p.curr.Type == token.NEWLINE {
		p.proceed()
	}

	for p.curr.Type != token.RIGHT_PAREN && p.curr.Type != token.EOF {
		cmdList := p.parseCommandList()
		if cmdList == nil {
			return nil
		}
		cmds = append(cmds, cmdList)
		if p.curr.Type == token.SEMICOLON || p.curr.Type == token.AMPERSAND {
			p.proceed()
		}
		for p.curr.Type == token.BLANK || p.curr.Type == token.NEWLINE {
			p.proceed()
		}
	}

	if len(cmds) == 0 {
		p.error("expeceted a command list after `$(`")
	}

	if p.curr.Type != token.RIGHT_PAREN {
		p.error("unexpected end of file, expeceted `)`")
	}

	return cmds
}

func (p *Parser) parseProcessSubstitution() ast.Expression {
	tok := p.curr.Literal
	var process ast.ProcessSubstitution

	process.Direction = '>'
	if tok == "<(" {
		process.Direction = '<'
	}

	p.proceed()
	for p.curr.Type == token.BLANK || p.curr.Type == token.NEWLINE {
		p.proceed()
	}

	for p.curr.Type != token.RIGHT_PAREN && p.curr.Type != token.EOF {
		cmdList := p.parseCommandList()
		if cmdList == nil {
			return nil
		}
		process.Body = append(process.Body, cmdList)
		if p.curr.Type == token.SEMICOLON || p.curr.Type == token.AMPERSAND {
			p.proceed()
		}
		for p.curr.Type == token.BLANK || p.curr.Type == token.NEWLINE {
			p.proceed()
		}
	}

	if len(process.Body) == 0 {
		p.error("expeceted a command list after `%s`", tok)
	}

	if p.curr.Type != token.RIGHT_PAREN {
		p.error("unexpected end of file, expeceted `)`")
	}

	return process
}

func (p *Parser) parseParameterExpansion() ast.Expression {
	var exp ast.Expression
	p.proceed()

	if p.curr.Type == token.HASH {
		p.proceed()
		exp = ast.VarCount(p.parseParameter())

		if p.curr.Type != token.RIGHT_BRACE {
			p.error("expected closing brace `}`, found `%s`", p.curr.Literal)
		}
		return exp
	}

	param := p.parseParameter()

	switch p.curr.Type {
	case token.RIGHT_BRACE:
		exp = ast.Var(param)
	case token.MINUS, token.COLON_MINUS:
		checkForNull := p.curr.Type == token.COLON_MINUS
		p.proceed()
		exp = ast.VarOrDefault{
			Name:         param,
			Default:      p.parseExpansionOperandExpression(0),
			CheckForNull: checkForNull,
		}
	case token.COLON_ASSIGN:
		p.proceed()
		exp = ast.VarOrSet{
			Name:    param,
			Default: p.parseExpansionOperandExpression(0),
		}
	case token.COLON_QUESTION:
		p.proceed()
		exp = ast.VarOrFail{
			Name:  param,
			Error: p.parseExpansionOperandExpression(0),
		}
	case token.COLON_PLUS:
		p.proceed()
		exp = ast.CheckAndUse{
			Name:  param,
			Value: p.parseExpansionOperandExpression(0),
		}
	case token.CIRCUMFLEX, token.DOUBLE_CIRCUMFLEX, token.COMMA, token.DOUBLE_COMMA:
		operator := p.curr.Literal
		p.proceed()
		exp = ast.ChangeCase{
			Name:     param,
			Operator: operator,
			Pattern:  p.parseExpansionOperandExpression(0),
		}
	case token.HASH, token.PERCENT, token.DOUBLE_PERCENT:
		operator := p.curr.Literal
		if p.curr.Type == token.HASH && p.next.Type == token.HASH {
			p.proceed()
			operator += p.curr.Literal
		}
		p.proceed()

		exp = ast.MatchAndRemove{
			Name:     param,
			Operator: operator,
			Pattern:  p.parseExpansionOperandExpression(0),
		}
	case token.SLASH:
		operator := p.curr.Literal
		p.proceed()
		if p.curr.Type == token.SLASH || p.curr.Type == token.HASH || p.curr.Type == token.PERCENT {
			operator += p.curr.Literal
			p.proceed()
		}

		var pattern ast.Expression
		if p.curr.Type == token.SLASH {
			pattern = ast.Word(p.curr.Literal)
		} else {
			pattern = p.parseExpansionOperandExpression(token.SLASH)
		}

		mar := ast.MatchAndReplace{Name: param, Operator: operator, Pattern: pattern}

		if p.curr.Type == token.SLASH {
			p.proceed()
			mar.Value = p.parseExpansionOperandExpression(0)
		}

		exp = mar
	case token.COLON:
		p.proceed()
		slice := ast.Slice{
			Name:   param,
			Offset: p.parseArithmetics(),
		}

		if p.curr.Type == token.COLON {
			p.proceed()
			slice.Length = p.parseArithmetics()
		}

		exp = slice
	case token.AT:
		p.proceed()
		exp = ast.Transform{
			Name:     param,
			Operator: p.curr.Literal,
		}
		p.proceed()
	}

	if p.curr.Type != token.RIGHT_BRACE {
		p.error("expected closing brace `}`, found `%s`", p.curr.Literal)
	}

	return exp
}

func (p *Parser) parseExpansionOperandExpression(stopAt token.TokenType) ast.Expression {
	var exprs []ast.Expression

loop:
	for {
		switch p.curr.Type {
		case token.RIGHT_BRACE, token.EOF:
			break loop
		case token.SIMPLE_EXPANSION:
			exprs = append(exprs, ast.Var(p.curr.Literal))
		case token.SINGLE_QUOTE:
			exprs = append(exprs, p.parseLiteralString())
		case token.DOUBLE_QUOTE:
			exprs = append(exprs, p.parseString())
		case token.DOLLAR_PAREN:
			exprs = append(exprs, p.parseCommandSubstitution())
		case token.GT_PAREN, token.LT_PAREN:
			exprs = append(exprs, p.parseProcessSubstitution())
		case token.DOLLAR_BRACE:
			exprs = append(exprs, p.parseParameterExpansion())
		default:
			if p.curr.Type == stopAt {
				break loop
			}

			exprs = append(exprs, ast.Word(p.curr.Literal))
		}

		p.proceed()
	}

	return concat(exprs)
}

func (p *Parser) parseParameter() string {
	if p.curr.Type != token.WORD {
		p.error("couldn't find a valid parameter name, found `%s`", p.curr.Literal)
	}

	v := p.curr.Literal
	p.proceed()
	// TODO: handle arrays

	return v
}
