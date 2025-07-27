package parser

import "expenses/internal/models"

// Parser defines the interface for different bank statement parsers.
type Parser interface {
	Parse(fileBytes []byte, metadata string, fileName string) ([]models.CreateTransactionInput, error)
}

var parserRegistry = make(map[models.BankType]Parser)

func RegisterParser(bankType models.BankType, parser Parser) {
	parserRegistry[bankType] = parser
}

func GetParser(bankType models.BankType) (Parser, bool) {
	parser, ok := parserRegistry[bankType]
	return parser, ok
}
