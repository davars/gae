// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rawdatastore

import (
	"golang.org/x/net/context"
)

type key int

var (
	rawDatastoreKey       key
	rawDatastoreFilterKey key = 1
)

// Factory is the function signature for factory methods compatible with
// SetFactory.
type Factory func(context.Context) Interface

// Filter is the function signature for a filter RDS implementation. It
// gets the current RDS implementation, and returns a new RDS implementation
// backed by the one passed in.
type Filter func(context.Context, Interface) Interface

// GetUnfiltered gets gets the Interface implementation from context without
// any of the filters applied.
func GetUnfiltered(c context.Context) Interface {
	if f, ok := c.Value(rawDatastoreKey).(Factory); ok && f != nil {
		return f(c)
	}
	return nil
}

// Get gets the Interface implementation from context.
func Get(c context.Context) Interface {
	ret := GetUnfiltered(c)
	if ret == nil {
		return nil
	}
	for _, f := range getCurFilters(c) {
		ret = f(c, ret)
	}
	return ret
}

// SetFactory sets the function to produce Datastore instances, as returned by
// the Get method.
func SetFactory(c context.Context, rdsf Factory) context.Context {
	return context.WithValue(c, rawDatastoreKey, rdsf)
}

// Set sets the current Datastore object in the context. Useful for testing with
// a quick mock. This is just a shorthand SetFactory invocation to set a factory
// which always returns the same object.
func Set(c context.Context, rds Interface) context.Context {
	return SetFactory(c, func(context.Context) Interface { return rds })
}

func getCurFilters(c context.Context) []Filter {
	curFiltsI := c.Value(rawDatastoreFilterKey)
	if curFiltsI != nil {
		return curFiltsI.([]Filter)
	}
	return nil
}

// AddFilters adds Interface filters to the context.
func AddFilters(c context.Context, filts ...Filter) context.Context {
	if len(filts) == 0 {
		return c
	}
	cur := getCurFilters(c)
	newFilts := make([]Filter, 0, len(cur)+len(filts))
	newFilts = append(newFilts, getCurFilters(c)...)
	newFilts = append(newFilts, filts...)
	return context.WithValue(c, rawDatastoreFilterKey, newFilts)
}