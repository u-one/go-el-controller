package echonetlite

import (
	"fmt"
	"log"
)

// Object is object
type Object struct {
	ClassGroup ClassGroupCode
	Class      ClassCode
	Num        int
}

// NewObject returns Object
func NewObject(classGroup ClassGroupCode, class ClassCode, instance int) Object {
	return Object{classGroup, class, instance}
}

// NewObjectFromData returns Object
func NewObjectFromData(d Data) Object {
	if len(d) < 3 {
		log.Println("NewObjectFromData: invalid data")
		return Object{}
	}
	return Object{ClassGroupCode(d[0]), ClassCode(d[1]), int(d[2])}
}

func (o Object) classGroupCode() ClassGroupCode {
	return o.ClassGroup
}

func (o Object) classCode() ClassCode {
	return o.Class
}

func (o Object) isNodeProfile() bool {
	if o.ClassGroup == ProfileGroup &&
		o.Class == Profile {
		return true
	}
	return false
}

func (o Object) String() string {
	return fmt.Sprintf("%02x %02x %02x", o.ClassGroup, o.Class, o.Num)
}

func (o Object) Data() []byte {
	return []byte{byte(o.ClassGroup), byte(o.Class), byte(o.Num)}
}
