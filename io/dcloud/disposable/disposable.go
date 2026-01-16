package disposable

type Disposable interface {
	Dispose()
}

type Box[T Disposable] struct {
	value T
	delete bool
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

func (b *Box[T]) Delete() {
	b.delete = true
	b.value.Dispose()
}