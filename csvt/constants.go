package csvt

const (
	HEA_SEPARATOR rune = ';' 
	MAP_SEPARATOR rune = ','
	MAP_LINKER rune = '='
	MAP_CLOSING rune = '^'
	ARR_SEPARATOR rune = ','
	ARR_CLOSING rune = '|'
	STR_SEPARATOR rune = ';'
	STR_CLOSING rune = ':'
	PTR_HEADER rune = '$'
	PTR_SEPARATOR rune = '_'
	TBL_HEAD_BASE rune = '/'
	TBL_HEAD_ROOT rune = '*'
	TBL_INDEX_HEAD rune = 'H'
)