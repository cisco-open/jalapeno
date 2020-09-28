package arangodb

// type StackItem interface {
// 	StackableItem()
// }

type FIFO interface {
	Push(DBRecord)
	Pop() DBRecord
	Len() int
}

type entry struct {
	next     *entry
	previous *entry
	data     DBRecord
}
type fifo struct {
	head *entry
	tail *entry
	len  int
}

func (f *fifo) Push(o DBRecord) {
	// Empty stack case
	e := &entry{
		next: f.tail,
		data: o,
	}
	if f.head == nil && f.tail == nil {
		f.head = e
	}
	f.tail = e
	if f.tail.next != nil {
		f.tail.next.previous = f.tail
	}
	f.len++
}

func (f *fifo) Pop() DBRecord {
	if f.head == nil {
		// Stack is empty
		return nil
	}
	data := f.head.data
	f.head = f.head.previous
	f.len--
	return data
}

func (f *fifo) Len() int {
	return f.len
}

func newFIFO() FIFO {
	return &fifo{
		head: nil,
		tail: nil,
	}
}
