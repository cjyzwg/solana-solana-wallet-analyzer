package analyze

import "encoding/json"

// Struct to represent the JSON response structure
type Response struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Result  []Transaction `json:"result"`
}

type Transaction struct {
	Timestamp       string          `json:"timestamp"`
	Fee             float64         `json:"fee"`
	FeePayer        string          `json:"fee_payer"`
	Signers         []string        `json:"signers"`
	Signatures      []string        `json:"signatures"`
	Protocol        Protocol        `json:"protocol"`
	Type            string          `json:"type"`
	Status          string          `json:"status"`
	Actions         []Action        `json:"actions"`
	Events          []Event         `json:"events"`
	Raw             RawData         `json:"raw"`
	ParsedTranction ParsedTranction `json:"parsed_transaction"`
}
type ParsedTranction struct {
	BlocktimeUTC string `json:"blocktime_utc"`
	Status       string `json:"status"`
	SlotStr      string `json:"slot_str"`
	FeeStr       string `json:"fee_str"`
	ComputeUnit  string `json:"computer_unit"`
	TokenIn      Token  `json:"token_in"`
	TokenOut     Token  `json:"token_out"`
	Memo         string `json:"memo"`
	Signature    string `json:"signature"`
}

// type sell insymbol solly inamount 573750.211 inaddress outsymbol sol outamount 1.98 outadress sol11112 signature
type TestParsedTranction struct {
	BlocktimeUTC    string `json:"blocktime_utc"`
	Status          string `json:"status"`
	SlotStr         string `json:"slot_str"`
	FeeStr          string `json:"fee_str"`
	ComputeUnit     string `json:"computer_unit"`
	Memo            string `json:"memo"`
	Type            string `json:"type"`
	TokenInName     string `json:"token_in_name"`
	TokenInSymbol   string `json:"token_in_symbol"`
	TokenInAmount   string `json:"token_in_amount"`
	TokenInAddress  string `json:"token_in_address"`
	TokenOutName    string `json:"token_out_name"`
	TokenOutSymbol  string `json:"token_out_symbol"`
	TokenOutAmount  string `json:"token_out_amount"`
	TokenOutAddress string `json:"token_out_address"`
	Signature       string `json:"signature"`
}
type Protocol struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type Action struct {
	Info           ActionInfo `json:"info"`
	SourceProtocol Protocol   `json:"source_protocol"`
	Type           string     `json:"type"`
	IxIndex        int        `json:"ix_index"`
}

type ActionInfo struct {
	Sender        string        `json:"sender,omitempty"`
	Receiver      string        `json:"receiver,omitempty"`
	Amount        float64       `json:"amount,omitempty"`
	AmountRaw     float64       `json:"amount_raw,omitempty"`
	Message       string        `json:"message,omitempty"`
	Swapper       string        `json:"swapper"`
	TokensSwapped TokensSwapped `json:"tokens_swapped"`
	Swaps         []Swap        `json:"swaps"`
}

type Event struct {
	// Define fields if events have specific structures
}

type RawData struct {
	BlockTime   int64          `json:"blockTime"`
	Meta        Meta           `json:"meta"`
	Slot        int            `json:"slot"`
	Transaction RawTransaction `json:"transaction"`
	Version     interface{}    `json:"version"`
}

type Meta struct {
	ComputeUnitsConsumed int           `json:"computeUnitsConsumed"`
	Err                  interface{}   `json:"err"`
	Fee                  float64       `json:"fee"`
	InnerInstructions    []interface{} `json:"innerInstructions"`
	LogMessages          []string      `json:"logMessages"`
	PostBalances         []int         `json:"postBalances"`
	PostTokenBalances    []interface{} `json:"postTokenBalances"`
	PreBalances          []int         `json:"preBalances"`
	PreTokenBalances     []interface{} `json:"preTokenBalances"`
	Rewards              []interface{} `json:"rewards"`
	Status               MetaStatus    `json:"status"`
}

type MetaStatus struct {
	Ok interface{} `json:"Ok"`
}

type RawTransaction struct {
	Message    RawMessage `json:"message"`
	Signatures []string   `json:"signatures"`
}

type RawMessage struct {
	AccountKeys         []AccountKey  `json:"accountKeys"`
	AddressTableLookups []interface{} `json:"addressTableLookups"`
	Instructions        []Instruction `json:"instructions"`
	RecentBlockhash     string        `json:"recentBlockhash"`
}

type AccountKey struct {
	Pubkey   string `json:"pubkey"`
	Signer   bool   `json:"signer"`
	Source   string `json:"source"`
	Writable bool   `json:"writable"`
}

type Instruction struct {
	Parsed      json.RawMessage `json:"parsed"`
	Program     string          `json:"program"`
	ProgramId   string          `json:"programId"`
	StackHeight interface{}     `json:"stackHeight"`
}

// 定义与 JSON 数据结构对应的 Go 结构体

type TokensSwapped struct {
	In  Token `json:"in"`
	Out Token `json:"out"`
}

type Token struct {
	TokenAddress string  `json:"token_address"`
	Name         string  `json:"name"`
	Symbol       string  `json:"symbol"`
	ImageURI     string  `json:"image_uri"`
	Amount       float64 `json:"amount"`
	AmountRaw    int64   `json:"amount_raw"`
}

type Swap struct {
	LiquidityPoolAddress string `json:"liquidity_pool_address"`
	Name                 string `json:"name"`
	Source               string `json:"source"`
	In                   Token  `json:"in"`
	Out                  Token  `json:"out"`
}

type SourceProtocol struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
