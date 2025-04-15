# 🧠 Collector

**Collector** is a real-time cryptocurrency market data aggregator written in Go. It connects to various exchanges such as **MEXC Spot & Futures**, and (soon) **DexScreener**, normalizes data, and prepares it for publishing via NATS.

---

## 🚀 Features

- 📡 WebSocket connections to multiple crypto exchanges
- 🔁 Aggregates prices across sources (Spot vs Futures, etc.)
- 🧠 Internal caching & deduplication
- ⏱ Graceful shutdown with context support
- 🧹 Auto-cleans stale or incomplete data
- 🔄 Ready for pluggable publishers (NATS, etc.)
- ✅ Clean modular structure

---

## Ensure your .env file contains:
1. MEXC_SPOT_WS=wss://wbs.mexc.com/ws
1. MEXC_FUTURES_WS=wss://contract.mexc.com/ws

---

## 🛠 Tech Stack
- Language: Go (1.21+)
- Communication: WebSocket. Protobuf in future
- Message broker: NATS

--- 
## 📌 TODO / Roadmap
- [x] MEXC Spot & Futures support
- [x] NATS publisher
- [ ] DexScreener integration
- [ ] Tests
