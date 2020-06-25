package pool

const (
	// Reduce Memory for iOS Devices
	RelayBufferSize = 10 * 1024
)

func Get(size int) []byte {
	return defaultAllocator.Get(size)
}

func Put(buf []byte) error {
	return defaultAllocator.Put(buf)
}
