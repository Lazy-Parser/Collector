package exchange_internal

var (
	futures_sub = `{
    "method":"sub.depth",
    "param":{
        "symbol":"%s",
        "limit":5,
        "compress":true
    }
}`

	futures_unsub = `{
   "method":"unsub.depth",
   "param":{
       "symbol":"%s",
   }
}`

	futuresPingMsg = `{"method": "ping"}`
)
