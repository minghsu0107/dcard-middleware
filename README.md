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

此應用場景很適合使用 Redis 這類 in-memory 的 key-value store。實作上我以 client IP 當作 key 並記錄個別 IP 的請求總數，並我們可以利用 Redis 的 TTL 機制維護 rate limit 的歸零時間。另外，我使用了 gin 這個 web framework 構建 API 與 middleware、運用工廠模式與 wire 實現依賴注入、撰寫了單元測試並串接到 CI 工具 (使用 TravisCI)。

與 Redis 的互動我使用了 interface 以解耦對資料庫的依賴。這樣的優點除了便於撰寫 mock test，在將來若要換別種資料庫也相對容易，只要實作此介面即可。
## Redis Configuration
如果我們的記憶體最大容量只有 5G，但是卻寫了 10G 的資料怎麼辦？這時就需要 Redis 的淘汰機制去刪除不需要的資料。Redis 採用的淘汰機制是定期刪除 + 懶惰刪除。當一個 key 過期時，Redis 其實不會馬上刪除它。Redis 預設每 100 ms「隨機」抽樣 key 出來檢查，若這個 key 過期才刪除它。然而若只有隨機抽樣刪除可能造成許多過期的 key 沒有被刪除，進而佔用記憶體空間。因此 Redis 還使用了懶惰刪除，當我們在獲取某個 key 時，Redis 會先去檢查它是否過期，如果過期就會將其刪除。但即使使用了定期刪除 + 懶惰刪除還是不夠，若定期刪除沒刪除到 key，我們也沒有去訪問那個 key，這樣就會讓記憶體佔用率越來越高。因此，Redis 還設計了一個很重要的參數：`maxmemory-policy`。

`maxmemory-policy` 這個參數規範了當 memory 滿的時候 Redis 該如何選擇淘汰的 key。在此應用場景我們可以使用 `volatile-lru`，此設定會淘汰最近最少使用 (Least-Recently-Used, LRU) 並且 TTL 已經 expire 的 key。使用 `volatile-lru` 的好處是我們只會丟棄過期的 key 而不會影響其他仍尚未過期的資料。`volatile-lru` 為 `maxmemory-policy` 的預設值，但我認為仍值得拿出來討論。

順帶一提，若我們只是把 Redis 當作緩存而非像此場景使用 Redis 儲存 volatile data，我們可以使用 `llkeys-lru`，這個參數會從整個 keyset 中丟棄最近最少使用的 key，包含那些尚未過期的資料，讓 Redis 在淘汰 key 時有更大的彈性。
## Usage
啟動服務：
```
sh run.sh
```
這個 script 會 export 所有需要的環境變數並啟動容器。

執行測試：
```
cd app && go test ./...
```
