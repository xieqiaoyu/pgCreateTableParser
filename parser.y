%{
package tableParser

%}

%union{
	stringVal string
	stringsVal []string
	boolVal bool
	column columnObj
	t_constraint  TableConstraint
	t_header tableHeader
	t_body tableBody
}

%token <stringVal> tokenError
       tokenEOF
       tokenUnknown
       tokenString
       tokenPgSymbol
       tokenPgValue
       tokenLeftParen
       tokenRightParen
       tokenComma
       tokenSemicolon
       tokenDot

%token tokenKeyword

%token tokenCreate
       tokenTable
       tokenIF
       tokenNOT
       tokenEXISTS
       tokenNULL
       tokenDEFAULT
       tokenUNIQUE
       tokenPRIMARY
       tokenKEY

%type <column> ddl_table_column ddl_column_constraint
%type <t_header> ddl_create_table_header ddl_tableName
%type <t_body> ddl_create_table_body
%type <t_constraint>  ddl_table_constraint

%type <stringVal> ddl_symbol ddl_column_name ddl_data_type ddl_value
%type <stringsVal> ddl_column_names
%type <boolVal> ddl_column_primary_key

%%
stmtblock: ddlmulti

ddlmulti: ddlmulti tokenSemicolon ddl
		| ddl

ddl: ddl_create_table
   | /* Empty */

ddl_create_table
	: ddl_create_table_header tokenLeftParen ddl_create_table_body tokenRightParen
	{
		columns := []*TableColumn{}
		constraint := $3.constraint
		for _,obj := range $3.columns {
			columns = append(columns,obj.Column())
			if obj.PrimaryKey {
				constraint.PrimaryKey = append(constraint.PrimaryKey,obj.Name)
			}
			if obj.Unique {
				constraint.Uniques = append(constraint.Uniques,[]string{obj.Name})
			}
		}
		ast := yylex.(*lexer).ast
		yylex.(*lexer).ast = append(ast,&TableDefine{
			Schema: $1.Schema,
			Table: $1.Table,
			Columns: columns,
			Constraint: &constraint,
		})
	}
ddl_create_table_header
	 :tokenCreate tokenTable ddl_tableName
	 {
		$$ = $3
	 }
	 |tokenCreate tokenTable tokenIF tokenNOT tokenEXISTS ddl_tableName
	 {
		$$ = $6
	 }

ddl_tableName
	: ddl_symbol
	{
		$$.Table = $1
	}
	| ddl_symbol tokenDot ddl_symbol
	{
		$$.Schema = $1
		$$.Table = $3
	}

ddl_create_table_body
	: ddl_table_column
	{
		$$.columns = []columnObj{$1}
	}
	| ddl_create_table_body tokenComma ddl_table_column
	{
		$$.columns = append($$.columns,$3)
	}
	| ddl_table_constraint
	{
		$$.constraint = $1
	}
	| ddl_create_table_body tokenComma ddl_table_constraint
	{
		$$.constraint = combineConstraint($$.constraint,$3)
	}

ddl_table_column
	: ddl_column_name ddl_data_type ddl_column_constraint
	{
	 $$ = $3
	 $$.Name = $1
	 $$.Type = $2
	}

ddl_column_name
	: ddl_symbol
ddl_data_type
	: ddl_symbol
	| ddl_symbol tokenDot ddl_symbol
	{
		$$= __yyfmt__.Sprintf("%s.%s",$1,$3)
	}

ddl_column_constraint
	: tokenUNIQUE
	{
		$$.Unique = true
	}
	| ddl_column_primary_key
	{
		$$.PrimaryKey = true
	}
	| tokenNOT tokenNULL
	{
		$$.NotNull = true
	}
	| tokenDEFAULT ddl_default_expr
	{
	}
	| ddl_column_constraint tokenUNIQUE
	{
		$$.PrimaryKey = true
	}
	| ddl_column_constraint ddl_column_primary_key
	{
		$$.PrimaryKey = true
	}
	| ddl_column_constraint tokenNOT tokenNULL
	{
		$$.NotNull = true
	}
	| ddl_column_constraint tokenDEFAULT ddl_default_expr

ddl_column_primary_key
	: tokenPRIMARY tokenKEY{}

ddl_default_expr
	: ddl_value
	| ddl_symbol tokenDot ddl_symbol tokenLeftParen tokenRightParen
	| ddl_symbol tokenLeftParen tokenRightParen

ddl_table_constraint
	: tokenUNIQUE tokenLeftParen ddl_column_names tokenRightParen
	{
		$$.Uniques = append($$.Uniques,$3)
	}

ddl_column_names
	: ddl_column_name
	{
		$$= []string{$1}
	}
	| ddl_column_names tokenComma ddl_column_name
	{
		$$= append($1,$3)
	}

ddl_symbol
	: tokenString
	| tokenPgSymbol
ddl_value
	: tokenString
	| tokenPgValue
%%
