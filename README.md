# Arbitrage Bot Core Library

This repository contains the **core Go library** for our **private arbitrage trading bot**.  
It is not a standalone bot — rather, it serves as the **engine** that collects and processes market data from various sources for further arbitrage analysis and execution.

---

## 📌 Purpose

The library’s main goal is to **fetch and unify market data** from all connected sources, providing a consistent interface for our trading logic to consume.

Currently supported sources:
- **DEX**: Ethereum (Uniswap & other supported protocols)  
- **CEX**: MEXC  

---

## 🚀 Roadmap

Planned upcoming integrations:
- **Networks**: Solana (first priority), followed by other EVM and non-EVM chains.
- **CEX**: Grass and additional centralized exchanges.

---

## ⚠️ Status

The project is **under active development**.  
Breaking changes may occur as we refine the architecture and add new integrations.
