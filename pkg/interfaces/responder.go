package interfaces

type responder interface {
	ListenFor(these ...string) ResponseBuilder
	ListenForExactly(these ...string) ResponseBuilder
	RespondTo(line string)
}

type ResponseBuilder interface {
	WithAnyOf(...string) ResponseBuilder
	WithAllOf(...string) ResponseBuilder
	WithExactlyOneOf(...string) ResponseBuilder
	WithNoneOf(...string) ResponseBuilder
	Ignore(...string) ResponseBuilder
	Chance(float64) ResponseBuilder
	RespondWithOneOf(...string)
}
