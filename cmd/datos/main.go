package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/erizocosmico/datos"
	"github.com/sirupsen/logrus"
)

var verbose bool

func main() {
	var title, keyword, theme, publisher, format, output string
	var num uint

	flag.StringVar(&title, "title", "", "filter by title")
	flag.StringVar(&keyword, "keyword", "", "filter by keyword")
	flag.StringVar(&theme, "theme", "", "filter by theme")
	flag.StringVar(&publisher, "publisher", "", "filter by publisher")
	flag.StringVar(&format, "format", "", "filter by format")
	flag.StringVar(&output, "o", "", "folder to store the datasets")
	flag.UintVar(&num, "n", 0, "maximum number of datasets to download")
	flag.BoolVar(&verbose, "v", false, "verbose mode")

	flag.Parse()

	var err error
	if output == "" {
		output, err = os.Getwd()
		check(err)
	} else {
		output, err = filepath.Abs(output)
		check(err)
	}

	if fi, err := os.Stat(output); os.IsNotExist(err) {
		check(os.MkdirAll(output, 0755))
	} else if err != nil {
		check(err)
	} else if !fi.IsDir() {
		logrus.Errorf("output directory %s exists and is not a directory", output)
		os.Exit(1)
	}

	if title == "" && keyword == "" && theme == "" && publisher == "" && format == "" {
		logrus.Error("at least one of -title, -keyword, -theme, -publisher or -format must be provided")
		os.Exit(1)
	}

	client, err := datos.NewClient()
	check(err)

	var f getFunc
	if title != "" {
		f = func(p datos.Params) ([]datos.Dataset, error) {
			return client.DatasetsByTitle(title, p)
		}
	}

	if keyword != "" && f == nil {
		f = func(p datos.Params) ([]datos.Dataset, error) {
			return client.DatasetsByKeyword(keyword, p)
		}
	} else if keyword != "" && f != nil {
		logrus.Warn("ignoring -keyword, because another filter parameter has already been provided")
	}

	if theme != "" && f == nil {
		f = func(p datos.Params) ([]datos.Dataset, error) {
			return client.DatasetsByTheme(theme, p)
		}
	} else if theme != "" && f != nil {
		logrus.Warn("ignoring -theme, because another filter parameter has already been provided")
	}

	if publisher != "" && f == nil {
		f = func(p datos.Params) ([]datos.Dataset, error) {
			return client.DatasetsByPublisher(publisher, p)
		}
	} else if publisher != "" && f != nil {
		logrus.Warn("ignoring -publisher, because another filter parameter has already been provided")
	}

	if format != "" && f == nil {
		f = func(p datos.Params) ([]datos.Dataset, error) {
			return client.DatasetsByFormat(format, p)
		}
	}

	datasets, err := findAllDatasets(f, int(num), formats[strings.ToLower(format)])
	check(err)

	check(downloadAll(datasets, output))
}

type getFunc func(datos.Params) ([]datos.Dataset, error)

type dataset struct {
	url   string
	title string
	id    string
}

func findAllDatasets(f getFunc, max int, format string) ([]dataset, error) {
	var result []dataset
	params := datos.Params{
		Page:     0,
		PageSize: 100,
	}

	for {
		datasets, err := f(params)
		if err != nil {
			return nil, err
		}

		for _, ds := range datasets {
			var url string
			for _, d := range ds.Distribution {
				if format == "" && allowedFormats[d.Format.Value] {
					url = d.AccessURL
					break
				}

				if d.Format.Value == format {
					url = d.AccessURL
					break
				}
			}

			var id = ds.Identifier
			var title string
			if len(ds.Title) > 0 {
				title = ds.Title[0]
			}

			if id == "" && len(ds.Title) > 0 {
				id = ds.Title[0]
			}

			if ds.Identifier == "" && verbose {
				logrus.Warn("found dataset with no identifier")
				continue
			}

			if url == "" && verbose {
				var txt = id
				if id == "" {
					txt = title
				}

				logrus.Warnf("no suitable distribution found for dataset: %s", txt)
				continue
			}

			result = append(result, dataset{url, title, slugify(id, ds.Issued.Time)})

			if len(result) >= max {
				return result, nil
			}
		}

		if len(datasets) < int(params.PageSize) {
			return result, nil
		}

		params.Page++
	}
}

func downloadAll(datasets []dataset, output string) error {
	for _, d := range datasets {
		if err := download(d, output); err != nil {
			return err
		}
	}
	return nil
}

func download(d dataset, output string) error {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(d.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var ext string
	typ := resp.Header.Get("Content-Type")
	if strings.Contains(typ, "xml") {
		ext = ".xml"
	} else if strings.Contains(typ, "json") {
		ext = ".json"
	} else if strings.Contains(typ, "csv") {
		ext = ".csv"
	}

	path := filepath.Join(output, d.id+ext)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		logrus.Errorf("error downoading dataset: %s", d.id)
		return err
	}

	logrus.Infof("downloaded dataset %q to %s", d.title, path)

	return nil
}

func check(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

func slugify(name string, issued time.Time) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i:]
	}

	var result []rune
	var last rune
	for _, r := range name {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			r == '_' || r == '-' {
			result = append(result, r)
		} else if last != '-' {
			result = append(result, '-')
		}

		last = result[len(result)-1]
	}
	return fmt.Sprintf("%s-%d", string(result), issued.Unix())
}

var formats = map[string]string{
	"csv":  "text/csv",
	"json": "application/json",
	"xml":  "application/xml",
}

var allowedFormats = map[string]bool{
	"text/csv":         true,
	"application/json": true,
	"application/xml":  true,
}
