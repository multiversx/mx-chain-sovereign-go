package main

import (
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"
)

const (
	ethMethodSubscribe = "eth_subscribe"
	ethMethodGetLogs   = "eth_getLogs"
)

type ethMethodHandler func(req jsonRPCReq, conn *websocket.Conn) error

var upgrader = websocket.Upgrader{}

type jsonRPCReq struct {
	ID     int64           `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type jsonRPCResp struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func startETHMockNotifier() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("upgrade error:", err)
			return
		}
		defer func(conn *websocket.Conn) {
			err = conn.Close()
			log.LogIfError(err)
		}(conn)

		handlers := map[string]ethMethodHandler{
			ethMethodSubscribe: handleBlockSubscribe,
			ethMethodGetLogs:   handleGetLogs,
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Error("read error:", err)
				continue
			}

			var req jsonRPCReq
			if err := json.Unmarshal(msg, &req); err != nil {
				log.Error("unmarshal error:", err)
				continue
			}

			requestHandler, found := handlers[req.Method]
			if !found {
				log.Error("unknown ETH request method:", req.Method)
				continue
			}

			err = requestHandler(req, conn)
			if err != nil {
				log.Error("requestHandler failed", "error", err, "req type", req.Method)
			}
		}

	})

	log.Info("mock Ethereum WS server listening on :8546")
	err := http.ListenAndServe(":8546", nil)
	log.LogIfError(err)

	return nil
}

func handleBlockSubscribe(req jsonRPCReq, conn *websocket.Conn) error {
	subscriptionID := "0x1"

	resp := jsonRPCResp{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  subscriptionID,
	}
	err := conn.WriteJSON(resp)
	if err != nil {
		log.Error("handleBlockSubscribe conn.WriteJSON", "error", err)
		return err
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		blockNumber := big.NewInt(1000)

		for range ticker.C {
			blockNumber = big.NewInt(blockNumber.Int64() + 1)
			header := &types.Header{
				Number:     blockNumber,
				Difficulty: big.NewInt(1),
			}

			event := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "eth_subscription",
				"params": map[string]interface{}{
					"subscription": subscriptionID,
					"result":       header,
				},
			}

			if err := conn.WriteJSON(event); err != nil {
				log.Error("handleBlockSubscribe write error:", "error", err)
				return
			}
			log.Info("sending ETH block", "number", header.Number.Uint64())
		}
	}()

	return nil
}

func handleGetLogs(req jsonRPCReq, conn *websocket.Conn) error {
	var reqParams []struct {
		FromBlock string `json:"fromBlock"`
	}
	if err := json.Unmarshal(req.Params, &reqParams); err != nil {
		return err
	}

	incomingLog := map[string]interface{}{
		"address":          "0x1111111111111111111111111111111111111111",
		"blockHash":        "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"blockNumber":      reqParams[0].FromBlock,
		"transactionHash":  "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"transactionIndex": "0x0",
		"logIndex":         "0x0",
		"removed":          false,
		"data":             "0xdeadbeef",
		"topics": []string{
			"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	resp := jsonRPCResp{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  []interface{}{incomingLog},
	}

	// TODO: For now have this empty, since we do not have any incoming event processor handler for this
	resp.Result = []interface{}{}
	return conn.WriteJSON(resp)
}
