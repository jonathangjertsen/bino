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
