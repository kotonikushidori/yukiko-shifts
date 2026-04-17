// push.go — Web Push 送信ロジック (VAPID)
package push

import (
	"encoding/json"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/yourorg/shift-app/internal/model"
)

// Payload はプッシュ通知の JSON ペイロード
type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

// Sender は VAPID キーを保持し、通知を送信する
type Sender struct {
	privateKey string
	publicKey  string
}

// NewSender は VAPID 秘密鍵・公開鍵を受け取り Sender を返す。
// どちらかが空文字のとき nil を返す（プッシュ無効状態）。
func NewSender(privateKey, publicKey string) *Sender {
	if privateKey == "" || publicKey == "" {
		return nil
	}
	return &Sender{privateKey: privateKey, publicKey: publicKey}
}

// PublicKey は VAPID 公開鍵を返す（フロントエンドへの配布用）
func (s *Sender) PublicKey() string { return s.publicKey }

// Send は 1 件のサブスクリプションに通知を送る
func (s *Sender) Send(sub *model.PushSubscription, title, body, url string) error {
	payload, _ := json.Marshal(Payload{Title: title, Body: body, URL: url})

	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}, &webpush.Options{
		VAPIDPrivateKey: s.privateKey,
		VAPIDPublicKey:  s.publicKey,
		Subscriber:      "mailto:admin@shift-app.local",
		TTL:             86400,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SendAll は複数サブスクリプションに通知を送る（エラーはログのみ）
func (s *Sender) SendAll(subs []model.PushSubscription, title, body, url string) {
	for i := range subs {
		if err := s.Send(&subs[i], title, body, url); err != nil {
			ep := subs[i].Endpoint
			if len(ep) > 40 {
				ep = ep[:40] + "…"
			}
			log.Printf("push: send error (endpoint=%s): %v", ep, err)
		}
	}
}
