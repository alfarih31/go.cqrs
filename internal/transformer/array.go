package transformer

type StaticArrayTransformer[S any, T any] func(data []S) ([]T, error)
type ArrayTransformerMapper[S any, T any] func(each S) (T, error)

func ArrayTransformer[S any, T any](source []S, mapper ArrayTransformerMapper[S, T]) ([]T, error) {
	o := make([]T, len(source))

	for i, e := range source {
		d, err := mapper(e)
		if err != nil {
			return nil, err
		}

		o[i] = d
	}

	return o, nil
}

func NewStaticArrayTransformer[S any, T any](mapper ArrayTransformerMapper[S, T]) StaticArrayTransformer[S, T] {
	return func(data []S) ([]T, error) {
		return ArrayTransformer[S, T](data, mapper)
	}
}

func ArrayUniqBy[T any, TI string | int](arr []T, IDExtractor func(a T) TI) []T {
	arrMap := map[TI]bool{}
	o := []T{}
	for _, a := range arr {
		id := IDExtractor(a)
		if _, exist := arrMap[id]; exist {
			continue
		}
		arrMap[id] = true
		o = append(o, a)
	}

	return o
}
