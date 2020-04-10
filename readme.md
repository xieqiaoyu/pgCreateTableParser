### Instruction

This  library can parse postgres create table statement to a struct with table metadata.

**It's just a toy now , not cover all situration**

### Example 

```go
package main

import (
    "fmt"
    parser "github.com/xieqiaoyu/pgCreateTableParser"
)

var stmt = `CREATE TABLE IF NOT EXISTS admin.users (
  "id"  SERIAL PRIMARY KEY,
  "uuid" UUID UNIQUE NOT NULL DEFAULT public.uuid_generate_v4(),
  "name" TEXT NOT NULL DEFAULT '',
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  "info" JSONB NOT NULL DEFAULT '{}'
);`

func main() {
    defs, err := parser.ParseTable("bazinga", stmt)
    if err != nil {
        panic(err)
    }
    fmt.Println(parser.Define2String(defs[0]))
}
```

the output:

```bash

 # go run .
 Table "admin"."users"
	---------------+---------------
	id             |SERIAL
	uuid           |UUID
	name           |TEXT
	created_at     |TIMESTAMPTZ
	info           |JSONB

	Constraints:
		PK: id
		Unique: (uuid)
```



