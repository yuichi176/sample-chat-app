// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chat

type Hub struct {
	clients    map[*Client]bool // 登録済みの Client
	broadcast  chan []byte      // Client から受信したメッセージをブロードキャストするためのチャネル
	register   chan *Client     // Client の登録リクエストを受け付けるチャネル
	unregister chan *Client     // Client の登録解除リクエストを受け付けるチャネル
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for { // 非同期処理制御用の無限ループ
		select { // 複数のチャネル操作を待機し、最初に受信可能な操作を実行
		// Client が登録リクエストを送信した場合、クライアントを登録する
		case client := <-h.register:
			h.clients[client] = true
		// Client が登録解除リクエストを送信した場合、クライアントを登録解除し、関連するチャネルを閉じる
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		// Client からのメッセージを受信し、登録された全てのクライアントにブロードキャストする
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				// メッセージをクライアントに送信する
				case client.send <- message:
				// クライアントの送信チャネルがブロックしている場合、クライアントを登録解除し、チャネルを閉じる
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
