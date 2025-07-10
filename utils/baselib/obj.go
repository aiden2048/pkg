package baselib

import "reflect"

// 利用反射机制New一个跟v类型相同的对象
func NewTypeOf(v interface{}) interface{} {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Ptr:
		// v是个指针，Elem()取到指针指向的对象，Type()取到对象的类型，New一个对象，返回对象指针，Interface()转成interface{}
		return reflect.New(reflect.ValueOf(v).Elem().Type()).Interface()
	default:
		// v是个对象，Type()取到对象的类型，New一个对象，返回对象指针，Elem()取到指针指向的对象，Interface()转成interface{}
		return reflect.New(reflect.ValueOf(v).Type()).Elem().Interface()
	}
}

// 利用反射机制：
// 1、New一个跟v类型相同的对象的指针（当v为对象）
// 2、New一个跟v类型相同的指针（当v为指针）
func NewPtrOfType(v interface{}) interface{} {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Ptr:
		// v是个指针，Elem()取到指针指向的对象，Type()取到对象的类型，New一个对象，返回对象指针，Interface()转成interface{}
		return reflect.New(reflect.ValueOf(v).Elem().Type()).Interface()
	default:
		// v是个对象，Type()取到对象的类型，New一个对象，返回对象指针，Interface()转成interface{}
		return reflect.New(reflect.ValueOf(v).Type()).Interface()
	}
}

func IsPtr(v interface{}) bool {
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}
