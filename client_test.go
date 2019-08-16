package datos

import (
	"strings"
	"testing"
	"time"
)

var datasetParams = Params{PageSize: 10, Sort: "-issued"}

func TestDatasets(t *testing.T) {
	ds, err := newClient(t).Datasets(datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) < 10 {
		t.Errorf("wrong number of results, expected: 10, got: %d", len(ds))
	}
}

func TestDataset(t *testing.T) {
	t.Skip("endpoint does not work")
}

func TestDatasetsByTitle(t *testing.T) {
	ds, err := newClient(t).DatasetsByTitle("mirador", datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		var found bool
		for _, t := range d.Title {
			if strings.Contains(strings.ToLower(t), "mirador") {
				found = true
			}
		}

		if !found {
			t.Errorf(
				"expected title %q to contain %q",
				strings.Join([]string(d.Title), ", "),
				"mirador",
			)
		}
	}
}

func TestDatasetsByPublisher(t *testing.T) {
	pub := "L01280066"
	ds, err := newClient(t).DatasetsByPublisher(pub, datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		if !strings.HasSuffix(d.Publisher, "/"+pub) {
			t.Errorf("wrong publisher, expected: %s, got: %s", pub, d.Publisher)
		}
	}
}

func TestDatasetsByTheme(t *testing.T) {
	ds, err := newClient(t).DatasetsByTheme("sector-publico", datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		var found bool
		for _, t := range d.Theme {
			if strings.HasSuffix(t, "/sector-publico") {
				found = true
			}
		}

		if !found {
			t.Errorf(
				"expected theme %q to contain %q",
				strings.Join([]string(d.Theme), ", "),
				"sector-publico",
			)
		}
	}
}

func TestDatasetsByFormat(t *testing.T) {
	ds, err := newClient(t).DatasetsByFormat("csv", datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		var found bool
		var formats []string
		for _, dist := range d.Distribution {
			if dist.Format.Value == "text/csv" {
				found = true
			}
			formats = append(formats, dist.Format.Value)
		}

		if !found {
			t.Errorf("expected csv format, got: %s", strings.Join(formats, ", "))
		}
	}
}

func TestDatasetsByKeyword(t *testing.T) {
	ds, err := newClient(t).DatasetsByKeyword("turismo", datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		var found bool
		for _, t := range d.Keywords {
			if strings.Contains(strings.ToLower(t), "turismo") {
				found = true
			}
		}

		if !found {
			t.Errorf(
				"expected keywords %q to contain %q",
				strings.Join([]string(d.Keywords), ", "),
				"turismo",
			)
		}
	}
}

func TestDatasetsBySpatial(t *testing.T) {
	ds, err := newClient(t).DatasetsBySpatial(Autonomy, "Aragon", datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		var found bool
		for _, t := range d.Spatial {
			if strings.HasSuffix(t, "/Autonomia/Aragon") {
				found = true
			}
		}

		if !found {
			t.Errorf(
				"expected keywords %q to contain %q",
				strings.Join([]string(d.Spatial), ", "),
				"/Autonomia/Aragon",
			)
		}
	}
}

func TestDatasetsModifiedBetween(t *testing.T) {
	from := time.Date(2016, time.April, 18, 0, 0, 0, 0, time.UTC)
	to := time.Date(2016, time.June, 30, 0, 0, 0, 0, time.UTC)
	ds, err := newClient(t).DatasetsModifiedBetween(from, to, datasetParams)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		if !d.Modified.After(from) || !d.Modified.Before(to) {
			t.Errorf(
				"expected modified to be between %s and %s, got: %s",
				from, to,
				d.Modified,
			)
		}
	}
}

func TestDistributions(t *testing.T) {
	ds, err := newClient(t).Distributions(Params{PageSize: 10})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) < 10 {
		t.Errorf("wrong number of results, expected: 10, got: %d", len(ds))
	}
}

func TestDistributionsByFormat(t *testing.T) {
	ds, err := newClient(t).DistributionsByFormat("csv", Params{PageSize: 10})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("expecting results, got none")
	}

	for _, d := range ds {
		if d.Format.Value != "text/csv" {
			t.Errorf("unexpected format, expected: text/csv, got: %s", d.Format.Value)
		}
	}
}

func TestDistributionsByDataset(t *testing.T) {
	t.Skip("endpoint does not work")
}

func TestPublishers(t *testing.T) {
	ps, err := newClient(t).Publishers(Params{PageSize: 10})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ps) < 10 {
		t.Errorf("wrong number of results, expected: 10, got: %d", len(ps))
	}
}

func TestSpatials(t *testing.T) {
	ps, err := newClient(t).Spatials(Params{PageSize: 10})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ps) < 10 {
		t.Errorf("wrong number of results, expected: 10, got: %d", len(ps))
	}
}

func TestThemes(t *testing.T) {
	ps, err := newClient(t).Themes(Params{PageSize: 10})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ps) < 10 {
		t.Errorf("wrong number of results, expected: 10, got: %d", len(ps))
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
