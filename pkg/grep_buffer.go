package grep

type grepBuffer struct {
	buff []string
	size int
}

func NewGrepBuffer(size int) grepBuffer {
	if size == 0 {
		return grepBuffer{}
	}
	return grepBuffer{
		buff: make([]string, 0),
		size: size,
	}
}

func(b *grepBuffer) Push(data string) {
	if len(b.buff) == b.size {
		b.buff = b.buff[1:]
	}
	b.buff = append(b.buff, data)
}

func(b *grepBuffer) Dump() []string {
	return b.buff
}