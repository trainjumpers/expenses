package parser

import "expenses/internal/models"

var parserRegistry = make(map[models.BankType]BankStatementParser)

type BankStatementParser interface {
	Parse(fileBytes []byte) ([]models.CreateTransactionInput, error)
}

func RegisterParser(bankType models.BankType, parser BankStatementParser) {
	parserRegistry[bankType] = parser
}

func GetParser(bankType models.BankType) (BankStatementParser, bool) {
	parser, ok := parserRegistry[bankType]
	return parser, ok
}
