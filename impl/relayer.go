package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/kasplex-evm/kasplex-relayer/log"
)

type Config struct {
	EthRPC     string
	KasRPC     string
	Port       int
	PrivateKey string
	ToAddress  string
}

type Relayer struct {
	cfg    *Config
	server *http.Server
	wallet *Wallet
}

type JsonRPCRequest struct {
	Id     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type RawTransaction struct {
	Type                 string  `json:"type"`
	ChainID              string  `json:"chainId"`
	Nonce                string  `json:"nonce"`
	To                   string  `json:"to"`
	Gas                  string  `json:"gas"`
	GasPrice             string  `json:"gasPrice"`
	MaxPriorityFeePerGas *string `json:"maxPriorityFeePerGas"` // null 允许为指针
	MaxFeePerGas         *string `json:"maxFeePerGas"`
	Value                string  `json:"value"`
	Input                string  `json:"input"`
	V                    string  `json:"v"`
	R                    string  `json:"r"`
	S                    string  `json:"s"`
	Hash                 string  `json:"hash"`
}

func NewRelayer(cfg *Config) (*Relayer, error) {
	wallet, err := NewWallet(cfg.PrivateKey, cfg.KasRPC)
	if err != nil {
		return nil, err
	}

	return &Relayer{
		cfg:    cfg,
		wallet: wallet,
	}, nil
}

func (r *Relayer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.dispatch)

	r.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", r.cfg.Port),
		Handler: mux,
	}

	go func() {
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Infof("HTTP server error: %v", err)
		}
	}()
	log.Infof("Relayer started on port %d", r.cfg.Port)

	return nil
}

func (r *Relayer) Stop() error {
	if r.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return r.server.Shutdown(ctx)
	}
	return nil
}

func (r *Relayer) dispatch(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(body))

	var jsonRPCRequest JsonRPCRequest
	if err := json.Unmarshal(body, &jsonRPCRequest); err != nil {
		http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	switch jsonRPCRequest.Method {
	case "eth_sendRawTransaction":
		r.handleSendRawTransaction(w, req, &jsonRPCRequest)
	default:
		r.handleDefault(w, req)
	}
}

func (r *Relayer) handleDefault(w http.ResponseWriter, req *http.Request) {
	proxyReq, err := http.NewRequest(http.MethodPost, r.cfg.EthRPC, req.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func (r *Relayer) handleSendRawTransaction(w http.ResponseWriter, req *http.Request, jsonRPCRequest *JsonRPCRequest) {
	reqId := jsonRPCRequest.Id

	writeError := func(errStr string, code int) {
		w.Write([]byte(fmt.Sprintf(`{
			"jsonrpc": "2.0",
			"id": %d,
			"error": {
				"code": %d,
				"message": "%s"
			}
		}`, reqId, code, errStr)))
	}

	if len(jsonRPCRequest.Params) == 0 {
		writeError("data error", -1)
		return
	}

	var vmData []byte
	var err error
	isJson := false
	hash := ""
	switch jsonRPCRequest.Params[0].(type) {
	case string:
		vmData, err = decodeToHex(jsonRPCRequest.Params[0].(string))
		if err != nil {
			log.Infof("decodeToHex(jsonBytes) error:", err)
			writeError(err.Error(), -2)
			return
		}
		hash = fmt.Sprintf("0x%x", keccak256(vmData))
	case map[string]interface{}:
		vmData, err = json.Marshal(jsonRPCRequest.Params[0])
		if err != nil {
			log.Infof("json.Marshal(data[0]) error:", err)
			writeError(err.Error(), -3)
			return
		}
		hash = jsonRPCRequest.Params[0].(map[string]interface{})["hash"].(string)
		isJson = true
	default:
		log.Infof("unknown type:", reflect.TypeOf(jsonRPCRequest.Params[0]))
	}

	_, err = r.wallet.TransferVM(r.cfg.ToAddress, "30", vmData, isJson)
	if err != nil {
		writeError(err.Error(), -3)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": %d,
		"result": "%s"
	}`, reqId, hash)))
}
