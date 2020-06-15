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
	"encoding/json"
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

///
/// Int64
///

type Int64 int64

func ToInt64(rawJSON string) Int64 {
	var res Int64
	_ = res.UnmarshalJSON([]byte(rawJSON))
	return res
}

func (u Int64) ImplementsGraphQLType(name string) bool {
	return name == "Int64"
}

func (u *Int64) Native() int64 {
	if u == nil {
		return 0
	}
	return int64(*u)
}

func (u *Int64) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		res, err2 := strconv.ParseInt(input, 10, 64)
		if err2 != nil {
			err = err2
			break
		}
		*u = Int64(res)
	case float64:
		// FIXME: do bound checks, ensure it fits within the Int64
		// boundaries before truncating silently.
		*u = Int64(input)
	case float32:
		*u = Int64(input)
	case int64:
		*u = Int64(input)
	case uint64:
		*u = Int64(input)
	case uint32:
		*u = Int64(input)
	case int32:
		*u = Int64(input)
	default:
		err = errors.New("wrong type")
	}
	return err
}

func (u Int64) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(u.Native(), 10) + `"`), nil
}

func (i *Int64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty value")
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}

		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		*i = Int64(val)

		return nil
	}

	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*i = Int64(v)

	return nil
}

///
/// Uint64
///

type Uint64 uint64

// ToUint64 does JSON decoding of an uint64, and returns a Uint64 from this package.
func ToUint64(rawJSON string) Uint64 {
	var res Uint64
	_ = res.UnmarshalJSON([]byte(rawJSON))
	return res
}

func (u Uint64) ImplementsGraphQLType(name string) bool {
	return name == "Uint64"
}

func (u *Uint64) Native() uint64 {
	if u == nil {
		return 0
	}
	return uint64(*u)
}

func (u *Uint64) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		res, err2 := strconv.ParseUint(input, 10, 64)
		if err2 != nil {
			err = err2
			break
		}
		*u = Uint64(res)
	case float64:
		// FIXME: do bound checks, ensure it fits within the Uint64
		// boundaries before truncating silently.
		*u = Uint64(input)
	case float32:
		*u = Uint64(input)
	case int64:
		*u = Uint64(input)
	case uint64:
		*u = Uint64(input)
	case uint32:
		*u = Uint64(input)
	case int32:
		*u = Uint64(input)
	default:
		err = errors.New("wrong type")
	}
	return err
}

func (u Uint64) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatUint(u.Native(), 10) + `"`), nil
}

func (i *Uint64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty value")
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}

		val, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}

		*i = Uint64(val)

		return nil
	}

	var v uint64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*i = Uint64(v)

	return nil
}
