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

package dgraphql

import (
	"encoding/hex"
	"fmt"

	"github.com/dfuse-io/opaque"
	"github.com/golang/protobuf/proto"
)

func MustProtoToOpaqueCursor(entity proto.Message, entityTag string) string {
	data, err := proto.Marshal(entity)
	if err != nil {
		panic(fmt.Errorf("unable to marshal cursor proto %q: %s", entityTag, err))
	}

	result, err := opaque.ToOpaque(hex.EncodeToString(data))
	if err != nil {
		panic(fmt.Errorf("unable to opacify cursor: %s", err))
	}

	return result
}

func UnmarshalCursorProto(opaqued string, entity proto.Message) error {
	h, err := opaque.FromOpaque(opaqued)
	if err != nil {
		return fmt.Errorf("unable to decode opaque cursor %q: %s", opaqued, err)
	}

	data, err := hex.DecodeString(h)
	if err != nil {
		return fmt.Errorf("unable to decode cursor %q: %s", opaqued, err)
	}

	err = proto.Unmarshal(data, entity)
	if err != nil {
		return fmt.Errorf("invalid cursor: %s", err)
	}

	return nil
}
