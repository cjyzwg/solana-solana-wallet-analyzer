package analyze

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
)

var (
	APIURL  string
	Network string
	Account string
	APIKey  string
	Debug   bool
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	APIURL = getEnv("API_URL")
	Network = getEnv("NETWORK")
	Account = getEnv("ACCOUNT")
	APIKey = getEnv("API_KEY")
	Debug = getEnv("DEBUG") == "true"
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return value
}

func getLatestTransactionSignature(apiURL, network, account string) (string, int64, error) {
	log.Println("Fetching latest transaction for account:", account)
	headers := map[string]string{
		"x-api-key": APIKey,
	}
	url := fmt.Sprintf("%s?network=%s&account=%s&tx_num=1&enable_raw=true&enable_events=true", apiURL, network, account)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {

		return "", 0, fmt.Errorf("error in API request: %d", resp.StatusCode)
	}
	var data Response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", 0, fmt.Errorf("error decoding JSON: %v", err)
	}

	if !data.Success {
		return "", 0, fmt.Errorf("failed to fetch transactions: %s", data.Message)
	}

	// err = json.NewDecoder(resp.Body).Decode(&data)
	// if err != nil {
	// 	return "", 0, err
	// }
	if len(data.Result) > 0 {
		signature := data.Result[0].Signatures[0]
		blockTime := data.Result[0].Raw.BlockTime
		return signature, blockTime, nil
	}
	return "", 0, fmt.Errorf("failed to fetch the latest transaction")
}

func FetchAndParseTransactions(apiURL, network, account string, timeDelta time.Duration) ([]Transaction, error) {
	signature, blockTime, err := getLatestTransactionSignature(apiURL, network, account)
	if err != nil {
		return nil, err
	}

	endTime := time.Unix(blockTime, 0)
	startTime := endTime.Add(-timeDelta)
	if timeDelta == 0 {
		startTime = time.Time{}
	}
	log.Printf("Fetching transactions from %s to %s", startTime, endTime)
	var transactions []Transaction
	beforeTxSignature := signature
	apiCalls := 0
	continueFetching := true

	for continueFetching {
		apiCalls++
		log.Printf("API call #%d, before_tx_signature: %s", apiCalls, beforeTxSignature)
		// params := map[string]string{
		// 	"network":             network,
		// 	"account":             account,
		// 	"tx_num":              "100",
		// 	"enable_raw":          "true",
		// 	"enable_events":       "true",
		// 	"before_tx_signature": beforeTxSignature,
		// }
		url := fmt.Sprintf("%s?network=%s&account=%s&tx_num=100&enable_raw=true&enable_events=true&before_tx_signature=%s", apiURL, network, account, beforeTxSignature)
		log.Println(url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("x-api-key", APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error in API request: %d", resp.StatusCode)
		}

		var data Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("error decoding JSON: %v", err)
		}

		batch := data.Result
		if len(batch) == 0 {
			log.Println("No more transactions to fetch")
			break
		}

		log.Printf("Fetched %d transactions in this batch", len(batch))

		batchStartTime := time.Unix(batch[len(batch)-1].Raw.BlockTime, 0)
		batchEndTime := time.Unix(batch[0].Raw.BlockTime, 0)

		for _, tx := range batch {
			txTime := time.Unix(tx.Raw.BlockTime, 0)
			if !startTime.IsZero() && txTime.Before(startTime) {
				continueFetching = false
				break
			}
			if txTime.After(endTime) {
				continue
			}
			transactions = append(transactions, ParseTransaction(tx))
		}

		if continueFetching && len(batch) > 0 {
			beforeTxSignature = batch[len(batch)-1].Signatures[0]
		} else {
			break
		}

		// Update progress
		progressMsg := fmt.Sprintf("\rAPI calls: %d, Parsed transactions from %s to %s", apiCalls, batchStartTime, batchEndTime)
		fmt.Print(progressMsg)
	}

	fmt.Println() // New line after progress updates
	// Reverse transactions to get chronological order
	for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
		transactions[i], transactions[j] = transactions[j], transactions[i]
	}

	log.Printf("Total API calls made: %d", apiCalls)
	log.Printf("Total transactions fetched: %d", len(transactions))
	return transactions, nil
}

func ParseTransaction(tx Transaction) Transaction {
	blockTime := tx.Raw.BlockTime
	blocktimeUTC := time.Unix(blockTime, 0).UTC().Format("2006-01-02 15:04:05")
	signature := "N/A"
	if len(tx.Signatures) > 0 {
		signature = tx.Signatures[0]
	}
	slot := tx.Raw.Slot
	status := tx.Status
	computeUnit := strconv.Itoa(tx.Raw.Meta.ComputeUnitsConsumed)
	fee := tx.Raw.Meta.Fee

	tokenIn, tokenOut := Token{}, Token{}
	if len(tx.Actions) > 0 {
		action := tx.Actions[0]
		tokenSwapped := action.Info.TokensSwapped
		tokenIn = tokenSwapped.In
		tokenOut = tokenSwapped.Out
	}

	memo := "N/A"
	for _, instruction := range tx.Raw.Transaction.Message.Instructions {
		var parsedString string
		if err := json.Unmarshal(instruction.Parsed, &parsedString); err == nil && parsedString != "" {
			memo = parsedString
			break
		}
	}
	slotStr := strconv.Itoa(slot)
	feeStr := fmt.Sprintf("%.2f", fee)

	tx.ParsedTranction.BlocktimeUTC = blocktimeUTC
	tx.ParsedTranction.SlotStr = slotStr
	tx.ParsedTranction.Status = status
	tx.ParsedTranction.FeeStr = feeStr
	tx.ParsedTranction.ComputeUnit = computeUnit

	tx.ParsedTranction.TokenIn = tokenIn
	tx.ParsedTranction.TokenOut = tokenOut
	tx.ParsedTranction.Memo = memo
	tx.ParsedTranction.Signature = signature
	return tx
	// return []string{blocktimeUTC, slotStr, status, feeStr, computeUnit, tokenName, tokenIn, profit, memo, signature}
}
func EncodeTokenJson(token Token) (string, error) {

	jsonStr, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return string(jsonStr), nil

}
func DecodeTokenJson(str string) (Token, error) {
	var data Token
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return data, err
	}
	return data, nil
}

