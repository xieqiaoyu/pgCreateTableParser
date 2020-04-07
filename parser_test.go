package tableParser

import (
	"testing"
)

const mysql = `CREATE TABLE IF NOT EXISTS admin.users (
  "id"  SERIAL PRIMARY KEY,
  "uuid" UUID UNIQUE NOT NULL DEFAULT public.uuid_generate_v4(),
  "name" TEXT NOT NULL DEFAULT '',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "info" JSONB NOT NULL DEFAULT '{}'
);`

func TestParser(t *testing.T) {
	yyDebug = 0
	yyErrorVerbose = true
	def, err := ParseTable("table1", mysql)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	t.Logf("schema:%s table:%s", def.Schema, def.Table)
	for index, col := range def.Columns {
		t.Logf("%d|%s|%s\n", index, col.Name, col.Type)
	}
}
