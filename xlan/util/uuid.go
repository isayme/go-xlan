package util

import (
	gonanoid "github.com/matoous/go-nanoid"
)

func UUID() string {
	id, err := gonanoid.Nanoid()
	if err != nil {
		panic(err)
	}
	return id
}
