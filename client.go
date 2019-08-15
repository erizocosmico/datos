package datos

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client to query data from the spanish government open data API.
type Client struct {
	c *http.Client
}

const baseURL = "https://datos.gob.es/apidata"

func getRemoteCertificates(url string) ([]*x509.Certificate, error) {
	d := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(d, "tcp", "datos.gob.es:443", &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return nil, err
	}

	defer conn.Close()
	return conn.ConnectionState().PeerCertificates, nil
}

// NewClient creates a new client to query data from the spanish government open data API.
// It will also install in the client the SSL certificates required to call the API.
func NewClient() (*Client, error) {
	certs, err := getRemoteCertificates(baseURL)
	if err != nil {
		return nil, fmt.Errorf("datos: unable to get certificates: %s", err)
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("datos: unable to get system cert pool: %s", err)
	}

	for _, c := range certs {
		pool.AddCert(c)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}

	return &Client{
		c: client,
	}, nil
}

// Params to control the page, page size and order of the results in any API call.
type Params struct {
	Sort     string
	Page     uint
	PageSize uint
}

func makeURL(path string, params Params) string {
	var queryParts []string
	if params.Sort != "" {
		queryParts = append(queryParts, fmt.Sprintf("_sort=%s", params.Sort))
	}

	if params.Page > 0 {
		queryParts = append(queryParts, fmt.Sprintf("_page=%d", params.Page))
	}

	if params.PageSize > 0 {
		queryParts = append(queryParts, fmt.Sprintf("_pageSize=%d", params.PageSize))
	}

	query := strings.Join(queryParts, "&")
	return fmt.Sprintf("%s%s?%s", baseURL, path, query)
}

func (c *Client) get(
	path string,
	params Params,
	decodeInto interface{},
) error {
	req, err := http.NewRequest("GET", makeURL(path, params), nil)
	if err != nil {
		return fmt.Errorf("datos: unable to create request: %s", err)
	}

	fmt.Println(makeURL(path, params))

	req.Header.Add("Accept", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		return fmt.Errorf("datos: unable to get data from %q: %s", path, err)
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("datos: error reading response body: %s", err)
	}

	if err := json.Unmarshal(bytes, decodeInto); err != nil {
		return fmt.Errorf("datos: unable to decode JSON response into %T: %s", decodeInto, err)
	}

	return nil
}

// Publisher is a data publisher.
type Publisher struct {
	// About contains a link to the information about this object.
	About    string `json:"_about"`
	Notation string `json:"notation"`
	Label    string `json:"prefLabel"`
}

