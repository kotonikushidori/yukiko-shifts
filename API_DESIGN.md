# API設計書 — 施工会社シフト管理システム

## 認証方式
- JWT（HS256）/ Authorizationヘッダー: `Bearer <token>`
- トークン有効期限: 24時間

---

## エンドポイント一覧

### 認証
| Method | Path | 説明 | 権限 |
|--------|------|------|------|
| POST | /api/auth/login | ログイン → JWT発行 | 不要 |

### シフト確認・設定
| Method | Path | 説明 | 権限 |
|--------|------|------|------|
| GET | /api/shifts/board?from=&to= | 全体シフトボード（日付範囲） | 両方 |
| GET | /api/shifts/my?from=&to= | 自分のシフト一覧 | worker |
| POST | /api/sites/{siteID}/shifts/{date}/assign | アサイン追加 ★衝突チェック付き | admin |
| DELETE | /api/shifts/assign/{id} | アサイン削除 | admin |

### 日報・月報
| Method | Path | 説明 | 権限 |
|--------|------|------|------|
| PUT | /api/reports/{date} | 日報登録・更新（YYYY-MM-DD） | worker |
| GET | /api/reports/my?year=&month= | 自分の月次日報一覧 | worker |
| POST | /api/reports/submit?year=&month= | 月報提出確定 | worker |
| GET | /api/reports/summary?year=&month= | 全作業者月次サマリ | admin |

### 現場マスタ（実装Phase 1）
| Method | Path | 説明 | 権限 |
|--------|------|------|------|
| GET | /api/sites | 現場一覧 | 両方 |
| GET | /api/sites/{id} | 現場詳細 | 両方 |
| POST | /api/sites | 現場新規登録 | admin |
| PUT | /api/sites/{id} | 現場更新 | admin |

### 作業者マスタ（実装Phase 1）
| Method | Path | 説明 | 権限 |
|--------|------|------|------|
| GET | /api/workers | 作業者一覧 | admin |
| POST | /api/workers | 作業者登録 | admin |
| PUT | /api/workers/{id} | 作業者更新 | admin |

---

## 二重アサインバリデーション仕様

```
POST /api/sites/{siteID}/shifts/{date}/assign
Body: { "user_id": 3, "time_slot": "AM" }
```

### 衝突判定ロジック
| 既存スロット | 追加スロット | 結果 |
|------------|------------|------|
| ALL | AM / PM / ALL | 409 Conflict |
| AM | AM / ALL | 409 Conflict |
| AM | PM | 201 Created ✅ |
| PM | PM / ALL | 409 Conflict |
| PM | AM | 201 Created ✅ |

### エラーレスポンス例
```json
HTTP 409 Conflict
{
  "error": "田中さんは5/25のAMに「トヨタ分室」へ配置済みのためALLは追加できません"
}
```

---

## 月報提出フロー

```
作業者                        サーバー
  |                              |
  |-- PUT /api/reports/2025-05-01 --> 日報保存（毎日）
  |-- PUT /api/reports/2025-05-02 --> 日報保存
  |         ...                  |
  |-- GET /api/reports/my?year=2025&month=5
  |<-- { reports: [...], missing: ["2025-05-10"] }
  |                              |
  |-- POST /api/reports/submit?year=2025&month=5
  |<-- 200 OK（submitted_at記録）  |
```

---

## 環境変数

| 変数名 | デフォルト | 説明 |
|--------|----------|------|
| JWT_SECRET | (要変更) | JWTシークレットキー（本番は32文字以上） |
| DB_PATH | ./shift.db | SQLiteファイルパス |
| PORT | 8080 | サーバーポート |
