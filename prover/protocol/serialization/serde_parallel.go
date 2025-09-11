package serialization

import (
	"fmt"
	"reflect"
	"sync"
)

/*
func unpackAllActions[T any](d *Deserializer, context string,
	out [][]T, actions2D [][]PackedRawData) *serdeError {

	var (
		res   reflect.Value
		se    *serdeError
		ct    reflect.Type
		ctStr string
		err   error
	)

	for round, actions := range actions2D {
		if len(actions) == 0 {
			out[round] = nil
			continue
		}

		out[round] = make([]T, len(actions))

		for idx, action := range actions {
			ctStr = d.PackedObject.Types[action.ConcreteType]
			ct, err = findRegisteredImplementation(ctStr)
			if err != nil {
				return newSerdeErrorf("could not find registered implementation for %s action: %w", context, err)
			}

			switch ct.Kind() {
			case reflect.Struct:
				res, se = d.UnpackStructObject(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath(fmt.Sprintf("(deser struct compiled-IOP-%s-actions)", context))
				}
				var valInterface any
				if action.WasPointer {
					valInterface = res.Addr().Interface()
				} else {
					valInterface = res.Interface()
				}

				v, ok := valInterface.(T)
				if !ok {
					return newSerdeErrorf("illegal cast of type %v with string rep %s to %s action", ct, ctStr, context)
				}
				out[round][idx] = v

			case reflect.Slice, reflect.Array:
				res, se = d.UnpackArrayOrSlice(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath(fmt.Sprintf("(deser slice/array compiled-IOP-%s-actions)", context))
				}

				var valInterface any
				if action.WasPointer {
					ptr := reflect.New(res.Type())
					ptr.Elem().Set(res)
					valInterface = ptr.Interface()
				} else {
					valInterface = res.Interface()
				}

				v, ok := valInterface.(T)
				if !ok {
					return newSerdeErrorf("illegal cast of %v with string rep %s to %s action", ct, ctStr, context)
				}
				out[round][idx] = v

			default:
				return newSerdeErrorf("unsupported kind:%v for %s action", ct.Kind(), context)
			}
		}
	}
	return nil
} */

func unpackAllActions[T any](d *Deserializer, context string,
	out [][]T, actions2D [][]PackedRawData) *serdeError {

	for round, actions := range actions2D {
		if len(actions) == 0 {
			out[round] = nil
			continue
		}
		out[round] = make([]T, len(actions))

		var wg sync.WaitGroup
		errCh := make(chan *serdeError, 1) // capture first error

		for idx, action := range actions {
			wg.Add(1)
			go func(idx int, action PackedRawData) {
				defer wg.Done()

				// These must be local to avoid races
				ctStr := d.PackedObject.Types[action.ConcreteType]
				ct, err := findRegisteredImplementation(ctStr)
				if err != nil {
					select {
					case errCh <- newSerdeErrorf(
						"could not find registered implementation for %s action: %w",
						context, err,
					):
					default:
					}
					return
				}

				var (
					res          reflect.Value
					se           *serdeError
					valInterface any
				)

				switch ct.Kind() {
				case reflect.Struct:
					res, se = d.UnpackStructObject(action.ConcreteValue, ct)
					if se != nil {
						select {
						case errCh <- se.wrapPath(fmt.Sprintf("(deser struct compiled-IOP-%s-actions)", context)):
						default:
						}
						return
					}
					if action.WasPointer {
						valInterface = res.Addr().Interface()
					} else {
						valInterface = res.Interface()
					}

				case reflect.Slice, reflect.Array:
					res, se = d.UnpackArrayOrSlice(action.ConcreteValue, ct)
					if se != nil {
						select {
						case errCh <- se.wrapPath(fmt.Sprintf("(deser slice/array compiled-IOP-%s-actions)", context)):
						default:
						}
						return
					}
					if action.WasPointer {
						ptr := reflect.New(res.Type())
						ptr.Elem().Set(res)
						valInterface = ptr.Interface()
					} else {
						valInterface = res.Interface()
					}

				default:
					select {
					case errCh <- newSerdeErrorf("unsupported kind:%v for %s action", ct.Kind(), context):
					default:
					}
					return
				}

				v, ok := valInterface.(T)
				if !ok {
					select {
					case errCh <- newSerdeErrorf(
						"illegal cast of type %v with string rep %s to %s action",
						ct, ctStr, context,
					):
					default:
					}
					return
				}

				out[round][idx] = v
			}(idx, action)
		}

		wg.Wait()

		// check if any goroutine reported error
		select {
		case e := <-errCh:
			return e
		default:
		}
	}

	return nil
}
