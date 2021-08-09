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

package apollo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/streamingfast/dauth/authenticator"
)

type messageIntention int

const (
	clientSends messageIntention = 0
	expectation messageIntention = 1
)

const (
	connectionACK = `{"type":"connection_ack"}`
)

type message struct {
	intention        messageIntention
	operationMessage string
}

func TestConnect(t *testing.T) {
	testTable := []struct {
		name     string
		svc      *gqlService
		messages []message
	}{
		{
			name: "connection_init_ok",
			messages: []message{
				{
					intention: clientSends,
					operationMessage: `{
						"type":"connection_init",
						"payload":{}
					}`,
				},
				{
					intention:        expectation,
					operationMessage: connectionACK,
				},
			},
		},
		{
			name: "connection_init_error",
			messages: []message{
				{
					intention: clientSends,
					operationMessage: `{
						"type": "connection_init",
						"payload": "invalid_payload"
					}`,
				},
				{
					intention: expectation,
					operationMessage: `{
						"type": "connection_error",
						"payload": {
							"message": "invalid payload for type: connection_init"
						}
					}`,
				},
			},
		},
		{
			name: "start_ok",
			svc:  newGQLService(`{"data":{},"errors":null}`),
			messages: []message{
				{
					intention: clientSends,
					operationMessage: `{
						"type": "start",
						"id": "a-id",
						"payload": {}
					}`,
				},
				{
					intention: expectation,
					operationMessage: `{
						"type": "data",
						"id": "a-id",
						"payload": {
							"data": {},
							"errors": null
						}
					}`,
				},
				{
					intention: expectation,
					operationMessage: `{
						"type":"complete",
						"id": "a-id"
					}`,
				},
			},
		},
		{
			name: "start_query_data_error",
			svc:  newGQLService(`{"data":null,"errors":[{"message":"a error"}]}`),
			messages: []message{
				{
					intention: clientSends,
					// this payload should fail?
					operationMessage: `{
						"id": "a-id",
						"type": "start",
						"payload": {}
					}`,
				},
				{
					intention: expectation,
					operationMessage: `{
						"id": "a-id",
						"type": "data",
						"payload": {
							"data": null,
							"errors": [{"message":"a error"}]
						}
					}`,
				},
				{
					intention: expectation,
					operationMessage: `{
						"type":"complete",
						"id": "a-id"
					}`,
				},
			},
		},
		{
			name: "start_query_error",
			svc: &gqlService{
				err: errors.New("some error"),
			},
			messages: []message{
				{
					intention: clientSends,
					operationMessage: `{
						"id": "a-id",
						"type": "start",
						"payload": {}
					}`,
				},
				{
					intention: expectation,
					operationMessage: `{
						"id": "a-id",
						"type": "error",
						"payload": {
							"message": "some error"
						}
					}`,
				},
			},
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ws := newConnection()
			go Connect("test", ws, tt.svc,
				Authentication(nil,
					func(ctx context.Context, r *http.Request, payload map[string]interface{}) (i context.Context, e error) {
						return authenticator.WithCredentials(ctx, &authenticator.AnonymousCredentials{}), nil
					},
				),
				FakeCredential(),
			)
			ws.test(t, tt.messages)
		})
	}
}

func FakeCredential() func(conn *connection) {
	return func(conn *connection) {
		conn.credentials = &authenticator.AnonymousCredentials{}
	}
}

type gqlService struct {
	payloads <-chan interface{}
	err      error
}

func newGQLService(pp ...string) *gqlService {
	c := make(chan interface{}, len(pp))
	for _, p := range pp {
		c <- json.RawMessage(p)
	}
	close(c)

	return &gqlService{payloads: c}
}

func (h *gqlService) Subscribe(ctx context.Context, document string, operationName string, variableValues map[string]interface{}) (payloads <-chan interface{}, err error) {
	return h.payloads, h.err
}

func newConnection() *testWSConnection {
	return &testWSConnection{
		in:  make(chan json.RawMessage),
		out: make(chan json.RawMessage),
	}
}

type testWSConnection struct {
	in  chan json.RawMessage
	out chan json.RawMessage
}

func (ws *testWSConnection) test(t *testing.T, messages []message) {
	for _, msg := range messages {
		switch msg.intention {
		case clientSends:
			ws.in <- json.RawMessage(msg.operationMessage)
		case expectation:
			select {
			case <-time.After(1 * time.Second):
				t.Errorf("no message received after 1s, expecting %s", msg.operationMessage)
			case v := <-ws.out:
				requireEqualJSON(t, msg.operationMessage, v)
			}
		}
	}
}

func (ws *testWSConnection) ReadJSON(v interface{}) error {
	msg := <-ws.in
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (ws *testWSConnection) WriteMessage(messageType int, data []byte) error {
	ws.out <- json.RawMessage(data)
	return nil
}

func (ws *testWSConnection) WriteJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ws.out <- json.RawMessage(data)
	return nil
}

func (ws *testWSConnection) SetReadLimit(limit int64) {}

func (ws *testWSConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func (ws *testWSConnection) Close() error {
	close(ws.in)
	close(ws.out)

	return nil
}

func requireEqualJSON(t *testing.T, expected string, got json.RawMessage) {
	var expJSON interface{}
	err := json.Unmarshal([]byte(expected), &expJSON)
	if err != nil {
		t.Fatalf("error mashalling expected json: %s", err.Error())
	}

	var gotJSON interface{}
	err = json.Unmarshal(got, &gotJSON)
	if err != nil {
		t.Fatalf("error mashalling got json: %s", err.Error())
	}

	if !reflect.DeepEqual(expJSON, gotJSON) {
		normalizedExp, err := json.Marshal(expJSON)
		if err != nil {
			panic(err)
		}
		t.Fatalf("expected [%s] but instead got [%s]", normalizedExp, got)
	}
}
