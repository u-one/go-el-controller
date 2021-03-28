package echonetlite

// Object is object
type Object struct {
	Data Data
}

// NewObject returns Object
func NewObject(classGroup ClassGroupCode, class ClassCode, instance int) Object {
	return Object{Data: Data{byte(classGroup), byte(class), byte(instance)}}
}

// NewObjectFromData returns Object
func NewObjectFromData(d Data) Object {
	return Object{Data: d}
}

func (o Object) classGroupCode() ClassGroupCode {
	return ClassGroupCode(o.Data[0])
}

func (o Object) classCode() ClassCode {
	return ClassCode(o.Data[1])
}

func (o Object) isNodeProfile() bool {
	if o.Data[0] == byte(ProfileGroup) &&
		o.Data[1] == byte(Profile) {
		return true
	}
	return false
}
