package greeting

type Greeting struct {
	Message string
}

func New() Greeting {
	return Greeting{Message: "hello"}
}
