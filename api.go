package tableParser

import (
	"fmt"
	"strings"
)

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

func Define2String(def *TableDefine) string {
	columns := "\n\t---------------+---------------"
	for _, col := range def.Columns {
		columns += fmt.Sprintf("\n\t%-15s|%-15s", col.Name, col.Type)
	}
	constraints := "\n\tConstraints:"
	if len(def.Constraint.PrimaryKey) > 0 {
		constraints += fmt.Sprintf("\n\t\tPK: %s", strings.Join(def.Constraint.PrimaryKey, ","))
	}
	if len(def.Constraint.Uniques) > 0 {
		for _, unique := range def.Constraint.Uniques {
			constraints += fmt.Sprintf("\n\t\tUnique: (%s)", strings.Join(unique, ","))
		}
	}

	return fmt.Sprintf(" Table \"%s\".\"%s\" %s\n%s\n", def.Schema, def.Table, columns, constraints)
}
