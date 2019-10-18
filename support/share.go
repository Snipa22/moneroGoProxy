package support

type Share struct {
	nonce       [4]byte // Nonce is 4 bytes, or 8 hex characters long.
	result      string  // Result is a 64 character string
	workerNonce [4]byte // Worker nonce is also 4 bytes long
}
