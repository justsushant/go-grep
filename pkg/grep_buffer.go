package grep

type GrepBuffer struct {
	buff []string
	size int
}

func NewGrepBuffer(size int) GrepBuffer {
	if size == 0 {
		return GrepBuffer{}
	}
	return GrepBuffer{
		buff: make([]string, 0),
		size: size,
	}
}

func(b *GrepBuffer) Push(data string) {
	if len(b.buff) == b.size {
		b.buff = b.buff[1:]
	}
	b.buff = append(b.buff, data)
}

func(b *GrepBuffer) Dump() []string {
	return b.buff
}