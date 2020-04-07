// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"errors"
	"strconv"
)

///
/// Uint16
///

type Uint16 uint16

func (u Uint16) ImplementsGraphQLType(name string) bool {
	return name == "Uint16"
}

func (u *Uint16) Native() uint16 {
	if u == nil {
		return 0
	}
	return uint16(*u)
}

func (u *Uint16) UnmarshalGraphQL(input interface{}) error {
	// FIXME: do bound checks, ensure it fits within the Uint32
	// boundaries before truncating silently.
	var err error
	switch input := input.(type) {
	case float64:
		*u = Uint16(input)
	case float32:
		*u = Uint16(input)
	case int64:
		*u = Uint16(input)
	case uint64:
		*u = Uint16(input)
	case uint32:
		*u = Uint16(input)
	case int32:
		*u = Uint16(input)
	case uint16:
		*u = Uint16(input)
	case int16:
		*u = Uint16(input)
	case string:
		res, err2 := strconv.ParseUint(input, 10, 16)
		if err2 != nil {
			err = err2
			break
		}
		*u = Uint16(res)
	default:
		err = errors.New("wrong type")
	}
	return err
}

func (u Uint16) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(u.Native()), 10)), nil
}

///
/// Uint32
///

type Uint32 uint32

func (u Uint32) ImplementsGraphQLType(name string) bool {
	return name == "Uint32"
}

func (u *Uint32) Native() uint32 {
	if u == nil {
		return 0
	}
	return uint32(*u)
}

func (u *Uint32) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case float64:
		// FIXME: do bound checks, ensure it fits within the Uint32
		// boundaries before truncating silently.
		*u = Uint32(input)
	case float32:
		*u = Uint32(input)
	case int64:
		*u = Uint32(input)
	case uint64:
		*u = Uint32(input)
	case uint32:
		*u = Uint32(input)
	case int32:
		*u = Uint32(input)
	case string:
		res, err2 := strconv.ParseUint(input, 10, 32)
		if err2 != nil {
			err = err2
			break
		}
		*u = Uint32(res)
	default:
		err = errors.New("wrong type")
	}
	return err
}

func (u Uint32) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(u.Native()), 10)), nil
}
