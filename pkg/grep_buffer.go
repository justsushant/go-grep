package grep

type grepBuffer struct {
	buff []string
	size int
}

// buffer to hold the data
func NewGrepBuffer(size int) grepBuffer {
	if size == 0 {
		return grepBuffer{}
	}
	return grepBuffer{
		buff: make([]string, 0),
		size: size,
	}
}

// insert operation
func(b *grepBuffer) Push(data string) {
	if b.size == 0 {
		return
	}

	// if buffer is full, remove the oldest element
	if len(b.buff) == b.size {
		b.buff = b.buff[1:]
	}
	b.buff = append(b.buff, data)
}

// get operation
func(b *grepBuffer) Dump() []string {
	return b.buff
}
