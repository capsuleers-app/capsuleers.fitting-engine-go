package fitengine_test

import (
	"bytes"
	"testing"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/bundle"
)

// TestEncodeJSONNulls pins the `encode:"null"` tag: a nil field so tagged is emitted as null (like the
// TS `null`), an untagged nil optional is omitted (TS `undefined`).
func TestEncodeJSONNulls(t *testing.T) {
	res, err := bundle.ComputeFromEft("[Rifter, enc]\n200mm AutoCannon II, Republic Fleet EMP S\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	d := res.Computed.Derived
	if d.Structure != nil {
		t.Fatal("a ship fit should carry no structure block")
	}
	out, err := fitengine.EncodeJSON(d)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte(`"structure":null`)) {
		t.Errorf("nil Structure (encode:\"null\") must encode as null: %.300s", out)
	}
	if bytes.Contains(out, []byte(`"secondsToEmpty":null`)) {
		t.Errorf("untagged nil optional must be omitted, not null")
	}
}
