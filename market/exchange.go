package market

type BufferTick struct {
	Token

	Withdraw    bool
	WithdrawFee string
	Deposit     bool

	Volume string
}

// A copy of [BufferTick] but with optional fields
type BufferTickUpdate struct {
	Token

	Withdraw    *bool
	WithdrawFee *string
	Deposit     *bool

	Volume *string
}
