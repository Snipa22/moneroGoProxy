package support

type Share struct {
	Nonce       [4]byte // Nonce is 4 bytes, or 8 hex characters long.
	Result      string  // Result is a 64 character string
	WorkerNonce [4]byte // Worker nonce is also 4 bytes long
}
