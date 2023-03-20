package gospel

var Event = EventStruct{}

type EventStruct struct {
	Target TargetStruct
}

type TargetStruct struct {
	Value ValueStruct
}

type ValueStruct struct {
}
