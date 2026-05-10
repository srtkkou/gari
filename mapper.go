package gari

type (
	MapperFunc func(*Record)
)

func DefaultMapper(ptr any) MapperFunc {
	return func(r *Record) {
		/*
			m, err := structPtrToMap(ptr)
			if err != nil {
				return
			}
			for name, value :=
		*/
	}
}
