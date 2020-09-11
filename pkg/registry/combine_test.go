// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package registry

import (
	"errors"
	"testing"

	"github.com/ozonep/drone/pkg/drone"
)

func TestCombine(t *testing.T) {
	a := &drone.Registry{}
	b := &drone.Registry{}
	aa := mockProvider{out: []*drone.Registry{a}}
	bb := mockProvider{out: []*drone.Registry{b}}
	p := Combine(&aa, &bb)
	out, err := p.List(noContext, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if len(out) != 2 {
		t.Errorf("Expect combined registry output")
		return
	}
	if out[0] != a {
		t.Errorf("Unexpected registry at index 0")
	}
	if out[1] != b {
		t.Errorf("Unexpected registry at index 1")
	}
}

func TestCombineError(t *testing.T) {
	e := errors.New("not found")
	m := mockProvider{err: e}
	p := Combine(&m)
	_, err := p.List(noContext, nil)
	if err != e {
		t.Errorf("Expect error")
	}
}
