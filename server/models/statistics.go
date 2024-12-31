package models

type SubcategoryExpenseBreakdown struct {
	CategoryName     string  `json:"category_name"`
	CategoryColor    string  `json:"category_color"`
	SubcategoryName  string  `json:"subcategory_name"`
	SubcategoryColor string  `json:"subcategory_color"`
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int     `json:"transaction_count"`
}


type MonthlySpending struct {
    Month       string  `json:"month"`
    TotalAmount float64 `json:"total_amount"`
}