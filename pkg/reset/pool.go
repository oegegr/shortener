// pool.go
package reset

import (
	"sync"
)

// Resetable ограничивает типы с методом Reset()
type Resetable interface {
	Reset()
}

// Pool представляет собой пул объектов с методом Reset()
type Pool[T Resetable] struct {
	pool *sync.Pool
}

func New[T Resetable]() *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				var t T
				t.Reset()
				return t
			},
		},
	}
}

// Get возвращает объект из пула (или создает новый)
func (p *Pool[T]) Get() T {
	obj := p.pool.Get()
	return obj.(T)
}

// Put помещает объект обратно в пул после сброса
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
