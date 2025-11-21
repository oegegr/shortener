// pool_test.go
package reset

import (
	"sync"
	"testing"
)

// TestResetableStruct - тестовая структура с методом Reset() для указателя
// generate:reset
type TestResetableStruct struct {
	ID    int
	Name  string
	Data  []byte
	Items []string
}

// AnotherResetableStruct - другая тестовая структура с Reset() для указателя
// generate:reset
type AnotherResetableStruct struct {
	Counter int
	Active  bool
	Values  []float64
	Config  map[string]int
}

func TestPoolPutResetsObject(t *testing.T) {
	pool := New[*TestResetableStruct]()

	// Получаем объект и изменяем его состояние
	obj := pool.Get()
	if obj == nil {
		obj = &TestResetableStruct{}
	}

	obj.ID = 100
	obj.Name = "test"
	obj.Data = []byte{1, 2, 3}
	obj.Items = []string{"a", "b", "c"}

	// Проверяем, что изменения применены
	if obj.ID != 100 {
		t.Errorf("Before Put: Expected ID to be 100, got %d", obj.ID)
	}

	// Возвращаем объект в пул (автоматически сбрасывается)
	pool.Put(obj)

	// Проверяем, что объект был сброшен
	if obj.ID != 0 {
		t.Errorf("After Put: Expected ID to be 0, got %d", obj.ID)
	}
	if obj.Name != "" {
		t.Errorf("After Put: Expected Name to be empty, got %s", obj.Name)
	}
	if len(obj.Data) != 0 {
		t.Errorf("After Put: Expected Data to be empty, got %v", obj.Data)
	}
	if len(obj.Items) != 0 {
		t.Errorf("After Put: Expected Items to be empty, got %v", obj.Items)
	}
}

func TestPoolGetAfterPut(t *testing.T) {
	pool := New[*TestResetableStruct]()

	// Получаем объект, изменяем и возвращаем
	obj1 := pool.Get()
	if obj1 == nil {
		obj1 = &TestResetableStruct{}
	}

	obj1.ID = 42
	obj1.Name = "original"
	pool.Put(obj1)

	// Получаем объект снова
	obj2 := pool.Get()

	// Проверяем, что объект сброшен
	if obj2.ID != 0 {
		t.Errorf("Expected object to be reset, got ID: %d", obj2.ID)
	}
	if obj2.Name != "" {
		t.Errorf("Expected object to be reset, got Name: %s", obj2.Name)
	}
}

func TestPoolReuse(t *testing.T) {
	pool := New[*TestResetableStruct]()

	// Получаем объект и запоминаем указатель
	obj1 := pool.Get()
	if obj1 == nil {
		obj1 = &TestResetableStruct{}
	}

	originalPointer := obj1

	// Изменяем и возвращаем
	obj1.ID = 1
	pool.Put(obj1)

	// Получаем снова - может быть тот же объект
	obj2 := pool.Get()

	// Проверяем, что это может быть тот же указатель
	if obj2 != originalPointer {
		t.Log("Pool returned different object (this is OK)")
	}

	// Проверяем, что объект сброшен
	if obj2.ID != 0 {
		t.Error("Reused object was not properly reset")
	}
}

func TestPoolConcurrentAccess(t *testing.T) {
	pool := New[*TestResetableStruct]()
	var wg sync.WaitGroup

	// Запускаем несколько горутин
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			obj := pool.Get()
			if obj == nil {
				obj = &TestResetableStruct{}
			}

			obj.ID = id
			obj.Name = "goroutine"
			pool.Put(obj)
		}(i)
	}

	wg.Wait()

	// Проверяем, что пул все еще работает
	obj := pool.Get()
	if obj.ID != 0 {
		t.Errorf("Expected object to be reset after concurrent access, got ID: %d", obj.ID)
	}
}

func TestPoolNilSafety(t *testing.T) {
	pool := New[*TestResetableStruct]()

	// Многократные операции не должны вызывать панику
	for i := 0; i < 1000; i++ {
		obj := pool.Get()
		pool.Put(obj)
	}

	t.Log("Completed 1000 Get/Put operations without panic")
}
