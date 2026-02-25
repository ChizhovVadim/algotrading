package trader

type Optional[T any] struct {
	Value    T
	HasValue bool
}

func (m *Optional[T]) SetValue(value T) {
	m.Value = value
	m.HasValue = true
}
