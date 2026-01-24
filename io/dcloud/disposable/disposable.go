package disposable

import "sync"

type Disposable interface {
	Dispose()
}

type Box[T Disposable] struct {
	value T
	delete bool
	mu sync.Mutex
}

func NewBox[T Disposable](value T) *Box[T] {
	return &Box[T]{value: value}
}

func (b *Box[T]) Value() T {
	if b.delete {
		var zero T
		return zero
	}
	return b.value
}

func (b *Box[T]) Set(value T){
	b.mu.Lock()
	defer b.mu.Unlock()

	b.value.Dispose()
	b.value = value
}

func (b *Box[T]) Delete() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.delete = true
	b.value.Dispose()
}