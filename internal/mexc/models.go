package mexc

type ConfType string

const (
	Futures ConfType = "FUTURES"
	Spot ConfType = "SPOT"
)

type MexcConf struct {
	Type		ConfType
	URL         string
	Subscribe   map[string]interface{}
	Unsubscribe map[string]interface{}
	ParseFunc 	func([]byte)  
}