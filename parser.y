%{
package tableParser
%}

%union{
	val string
	column *TableColumn
	columns []*TableColumn
}

%token tokenError
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

%%
ddl: ddl_create

ddl_create: ddl_create_table

ddl_create_table
	: tokenCreate tokenTable ddl_tableName ddl_create_table_body
	| tokenCreate tokenTable tokenIF tokenNOT tokenEXISTS ddl_tableName ddl_create_table_body

ddl_tableName
	: ddl_symbol
	{
		yylex.(*lexer).ast.Table=$1.val
	}
	| ddl_symbol tokenDot ddl_symbol
	{
		yylex.(*lexer).ast.Schema=$1.val
		yylex.(*lexer).ast.Table=$3.val
	}
	;

ddl_create_table_body
	: tokenLeftParen ddl_table_columns  ddl_table_column tokenRightParen tokenSemicolon
	{
		ast := yylex.(*lexer).ast
		ast.Columns = append($2.columns,$3.column)
	}
	 | tokenLeftParen ddl_table_column tokenRightParen tokenSemicolon
	 {
		ast := yylex.(*lexer).ast
		ast.Columns  = []*TableColumn{
			$2.column,
		}
	 }

ddl_table_columns
	: ddl_table_column tokenComma
	{
		$$.columns = []*TableColumn{
			$1.column,
		}
	}
	| ddl_table_columns ddl_table_column tokenComma
	{
		$$.columns = append($1.columns,$2.column)
	}

ddl_table_column
	: ddl_column_name ddl_data_type ddl_column_constraint
	{
	 $$.column = &TableColumn{
		Name: $1.val,
		Type: $2.val,
	 }
	}

ddl_column_name
	: ddl_symbol
	{
		$$.val = $1.val
	}
	;
ddl_data_type
	: tokenString
	{
		$$.val = $1.val
	}
	;

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
	{
		$$.val = $1.val
	}
	| tokenPgSymbol
	{
		$$.val = $1.val
	}
ddl_value
	: tokenString
	{
		$$.val = $1.val
	}
	| tokenPgValue
	{
		$$.val = $1.val
	}
%%