func AnalyzeMemoType(records []Transaction, memoType string) []string {
	var total, success, fail int

	for _, record := range records {
		memo := strings.TrimSpace(record.ParsedTranction.Memo)
		if memoType == "TOTAL" {
			total++
			if record.Status == "Success" {
				success++
			} else if record.Status == "Fail" {
				fail++
			}
		} else if memoType == "N/A" {
			if memo == "" || memo == "N/A" {
				total++
				if record.Status == "Success" {
					success++
				} else if record.Status == "Fail" {
					fail++
				}
			}
		} else if memo == memoType {
			total++
			if record.Status == "Success" {
				success++
			} else if record.Status == "Fail" {
				fail++
			}
		}
	}

	successRate := 0.0
	failRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
		failRate = float64(fail) / float64(total) * 100
	}

	displayName := memoType
	if memoType != "TOTAL" && memoType != "N/A" && !strings.Contains(memoType, "RPC") {
		displayName += " (jito)"
	}

	return []string{
		displayName,
		strconv.Itoa(total),
		strconv.Itoa(success),
		strconv.Itoa(fail),
		fmt.Sprintf("%.2f%%", successRate),
		fmt.Sprintf("%.2f%%", failRate),
	}
}

func GenerateStats(records []Transaction) [][]string {
	log.Println("Generating statistics from records")

	memoSet := make(map[string]struct{})
	for _, record := range records {
		memo := strings.TrimSpace(record.ParsedTranction.Memo)
		if memo != "" && memo != "N/A" {
			memoSet[memo] = struct{}{}
		}
	}

	var memoTypes []string
	for memo := range memoSet {
		memoTypes = append(memoTypes, memo)
	}
	sort.Strings(memoTypes)

	results := [][]string{
		AnalyzeMemoType(records, "TOTAL"),
		AnalyzeMemoType(records, "N/A"),
	}

	for _, memoType := range memoTypes {
		results = append(results, AnalyzeMemoType(records, memoType))
	}

	// Sort results: TOTAL and N/A first, then non-Jito, then Jito
	sort.Slice(results[2:], func(i, j int) bool {
		a := results[2+i][0]
		b := results[2+j][0]
		if strings.Contains(a, "jito") && !strings.Contains(b, "jito") {
			return false
		}
		if !strings.Contains(a, "jito") && strings.Contains(b, "jito") {
			return true
		}
		return a < b
	})

	log.Println("Statistics generation completed")
	return results
}

type TrackTokenStats struct {
	TotalBoughtSOL float64
	TotalSoldSOL   float64
}

func PrintStatistics(transactions []Transaction) {
	totalTxs := len(transactions)
	totalFee := 0.0
	totalComputeUnits := 0
	count := 0
	tokenStats := make(map[string]*TrackTokenStats)

	for _, tx := range transactions {
		parsed := tx.ParsedTranction
		in := parsed.TokenIn
		out := parsed.TokenOut
		totalFee += tx.Raw.Meta.Fee
		totalComputeUnits += tx.Raw.Meta.ComputeUnitsConsumed
		if tx.ParsedTranction.TokenIn.TokenAddress != "" {
			count += 1

		}
		// fmt.Println(in, out, parsed.Signature)
		if in.Symbol == "SOL" && out.Symbol != "SOL" {
			// Update the stats for the input token
			if stats, exists := tokenStats[out.Symbol]; exists {
				stats.TotalBoughtSOL += in.Amount
			} else {
				tokenStats[out.Symbol] = &TrackTokenStats{
					TotalBoughtSOL: in.Amount,
				}
			}
		}
		if in.Symbol != "SOL" && out.Symbol == "SOL" {
			// Update the stats for the input token
			if stats, exists := tokenStats[in.Symbol]; exists {
				stats.TotalSoldSOL += out.Amount
			} else {
				tokenStats[in.Symbol] = &TrackTokenStats{
					TotalSoldSOL: out.Amount,
				}
			}
		}

	}

	// Print results
	for token, stats := range tokenStats {
		profit := stats.TotalSoldSOL - stats.TotalBoughtSOL
		fmt.Printf("Token: %s\n", token)
		fmt.Printf("  Total Wrapped SOL Bought: %.8f SOL\n", stats.TotalBoughtSOL)
		fmt.Printf("  Total Wrapped SOL Sold: %.8f SOL\n", stats.TotalSoldSOL)
		fmt.Printf("  Profit/Loss: %.8f SOL\n\n", profit)
	}

	data := [][]string{
		{"Total Transactions", strconv.Itoa(totalTxs)},
		{"Total Fees", fmt.Sprintf("%.2f", totalFee)},
		{"Total Compute Units", strconv.Itoa(totalComputeUnits)},
	}

	// data := GenerateStats(transactions)
	table := tablewriter.NewWriter(os.Stdout)
	// table.SetHeader([]string{"Memo Type", "Total", "Success", "Fail", "Success %", "Fail %"})
	table.SetHeader([]string{"Total Transactions", "Total Fees", "Total Compute Units"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
