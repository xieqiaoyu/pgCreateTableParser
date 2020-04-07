package tableParser

type TableDefine struct {
	Schema  string
	Table   string
	Columns []*TableColumn
}

type TableColumn struct {
	Name string
	Type string
}

func ParseTable(name, sql string) (*TableDefine, error) {
	yyErrorVerbose = true
	l := lex(name, sql)
	p := &yyParserImpl{}
	if p.Parse(l) != 0 {
		return nil, l.lerror
	}
	return l.ast, nil
}
