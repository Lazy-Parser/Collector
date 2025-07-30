package server_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func SetupServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/mexc/config/getall", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(configGetAllResp))
	})
	mux.HandleFunc("/mexc/contract/detail", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		resp, _ := http.Get("https://contract.mexc.com/api/v1/contract/detail")
		defer resp.Body.Close()

		raw, _ := io.ReadAll(resp.Body)

		w.Write(raw)
	})
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "Hello there!\n")
		//...
	})

	return httptest.NewServer(mux)
}

const configGetAllResp = "[{\"coin\":\"SCR\",\"name\":\"Scroll\",\"networkList\":[{\"coin\":\"SCR\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":60,\"name\":\"Scroll\",\"network\":\"SCROLL\",\"withdrawEnable\":true,\"withdrawFee\":\"0.1\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"200000\",\"withdrawMin\":\"0.5\",\"sameAddress\":false,\"contract\":\"0xd29687c813D741E2F938F4aC377128810E217b1b\",\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"SCROLL\"}]},{\"coin\":\"BBF\",\"name\":\"BubblefongToken\",\"networkList\":[{\"coin\":\"BBF\",\"depositDesc\":\"Depositisprohibited(forceddepositmayresultinassetloss)\",\"depositEnable\":false,\"minConfirm\":96,\"name\":\"BubblefongToken\",\"network\":\"Ethereum(ERC20)\",\"withdrawEnable\":true,\"withdrawFee\":\"0\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"150000\",\"withdrawMin\":\"38568\",\"sameAddress\":false,\"contract\":\"0xDE075D9ADbD0240b4462F124af926452Ad0BAC91\",\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"ETH\"}]},{\"coin\":\"TRUMP\",\"name\":\"OFFICIALTRUMP\",\"networkList\":[{\"coin\":\"TRUMP\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":100,\"name\":\"OFFICIALTRUMP\",\"network\":\"Solana(SOL)\",\"withdrawEnable\":true,\"withdrawFee\":\"0.1476\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"50000\",\"withdrawMin\":\"0.2951\",\"sameAddress\":false,\"contract\":\"6p6xgHyF7AeE6TZkSmFsko444wqoP15icUSqi2jfGiPN\",\"withdrawTips\":\"\",\"depositTips\":\"\",\"netWork\":\"SOL\"}]},{\"coin\":\"SRC\",\"name\":\"SafeRoadClub\",\"networkList\":[{\"coin\":\"SRC\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":100,\"name\":\"SafeRoadClub\",\"network\":\"Solana(SOL)\",\"withdrawEnable\":true,\"withdrawFee\":\"1088\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"200000000\",\"withdrawMin\":\"2176\",\"sameAddress\":false,\"contract\":\"5gkd3yk3WmTEjYxiTsCXk3p8uPd9W85L5UNYNUQaheyb\",\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"SOL\"}]},{\"coin\":\"FARTCOIN\",\"name\":\"FARTCOIN\",\"networkList\":[{\"coin\":\"FARTCOIN\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":100,\"name\":\"FARTCOIN\",\"network\":\"Solana(SOL)\",\"withdrawEnable\":true,\"withdrawFee\":\"1.13\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"10000000\",\"withdrawMin\":\"2.25\",\"sameAddress\":false,\"contract\":\"9BB6NFEcjBCtnNLFko2FqVQBq8HHM13kCyYcdQbgpump\",\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"SOL\"}]},{\"coin\":\"GASS\",\"name\":\"Gasspas\",\"networkList\":[{\"coin\":\"GASS\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":96,\"name\":\"Gasspas\",\"network\":\"Ethereum(ERC20)\",\"withdrawEnable\":true,\"withdrawFee\":\"407563829\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"7000000000000\",\"withdrawMin\":\"1416645877.2\",\"sameAddress\":false,\"contract\":\"0x774eaF7A53471628768dc679dA945847d34b9a55\",\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"ETH\"}]},{\"coin\":\"XRPAYNET\",\"name\":\"XRPaynet\",\"networkList\":[{\"coin\":\"XRPAYNET\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":30,\"name\":\"XRPaynet\",\"network\":\"Ripple(XRP)\",\"withdrawEnable\":true,\"withdrawFee\":\"10\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"2860000000\",\"withdrawMin\":\"20\",\"sameAddress\":true,\"contract\":\"58525061794E6574000000000000000000000000\",\"withdrawTips\":\"AsXRPAYNETrunsonXRPchain,it'smandatoryforuserstoactivateXRPAYNETontheaddress.Otherwise,thewithdrawalwon'tbecredited.\",\"depositTips\":null,\"netWork\":\"XRP\"}]},{\"coin\":\"NEER\",\"name\":\"MNetPioneer\",\"networkList\":[{\"coin\":\"NEER\",\"depositDesc\":\"Depositsuspendedduetowalletmaintenance\",\"depositEnable\":false,\"minConfirm\":10,\"name\":\"MNetPioneer\",\"network\":\"NEER\",\"withdrawEnable\":false,\"withdrawFee\":\"2\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"660000\",\"withdrawMin\":\"5\",\"sameAddress\":false,\"contract\":null,\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"NEER\"}]},{\"coin\":\"SGT\",\"name\":\"AIAvatar\",\"networkList\":[{\"coin\":\"SGT\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":96,\"name\":\"AIAvatar\",\"network\":\"Ethereum(ERC20)\",\"withdrawEnable\":true,\"withdrawFee\":\"17\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"200000\",\"withdrawMin\":\"39.6\",\"sameAddress\":false,\"contract\":\"0x5b649C07E7Ba0a1C529DEAabEd0b47699919B4a2\",\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"ETH\"}]},{\"coin\":\"CELR\",\"name\":\"CelerNetwork\",\"networkList\":[{\"coin\":\"CELR\",\"depositDesc\":null,\"depositEnable\":true,\"minConfirm\":96,\"name\":\"CelerNetwork\",\"network\":\"Ethereum(ERC20)\",\"withdrawEnable\":true,\"withdrawFee\":\"289\",\"withdrawIntegerMultiple\":null,\"withdrawMax\":\"8000000\",\"withdrawMin\":\"577.2\",\"sameAddress\":false,\"contract\":\"0x4F9254C83EB525f9FCf346490bbb3ed28a81C667\",\"withdrawTips\":null,\"depositTips\":null,\"netWork\":\"ETH\"}]}]"