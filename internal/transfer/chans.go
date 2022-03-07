package transfer

type Channels struct {
	ComChannel     chan string
	ErrChannel     chan string
	BoundedChannel chan bool
}
