# Dcard Homework - Rate Limiter Middleware
![Build Status](https://travis-ci.com/minghsu0107/dcard-middleware.svg?branch=main)
## Description
Dcard 每天午夜都有大量使用者湧入抽卡，為了不讓伺服器過載，請設計一個 middleware：

- 限制每小時來自同一個 IP 的請求數量不得超過 1000
- 在 response headers 中加入剩餘的請求數量 (X-RateLimit-Remaining) 以及 rate limit 歸零的時間 (X-RateLimit-Reset)
- 如果超過限制的話就回傳 429 (Too Many Requests)
- 可以使用各種資料庫達成
## Idea
因申請條件需要我們熟悉 Golang 或是 Node.js，而我本身又是 Golang 的愛好者，因此這次作業直接選擇 Golang 作為開發語言。

此使用場景很適合使用 Redis 這類 in-memory 的 key-value store。實作上我以 client IP 當作 key 並記錄個別 IP 的請求總數。同時，我們可以利用 Redis 的 TTL 機制維護 rate limit 的歸零時間。另外，我使用了 gin 這個 web framework 構建 API、運用工廠模式與 dig 實現依賴注入，並撰寫了單元測試。

與 Redis 的互動我使用了 interface 以解耦對資料庫的依賴。這樣的優點除了便於撰寫 mock test，在將來若要換別種資料庫也相對容易，只要實作此介面即可。
## Usage
Start application:
```
sh run.sh
```
Run test:
```
cd app && go test ./...
```
