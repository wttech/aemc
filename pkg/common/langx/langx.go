package langx

import (
	"reflect"
)

// ChannelToSlice grabs all elements from channel and puts into slice (see: https://stackoverflow.com/a/20396392)
func ChannelToSlice(ch any) any {
	chv := reflect.ValueOf(ch)
	slv := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(ch).Elem()), 0, 0)
	for {
		v, ok := chv.Recv()
		if !ok {
			return slv.Interface()
		}
		slv = reflect.Append(slv, v)
	}
}
