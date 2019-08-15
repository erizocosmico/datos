package datos

import "testing"

func TestDatasets(t *testing.T) {
	ds, err := newClient(t).Datasets(Params{PageSize: 10})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) < 10 {
		t.Errorf("wrong number of results, expected: 10, got: %d", len(ds))
	}
}

func newClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewClient()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	return c
}
