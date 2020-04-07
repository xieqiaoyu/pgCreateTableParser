%{
package tableParser
%}

%union{
	val string
	column *TableColumn
	columns []*TableColumn
}

%token <val> tokenError
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

%type <column> ddl_table_column
%type <columns> ddl_table_columns

%type <val> ddl_symbol ddl_column_name ddl_data_type ddl_value

%%
ddl: ddl_create

ddl_create: ddl_create_table

ddl_create_table
	: tokenCreate tokenTable ddl_tableName ddl_create_table_body
	| tokenCreate tokenTable tokenIF tokenNOT tokenEXISTS ddl_tableName ddl_create_table_body

ddl_tableName
	: ddl_symbol
	{
		yylex.(*lexer).ast.Table=$1
	}
	| ddl_symbol tokenDot ddl_symbol
	{
		yylex.(*lexer).ast.Schema=$1
		yylex.(*lexer).ast.Table=$3
	}
	;

ddl_create_table_body
	: tokenLeftParen ddl_table_columns  ddl_table_column tokenRightParen tokenSemicolon
	{
		ast := yylex.(*lexer).ast
		ast.Columns = append($2,$3)
	}
	 | tokenLeftParen ddl_table_column tokenRightParen tokenSemicolon
	 {
		ast := yylex.(*lexer).ast
		ast.Columns  = []*TableColumn{
			$2,
		}
	 }

ddl_table_columns
	: ddl_table_column tokenComma
	{
		$$ = []*TableColumn{
			$1,
		}
	}
	| ddl_table_columns ddl_table_column tokenComma
	{
		$$ = append($1,$2)
	}

ddl_table_column
	: ddl_column_name ddl_data_type ddl_column_constraint
	{
	 $$ = &TableColumn{
		Name: $1,
		Type: $2,
	 }
	}

ddl_column_name
	: ddl_symbol
ddl_data_type
	: tokenString

ddl_column_constraint
	: tokenString
	| tokenNOT tokenNULL
	| tokenDEFAULT ddl_default_expr
	| ddl_column_constraint tokenString
	| ddl_column_constraint tokenNOT tokenNULL
	| ddl_column_constraint tokenDEFAULT ddl_default_expr

ddl_default_expr
	: ddl_value
	| ddl_symbol tokenDot ddl_symbol tokenLeftParen tokenRightParen
	| ddl_symbol tokenLeftParen tokenRightParen

ddl_symbol
	: tokenString
	| tokenPgSymbol
ddl_value
	: tokenString
	| tokenPgValue
%%