// Publishers lists all data publishers.
func (c *Client) Publishers(params Params) ([]Publisher, error) {
	var resp struct {
		Result struct {
			Items []Publisher `json:"items"`
		} `json:"result"`
	}
	if err := c.get("/catalog/publisher", params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// Spatial is the data about a country, municipality or province.
type Spatial struct {
	// About contains a link to the information about this object.
	About string `json:"_about"`
	Label string `json:"label"`
	// Country is a link to the country information.
	Country string `json:"pais"`
	// Type of the spatial.
	Type string `json:"type"`
}

// Spatials returns all spatials.
func (c *Client) Spatials(params Params) ([]Spatial, error) {
	var resp struct {
		Result struct {
			Items []Spatial `json:"items"`
		} `json:"result"`
	}
	if err := c.get("/catalog/spatials", params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// Theme of a dataset.
type Theme struct {
	// About contains a link to the information about this object.
	About    string   `json:"_about"`
	Notation string   `json:"notation"`
	Labels   []string `json:"prefLabel"`
}

// Themes returns all themes.
func (c *Client) Themes(params Params) ([]Theme, error) {
	var resp struct {
		Result struct {
			Items []Theme `json:"items"`
		} `json:"result"`
	}
	if err := c.get("/catalog/themes", params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// Dataset data.
type Dataset struct {
	// About contains a link to the information about this object.
	About       string   `json:"_about"`
	Modified    Datetime `json:"modified"`
	Description []struct {
		Text string `json:"text"`
		Lang string `json:"lang"`
	} `json:"description"`
	Distribution       Distribution `json:"distribution"`
	Identifier         string       `json:"identifier"`
	Keywords           []string     `json:"keyword"`
	Language           string       `json:"language"`
	License            string       `json:"license"`
	Publisher          string       `json:"publisher"`
	References         []string     `json:"references"`
	Spatial            string       `json:"spatial"`
	Temporal           string       `json:"temporal"`
	Theme              string       `json:"theme"`
	Title              Strings      `json:"title"`
	AccrualPeriodicity string       `json:"accrualPeriodicity"`
	ConformsTo         string       `json:"conformsTo"`
	Issued             Datetime     `json:"issued"`
	Valid              Datetime     `json:"valid"`
}

type datasetResp struct {
	Result struct {
		Items []Dataset `json:"items"`
	} `json:"result"`
}

// Datasets returns all datasets.
func (c *Client) Datasets(params Params) ([]Dataset, error) {
	var resp datasetResp
	if err := c.get("/catalog/dataset", params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// Dataset returns the dataset with the given ID.
func (c *Client) Dataset(id string, params Params) (Dataset, error) {
	var resp datasetResp
	if err := c.get("/catalog/dataset/"+url.PathEscape(id), params, &resp); err != nil {
		return Dataset{}, err
	}

	if len(resp.Result.Items) > 0 {
		return resp.Result.Items[0], nil
	}

	return Dataset{}, fmt.Errorf("datos: dataset not found with id %q", id)
}

// DatasetsByTitle returns the datasets matching the given title.
func (c *Client) DatasetsByTitle(title string, params Params) ([]Dataset, error) {
	var resp datasetResp
	if err := c.get("/catalog/dataset/title/"+url.PathEscape(title), params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// DatasetsByPublisher returns the datasets with the given publisher ID.
func (c *Client) DatasetsByPublisher(publisherID string, params Params) ([]Dataset, error) {
	var resp datasetResp
	if err := c.get("/catalog/dataset/publisher/"+url.PathEscape(publisherID), params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// DatasetsByTheme returns the datasets with the given theme ID.
func (c *Client) DatasetsByTheme(themeID string, params Params) ([]Dataset, error) {
	var resp datasetResp
	if err := c.get("/catalog/dataset/theme/"+url.PathEscape(themeID), params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// DatasetsByFormat returns the datasets with the given format.
func (c *Client) DatasetsByFormat(format string, params Params) ([]Dataset, error) {
	var resp datasetResp
	if err := c.get("/catalog/dataset/format/"+url.PathEscape(format), params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// DatasetsByKeyword returns the datasets with the given keyword.
func (c *Client) DatasetsByKeyword(keyword string, params Params) ([]Dataset, error) {
	var resp datasetResp
	if err := c.get("/catalog/dataset/format/"+url.PathEscape(keyword), params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// SpatialType is the type of a spatial.
type SpatialType byte

const (
	// Autonomy type.
	Autonomy SpatialType = iota
	// Country type.
	Country
	// Province type.
	Province
)

func (t SpatialType) String() string {
	switch t {
	case Country:
		return "Pais"
	case Province:
		return "Provincia"
	case Autonomy:
		return "Autonomia"
	default:
		return "Invalid"
	}
}

// DatasetsBySpatial returns the datasets with the given spatial.
func (c *Client) DatasetsBySpatial(typ SpatialType, spatial string, params Params) ([]Dataset, error) {
	var resp datasetResp
	err := c.get(
		fmt.Sprintf("/catalog/dataset/spatial/%s/%s", typ, url.PathEscape(spatial)),
		params,
		&resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// DatasetsModifiedBetween returns the datasets modified between the given date range.
func (c *Client) DatasetsModifiedBetween(from, to time.Time, params Params) ([]Dataset, error) {
	var resp datasetResp
	err := c.get(
		fmt.Sprintf(
			"/catalog/dataset/modified/begin/%s/end/%s",
			from.Format(time.RFC3339),
			to.Format(time.RFC3339),
		),
		params,
		&resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// Distribution of a dataset.
type Distribution struct {
	// About contains a link to the information about this object.
	About     string  `json:"_about"`
	AccessURL string  `json:"accessURL"`
	ByteSize  float64 `json:"byteSize"`
	Format    struct {
		About string `json:"_about"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"format"`
	Relation   string  `json:"relation"`
	Title      Strings `json:"title"`
	Identifier string  `json:"identifier"`
}

type distributionResp struct {
	Result struct {
		Items []Distribution `json:"items"`
	} `json:"result"`
}

// Distributions returns all distributions.
func (c *Client) Distributions(params Params) ([]Distribution, error) {
	var resp distributionResp
	if err := c.get("/catalog/distribution", params, &resp); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// DistributionsByDataset returns all distributions of a dataset.
func (c *Client) DistributionsByDataset(datasetID string, params Params) ([]Distribution, error) {
	var resp distributionResp
	if err := c.get(
		"/catalog/distribution/dataset/"+url.PathEscape(datasetID),
		params,
		&resp,
	); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

// DistributionsByFormat returns all distributions with the given format.
func (c *Client) DistributionsByFormat(format string, params Params) ([]Distribution, error) {
	var resp distributionResp
	if err := c.get(
		"/catalog/distribution/format/"+url.PathEscape(format),
		params,
		&resp,
	); err != nil {
		return nil, err
	}

	return resp.Result.Items, nil
}

type Strings []string

func (s *Strings) UnmarshalJSON(b []byte) error {
	var val interface{}
	err := json.Unmarshal(b, &val)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case string:
		*s = append(*s, v)
	case []interface{}:
		for _, elem := range v {
			str, ok := elem.(string)
			if !ok {
				return fmt.Errorf("expecting string, got %T", elem)
			}
			*s = append(*s, str)
		}
	default:
		return fmt.Errorf("expecting string or string array, got %T", val)
	}

	return nil
}
