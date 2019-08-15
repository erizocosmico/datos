# datos [![GoDoc](https://godoc.org/github.com/erizocosmico/datos?status.svg)](https://godoc.org/github.com/erizocosmico/datos) [![Build Status](https://travis-ci.org/erizocosmico/datos.svg?branch=master)](https://travis-ci.org/erizocosmico/datos) [![codecov](https://codecov.io/gh/erizocosmico/datos/branch/master/graph/badge.svg)](https://codecov.io/gh/erizocosmico/datos) [![Go Report Card](https://goreportcard.com/badge/github.com/erizocosmico/datos)](https://goreportcard.com/report/github.com/erizocosmico/datos) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[Spanish government open data API](https://datos.gob.es/es/apidata/) client for Go.

### Install

```
go get github.com/erizocosmico/datos
```

### Usage

```go
client, err := datos.NewClient()
if err != nil {
    // handle err
}

datasets, err := client.Datasets(datos.Params{Page: 1, PageSize: 50})
if err != nil {
    // handle err
}
```

### License

MIT, see [LICENSE](/LICENSE)
