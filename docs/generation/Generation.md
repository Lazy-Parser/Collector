# Generation. Theory 
This is a big and a little complicated part of the program. But if we will understand the base idea it will be super easy!

---

### 1. Why do we need to generate tokens \ pairs?
So lets first clarify, we do we need to make such a complicated thing?
The answer is that for listening prices of some pairs from **DEX** we need to know about each pair next info:
1. Address (contract). Its obvious that we need to know the address of some pair to listen its price. This data we get from Mexc or other exchange
2.  xNetwork + pool (`Ethereum + Pancakeswap V3` for example). We also need this info to **specify a specific** pair. Because just `Address` is not 100% unique for ALL existing pairs. But the combination of address and network + pool provides precise identification.
3. Decimal (8, 18 or other): its just a number. But very important one! Because of the specific inner architecture of smart contracts and blockchain, all pools store not human like prices. So to calculate unclear numbers from pools to the human format, we need to use Decimals.
   By the way, all pools have different formulas of calculating price. 
   For example, the formula from Uniswap V3:
   (image here)
4. What pair ( `Quote Token` in particular)  better to select for some token? (Because at the start we have only list of tokens. Not the pairs)
5. Which pairs are on futures of the exchange (Mexc in our case)
6. Also Filter pairs / tokens with not supported in the application networks. Filter by volume

---

### 2. Theory
Now we know that we need a lot of data to just listen pairs data from pool!
Let's watch my realization:
![Theory](https://raw.githubusercontent.com/Lazy-Parser/Collector/refs/heads/core-achitecture/docs/img/Generator.png)

As you can see, the whole process is just to make requests by chain. In each step we get more and more useful info.

---
### 3. Realization
Ok, so now we know what we need to fetch from different APIs. 
And for this i have the next:
