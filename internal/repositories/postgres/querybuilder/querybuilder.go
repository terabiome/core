package querybuilder

func Insert(table string) *insert {
	return &insert{table: table}
}
