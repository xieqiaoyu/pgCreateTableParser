package tableParser

import (
	"fmt"
	"reflect"
	"testing"
)

type parseTest struct {
	name         string
	input        string
	assertDefine *TableDefine
}

func defineEqual(d1, d2 *TableDefine) (bool, error) {
	if d1.Schema != d2.Schema {
		return false, fmt.Errorf("schema name is not equal")
	}
	if d1.Table != d2.Table {
		return false, fmt.Errorf("table name is not equal")
	}
	if len(d1.Columns) != len(d2.Columns) {
		return false, fmt.Errorf("column num is not euqal")
	}
	for index, column := range d1.Columns {
		if column.Name != d2.Columns[index].Name {
			return false, fmt.Errorf("%d column Name %s is not equal to %s", index, column.Name, d2.Columns[index].Name)
		}
		if column.Type != d2.Columns[index].Type {
			return false, fmt.Errorf("%d column Type %s is not equal to %s", index, column.Type, d2.Columns[index].Type)

		}
	}
	if d1.Constraint != nil {
		if d2.Constraint == nil {
			return false, fmt.Errorf("Constraint is not equal")
		}
		if !reflect.DeepEqual(d1.Constraint.PrimaryKey, d2.Constraint.PrimaryKey) {
			return false, fmt.Errorf("Primary key is not equal")
		}
	} else if d2.Constraint != nil {
		return false, fmt.Errorf("Constraint is not equal")
	}
	return true, nil
}

func makeDefine(schema, table string, columns [][]string, constraint *TableConstraint) *TableDefine {
	cols := []*TableColumn{}
	for _, column := range columns {
		cols = append(cols, &TableColumn{column[0], column[1], false}) // nullable false
	}
	return &TableDefine{
		Schema:     schema,
		Table:      table,
		Columns:    cols,
		Constraint: constraint,
	}
}

const (
	baseCreate = `CREATE TABLE IF NOT EXISTS admin.users (
  "id"  SERIAL PRIMARY KEY,
  "uuid" UUID UNIQUE NOT NULL DEFAULT public.uuid_generate_v4(),
  "name" TEXT NOT NULL DEFAULT '',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "info" JSONB NOT NULL DEFAULT '{}'
);`
	constantCreate = `CREATE TABLE IF NOT EXISTS admin.users2 (
    "id"  SERIAL PRIMARY KEY,
    "age" INTEGER NOT NULL,
    "type" admin.user_type NOT NULL,
    "trait" TEXT NOT NULL DEFAULT '',
    "extra" JSONB NOT NULL DEFAULT '{}',
    UNIQUE ("type","trait")
);`
	noconstantCreate = `CREATE TABLE noconstant (
    "id" SERIAL,
	"type" TEXT)
`
	numberDefaultCreate = ` CREATE TABLE numberDefault (
    "id"  SERIAL,
	"age" INTEGER NOT NULL DEFAULT 0
);
`
)

var parserTests = []parseTest{
	{"base", baseCreate, makeDefine("admin", "users", [][]string{
		{"id", "SERIAL"},
		{"uuid", "UUID"},
		{"name", "TEXT"},
		{"created_at", "TIMESTAMPTZ"},
		{"info", "JSONB"},
	}, &TableConstraint{
		PrimaryKey: []string{"id"},
		Uniques:    [][]string{[]string{"uuid"}},
	}),
	},
	{"constant", constantCreate, makeDefine("admin", "users2", [][]string{
		{"id", "SERIAL"},
		{"age", "INTEGER"},
		{"type", "admin.user_type"},
		{"trait", "TEXT"},
		{"extra", "JSONB"},
	}, &TableConstraint{
		PrimaryKey: []string{"id"},
		Uniques:    [][]string{[]string{"type", "trait"}},
	}),
	},
	{
		"noconstantCreate", noconstantCreate, makeDefine("", "noconstant", [][]string{
			{"id", "SERIAL"},
			{"type", "TEXT"},
		}, &TableConstraint{}),
	},
	{
		"defaultNumberCreate", numberDefaultCreate, makeDefine("", "numberDefault", [][]string{
			{"id", "SERIAL"},
			{"age", "INTEGER"},
		}, &TableConstraint{}),
	},
}

func TestParser(t *testing.T) {
	yyDebug = 0
	yyErrorVerbose = true
	for _, test := range parserTests {
		defs, err := ParseTable(test.name, test.input)
		if err != nil {
			t.Errorf("parse %s err :%s", test.name, err)
			continue
		}
		if len(defs) != 1 {
			t.Errorf("Wrong parse define num ,get %d assume 1 ", len(defs))
			continue
		}
		def := defs[0]
		equal, err := defineEqual(def, test.assertDefine)
		if !equal {
			t.Errorf("parse %s err :%s, got\n\t%+v\nexpect\n\t%+v", test.name, err, Define2String(def), Define2String(test.assertDefine))
			continue
		} else {
			t.Logf("Get Define:\n\t%s\n", Define2String(def))
		}

	}
}
