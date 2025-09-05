package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"
)

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
		defer conn.Close()

		var subscriptionID = "0x1"

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Error("read error:", err)
				return
			}

			var req jsonRPCReq
			if err := json.Unmarshal(msg, &req); err != nil {
				log.Error("unmarshal error:", err)
				continue
			}

			if req.Method == "eth_subscribe" {
				resp := jsonRPCResp{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result:  subscriptionID,
				}
				conn.WriteJSON(resp)

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

						// JSON-RPC subscription event
						event := map[string]interface{}{
							"jsonrpc": "2.0",
							"method":  "eth_subscription",
							"params": map[string]interface{}{
								"subscription": subscriptionID,
								"result":       header,
							},
						}

						if err := conn.WriteJSON(event); err != nil {
							log.Error("write error:", err)
							return
						}
						log.Info(fmt.Sprintf("mock: sent new head %d\n", header.Number))
					}
				}()
			}
			if req.Method == "eth_getLogs" {
				// todo: here, maybe read the block number to return it
				fakeLog := map[string]interface{}{
					"address":          "0x1111111111111111111111111111111111111111",
					"blockHash":        "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"blockNumber":      "0x3e9", // 1001 în hex
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
					Result:  []interface{}{fakeLog},
				}
				conn.WriteJSON(resp)
				continue
			}

		}

	})

	log.Info("mock Ethereum WS server listening on :8546")
	err := http.ListenAndServe(":8546", nil)
	log.LogIfError(err)

	return nil
}
