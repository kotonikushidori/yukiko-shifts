// cmd/genkeys/main.go — VAPID キーペアを生成して標準出力に出力する
// 使い方: docker run --rm -v $(pwd):/app -w /app golang:1.22-alpine go run ./cmd/genkeys
package main

import (
	"fmt"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatalf("VAPID キー生成エラー: %v", err)
	}
	fmt.Println("=== VAPID Keys ===")
	fmt.Printf("VAPID_PUBLIC_KEY=%s\n", publicKey)
	fmt.Printf("VAPID_PRIVATE_KEY=%s\n", privateKey)
	fmt.Println()
	fmt.Println("上記の値を Render の Environment Variables に設定してください。")
}
