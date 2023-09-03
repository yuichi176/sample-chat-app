// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chat

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // メッセージの書き込みの待機時間（タイムアウト）
	pongWait       = 60 * time.Second    // ポング（接続確認メッセージ）の待機時間（タイムアウト）
	pingPeriod     = (pongWait * 9) / 10 // ピング（接続確認メッセージ）を定期的に送信する間隔
	maxMessageSize = 512                 // 受信可能なメッセージの最大サイズ
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // 読み込みバッファサイズ
	WriteBufferSize: 1024, // 書き込みバッファサイズ
}

type Client struct {
	hub  *Hub            // WebSocketハブへの参照
	conn *websocket.Conn // WebSocket接続
	send chan []byte     // クライアントへのメッセージ送信用チャネル
}

func newClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// クライアントからのメッセージ受信処理を行うメソッド
// WebSocket接続を介してクライアントからのメッセージを非同期に読み込む
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c // クライアントをWebSocketハブから登録解除
		c.conn.Close()        // WebSocket接続を閉じる
	}()
	c.conn.SetReadLimit(maxMessageSize)                                                                        // 最大メッセージサイズを設定
	c.conn.SetReadDeadline(time.Now().Add(pongWait))                                                           // ピンポンメッセージの期限を設定
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil }) // ピンポンメッセージハンドラを設定
	for {                                                                                                      // クライアントからのメッセージを非同期で読み込む
		_, message, err := c.conn.ReadMessage() // クライアントからメッセージを読み込む
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break // クライアントとの接続がエラーまたは閉じた場合、ループを終了
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1)) // メッセージを整形
		c.hub.broadcast <- message                                            // 受信したメッセージをWebSocketハブにブロードキャスト
	}
}

// クライアントへのメッセージ送信処理を行うメソッド
// WebSocket接続を介してクライアントへのメッセージ送信とピングメッセージの送信を非同期に処理
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod) // 定期的なピングメッセージを送信するタイマーを作成
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send: // クライアントがメッセージを受信した場合
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage) // WebSocketメッセージの書き込み用の io.WriteCloser を取得
			if err != nil {
				return
			}
			w.Write(message) // メッセージの書き込み

			// キューに溜まったメッセージを現在のWebSocketメッセージに追加
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send) // メッセージを書き込む
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C: // 定期的なピングメッセージを送信する場合
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WebSocket接続の処理を行う関数
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := newClient(hub, conn)
	client.hub.register <- client // クライアントの登録

	go client.writePump() // クライアントへのメッセージ送信を非同期に処理
	go client.readPump()  // クライアントからのメッセージ受信を非同期に処理
}
