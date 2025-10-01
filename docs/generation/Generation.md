# Generation. Theory  
This part of the bot is both **critical and a bit tricky**. It may look complicated at first, but once you understand the core idea, everything becomes much easier. Essentially, generation is the foundation that allows the bot to know which pairs to monitor, how to interpret their prices, and how to filter out irrelevant data.  

---

### 1. Why do we need to generate tokens and pairs?  
Before diving into the technical details, let’s clarify: **why do we even need this step?**  

When we want to track prices of trading pairs on a **DEX** (Decentralized Exchange), the bot must have detailed information about each pair. Without this metadata, prices can’t be interpreted correctly. Here’s what we need to know for every pair:  

1. **Address (Contract Address)**  
   Each pool on the blockchain has a smart contract address. To fetch its data, we must know this address.  
   Example: On Uniswap, the pair `ETH/USDC` has a unique contract address.  

2. **Network + Pool (e.g., `Ethereum + Uniswap V3`)**  
   The contract address alone is **not always globally unique**. Different networks (Ethereum, BNB Chain, Polygon, etc.) and different pools (Uniswap, PancakeSwap, SushiSwap) can contain pools with similar addresses.  
   - That’s why we need the **combination of address + network + pool** to precisely identify a pair.  

3. **Decimals (e.g., 8, 18, etc.)**  
   This number defines how raw blockchain values should be converted into human-readable numbers.  

   #### ❓ Why Decimals Exist?  
   One important thing to understand: **smart contracts cannot store floating-point numbers**.  

   On blockchains (Ethereum, BNB Chain, etc.) all arithmetic is done with **integers only**. This is because floating-point math is:  
   - **imprecise** (rounding errors),  
   - **expensive** to implement in smart contracts,  
   - and could lead to security issues in financial applications.  

   So instead of floats, contracts use **decimals** as a way to represent fractional values.  

   **Example:**  
   Let’s say a token has `decimals = 18`.  
   - The smart contract stores balances in **wei** (the smallest indivisible unit).  
   - If your balance is `1.000000000000000000` ETH, the contract actually stores it as the integer:\
     `1000000000000000000` `(= 1 * 10^18)`\
     To convert this into a human-readable number, you divide by `10^decimals`.
     
     By the way, all pools have different formulas of calculating price.  
     For example, the formula from Uniswap V3:
     
     ![Formula](https://github.com/Lazy-Parser/Collector/blob/core-achitecture/docs/img/Decimals%20example.png?raw=true)


4. **Base vs. Quote Token Selection**  
   In arbitrage, it’s important to choose the correct **quote token** for comparison.  
   - Example: If you want to know the value of a random token `XYZ`, should you track `XYZ/USDT`, `XYZ/USDC`, or `XYZ/ETH`?  
   - The generation process must decide which pair is the most reliable.  

5. **Futures Compatibility**  
   Since our arbitrage bot also considers futures trading (MEXC in this case), we must link tokens/pairs from spot markets with their corresponding futures pairs.  

6. **Filtering**  
   Not every pair is useful. We need to filter out:  
   - Pairs from unsupported networks  
   - Pairs with very low liquidity or volume  
   - Tokens with no futures mapping
  
---

### 2. Theory — How Generation Works  
So, in short, we need **a lot of metadata** just to start listening to pair prices.  
That’s where the **generation pipeline** comes into play.  

The process is simple in concept:  
1. Start with a **list of tokens on futures** (e.g., from MEXC or another exchange).  
2. Step by step, **enrich this data** with more information by calling external APIs (Dexscreener, Coingecko).  
3. After each step, the dataset becomes more structured and ready for usage by the trading logic.  

Here is a diagram of the process:  
![Theory](https://raw.githubusercontent.com/Lazy-Parser/Collector/refs/heads/core-achitecture/docs/img/Generator.png)  

Think of it as a **chain of transformations**:  
- At the beginning → we only know token symbols, contract (address) and network.  
- At the end → we have fully described pairs with address (!Pair, not just token), decimals, quote tokens, and volume filters applied.  

---

### 3. Practice - Technical realization
Now we know all the theory, and we need to write some code.
Of course, this core repo has instruments for this:
 1. **`MexcWorker`**
    -  `GetAllTokens()`  - fetch all existing tokens that are on Mexc. Useful because this request has a lot of info for each pair. (Contract, network, withdraw, deposit, symbol, ...).\
       But there are some tokens that exist only on spot, not on futures. So how to know wich token is on futures?\
       
       > **Note!**\
       > There is some **problem** here: in mexc api we get reponse in this format:\
       > `[ {coin: 'symbol', networks: [ { network, contract, ... }, ... ] }, ... ]`\
       > As you can see, it can be more than one network for each coin (token). For now, we pick the first one. But its wrong.
       > **In future updates I will correct this problem!**
       
    -  `GetAllFutures()` - fetch all tokens that are on futures. This request has very little info. But it contains `symbol` field, and thats enough to 'pick' tokens from previous method.
   
    -  `FindContractBySymbol()`. A method to select tokens from `GetAllTokens()` by `GetAllFutures()` symbols.

   
   2. **`DexscreenerWorker`** - a little complicated part, but it contains just one public method that try to find the best pair from provided token.\
      Just to clarify: in this worker you need to pass only tokens from MexcWorker/

      **For now, we have all needed info, accept of `Decimals`**


   4. **`CoingeckoWroker`**:
      - `CreateChunks()` - a method that 'group' provided tokens from `DexscreenerWorker` by networks.\
        We do this because Coingecko API provides ability to pass multiple tokens (up to 30), that have the same network.

      - `CreateChunks()` - just fetch info about tokens from provided chunk. In this part we can get decimals and save it.


Thats all!

---

### 4. Additionally
1. You do not really need to know about this service, because its for an internal work.\
   But as you need to create it manually and pass to every worker, it will be better to know what does it do.
   
   **Problem**: Mexc, Coingecko and Dexscreener have different network names for the particular network name.
   For example: eth - Mexc, Dexscreener - ethereum, Coingecko - ETH.
   
   But in each step (from part #3) we need to provide network names and tecnically those network names are different, despite the fact that they mean the same thing.
   
   **Solution**: the solution is to write custom `chains` service, that will store all netoworks that are allowed, in all possible types.\
   In 'under the hood' every worker change provided network name to the custom.


2. `Worker`s are just a little bit of logic over `api`s. `api` its also just a service that contains ONLY request methods. And Worker  just change input data for convenient API operation. They are also have some logic.


      










   
   
