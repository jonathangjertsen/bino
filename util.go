package main

import "strconv"

func MapMap[K comparable, VIn any, VOut any](in map[K]VIn, f func(VIn) (VOut, error)) (map[K]VOut, error) {
	out := make(map[K]VOut)
	for k, vin := range in {
		vout, err := f(vin)
		if err != nil {
			return nil, err
		}
		out[k] = vout
	}
	return out, nil
}

func IDSlice(in []string) ([]int32, error) {
	out := make([]int32, len(in))
	for i, s := range in {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, err
		}
		out[i] = int32(v)
	}
	return out, nil
}

func MapSlice[TIn any, TOut any](in []TIn, f func(TIn) TOut) []TOut {
	out := make([]TOut, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func FilterSlice[TIn any](in []TIn, f func(v TIn) bool) []TIn {
	out := make([]TIn, 0, len(in))
	for _, v := range in {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}

func Find[TIn any](in []TIn, f func(v TIn) bool) int {
	for i, v := range in {
		if f(v) {
			return i
		}
	}
	return -1
}

func MoveToFront[T any](in []T, j int) {
	for i := j; i > 0; i-- {
		in[i], in[i-1] = in[i-1], in[i]
	}
}
