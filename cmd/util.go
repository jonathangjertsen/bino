package main

import (
	"fmt"
	"strconv"
)

func MapToMapErr[K comparable, VIn any, VOut any](in map[K]VIn, f func(VIn) (VOut, error)) (map[K]VOut, error) {
	if in == nil {
		return nil, fmt.Errorf("called MapToMapErr on nil map")
	}
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
	if in == nil {
		return nil, fmt.Errorf("called IDSlice on nil slice")
	}
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

func SliceToMap[TIn any, KOut comparable, VOut any](in []TIn, f func(TIn) (KOut, VOut)) map[KOut]VOut {
	if in == nil {
		return nil
	}
	out := make(map[KOut]VOut)
	for _, vin := range in {
		k, vout := f(vin)
		out[k] = vout
	}
	return out
}

func SliceToSlice[TIn any, TOut any](in []TIn, f func(TIn) TOut) []TOut {
	if in == nil {
		return nil
	}
	out := make([]TOut, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func SliceToSliceErr[TIn any, TOut any](in []TIn, f func(TIn) (TOut, error)) ([]TOut, error) {
	if in == nil {
		return nil, fmt.Errorf("called SliceToSliceErr on nil slice")
	}
	out := make([]TOut, len(in))
	for i, v := range in {
		vOut, err := f(v)
		if err != nil {
			return nil, err
		}
		out[i] = vOut
	}
	return out, nil
}

func SliceToMapErr[TIn any, KOut comparable, VOut any](in []TIn, f func(int, TIn) (KOut, VOut, error)) (map[KOut]VOut, error) {
	if in == nil {
		return nil, fmt.Errorf("called SliceToMapErr on nil slice")
	}
	out := make(map[KOut]VOut)
	for i, vin := range in {
		k, vout, err := f(i, vin)
		if err != nil {
			return nil, err
		}
		out[k] = vout
	}
	return out, nil
}

func FilterSlice[TIn any](in []TIn, f func(v TIn) bool) []TIn {
	if in == nil {
		return nil
	}
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

func Any[T any](in []T, f func(in T) bool) bool {
	for _, x := range in {
		if f(x) {
			return true
		}
	}
	return false
}
