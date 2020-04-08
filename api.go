package tableParser

type TableDefine struct {
	Schema     string
	Table      string
	Columns    []*TableColumn
	Constraint *TableConstraint
}

type TableColumn struct {
	Name string
	Type string
	//Collation string
	Nullable bool
	//Default  string
}

type TableConstraint struct {
	PrimaryKey []string
	Uniques    [][]string
}

func combineConstraint(c1, c2 TableConstraint) TableConstraint {
	return TableConstraint{
		PrimaryKey: append(c1.PrimaryKey, c2.PrimaryKey...),
		Uniques:    append(c1.Uniques, c2.Uniques...),
	}
}

type columnObj struct {
	Name       string
	Type       string
	PrimaryKey bool
	Unique     bool
	NotNull    bool
}

func (o columnObj) Column() *TableColumn {
	return &TableColumn{
		Name:     o.Name,
		Type:     o.Type,
		Nullable: !o.NotNull,
	}
}

type tableHeader struct {
	Schema string
	Table  string
}

type tableBody struct {
	columns    []columnObj
	constraint TableConstraint
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
