package mock

type A struct {
	A string
	B struct {
		C int
		D []int
	}
}

type T struct {
	A string
	B struct {
		RenamedC int   `yaml:"c"`
		D        []int `yaml:",flow"`
	}
}

type V struct {
	A string
	B struct {
		RenamedC int
		D        []int
	}
}
