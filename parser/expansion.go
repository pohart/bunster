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

	param := p.curr.Literal
	p.proceed()

	switch p.curr.Type {
	case token.RIGHT_BRACE:
		exp = ast.Var(param)
	case token.MINUS, token.COLON_MINUS:
		checkForNull := p.curr.Type == token.COLON_MINUS
		p.proceed()
		exp = ast.VarOrDefault{
			Name:         param,
			Default:      p.parseExpansionOperandExpression(),
			CheckForNull: checkForNull,
		}
	case token.COLON_ASSIGN:
		p.proceed()
		exp = ast.VarOrSet{
			Name:    param,
			Default: p.parseExpansionOperandExpression(),
		}
	}

	if p.curr.Type != token.RIGHT_BRACE {
		panic("Not }, it is: " + p.curr.Literal)
	}

	return exp
}

func (p *Parser) parseExpansionOperandExpression() ast.Expression {
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
			exprs = append(exprs, ast.Word(p.curr.Literal))
		}

		p.proceed()
	}

	return concat(exprs)
}
