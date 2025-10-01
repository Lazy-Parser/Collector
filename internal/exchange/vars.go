package exchange_internal

// "compress":true
var (
	futures_sub = `{
    "method":"sub.depth",
    "param":{
        "symbol": %s,
        "limit":5,
    }
}`
	futures_unsub = `{
   "method":"unsub.depth",
   "param":{
       "symbol": %s,
    }
}`

	spot_sub = `{
    "method": "SUBSCRIPTION",
    "params": [%s]
}`
	spot_unsub = `{
    "method": "UNSUBSCRIPTION",
    "params": [%s]
}`

	pingMsg   = `{"method": "PING"}`
	volumeMin = 100_000.0

	futuresPingMsg = `{"method": "ping"}`
)
