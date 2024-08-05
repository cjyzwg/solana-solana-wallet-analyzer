package main

import (
	"log"
	"testchromedp/analyze"
	"time"
)

func main() {
	timeDelta := time.Hour * 24 * 30 // Default to 30 days
	transactions, err := analyze.FetchAndParseTransactions(analyze.APIURL, analyze.Network, analyze.Account, timeDelta)
	if err != nil {
		log.Fatalf("Error fetching and parsing transactions: %v", err)
	}

	analyze.PrintStatistics(transactions)
}
