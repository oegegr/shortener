package main

// generate:reset
type ResetableStruct struct {
	I     int
	Str   string
	StrP  *string
	S     []int
	M     map[string]string
	Child *ResetableStruct
}

// generate:reset
type AnotherStruct struct {
	Flag    bool
	Numbers []float64
	Data    map[int]string
	Nested  *ResetableStruct
}
