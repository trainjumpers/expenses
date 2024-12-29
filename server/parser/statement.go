package parser

import (
	"expenses/entities"
	"expenses/logger"
	"strconv"
	"strings"
	"time"
)

func ParseBankStatement(records [][]string) ([]entities.Statement, error) {
	logger.Info("Parsing bank statement")
	records = cleanRecord(records)
	return parseBankRecords(records)
}

func cleanRecord(records [][]string) [][]string {
	logger.Info("Sanitizing records to remove empty rows")
	var cleanRecords [][]string
	for _, row := range records {
		if len(row) == 0 {
			continue
		}
		for i, cell := range row {
			row[i] = strings.TrimSpace(cell)
			if i > 3 {
				row[i] = strings.ReplaceAll(cell, ",", "")
			}
		}
		cleanRecords = append(cleanRecords, row)
	}
	return cleanRecords
}

func parseBankRecords(records [][]string) ([]entities.Statement, error) {
	var statements []entities.Statement
	logger.Info("Parsing bank statement")
	for _, row := range records {
		if !validateRow(row) {
			continue
		}
		transactionDate, err := time.Parse("02-Jan-06", row[0])
		if err != nil {
			logger.Error("Error parsing date: ", err)
			continue
		}
		amountRow := strings.TrimSpace(row[4])
		if amountRow == "" {
			amountRow = "-" + strings.TrimSpace(row[5])
		}
		amount, err := strconv.ParseFloat(amountRow, 64)
		if err != nil {
			logger.Error("Error parsing amount: ", err)
			continue
		}

		balance, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			logger.Error("Error parsing balance: ", err)
			continue
		}

		statement := entities.Statement{
			TrasactionId: row[3],
			Date:         transactionDate,
			Description:  row[2],
			Amount:       amount,
			Balance:      balance,
		}
		statements = append(statements, statement)
	}
	return statements, nil
}

func validateRow(row []string) bool {
	if len(row) != 7 {
		logger.Debug("Skipping row since this is probably headers: ", row, " with length: ", len(row))
		return false
	}
	_, err := time.Parse("02-Jan-06", row[0])
	if err != nil {
		logger.Debug("Invalid date format: ", row[0])
		return false
	}
	return true
}
