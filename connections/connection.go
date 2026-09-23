package connections

type Connection interface {
	Read(n int) ([]byte, error)
	ReadExactly(n int) ([]byte, error)
	Write(data []byte) error
	Close()
}
