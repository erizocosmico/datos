package datos

import (
	"testing"
	"time"
)

func TestDatetime(t *testing.T) {
	var d Datetime
	err := d.UnmarshalJSON([]byte(`"dom, 18 nov 2012 23:00:00 GMT+0000"`))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	expected := time.Date(2012, time.November, 18, 23, 0, 0, 0, time.UTC)
	if !d.Equal(expected) {
		t.Errorf("invalid date, expected: %s, got: %s", expected, d)
	}
}
