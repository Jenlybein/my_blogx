package elastic

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pingcap/errors"
)

// Client 是与ES通信的客户端。
// 尽管有很多Go语言的Elasticsearch客户端，我还是想自己实现一个。
// 因为我们只需要一些非常简单的功能。
type Client struct {
	Protocol string // 协议（http或https）
	Addr     string // ES服务器地址
	User     string // 用户名
	Password string // 密码

	c *http.Client // HTTP客户端
}

// ClientConfig 是客户端的配置。
type ClientConfig struct {
	HTTPS    bool   // 是否使用HTTPS协议
	Addr     string // ES服务器地址
	User     string // 用户名
	Password string // 密码
}

// NewClient 根据配置创建客户端。
func NewClient(conf *ClientConfig) *Client {
	c := new(Client)

	c.Addr = conf.Addr
	c.User = conf.User
	c.Password = conf.Password

	if conf.HTTPS {
		c.Protocol = "https"
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.c = &http.Client{Transport: tr}
	} else {
		c.Protocol = "http"
		c.c = &http.Client{}
	}

	return c
}

// ResponseItem 是响应中的ES项目。
type ResponseItem struct {
	ID      string                 `json:"_id"`      // 文档ID
	Index   string                 `json:"_index"`   // 索引名称
	Type    string                 `json:"_type"`    // 文档类型
	Version int                    `json:"_version"` // 版本号
	Found   bool                   `json:"found"`    // 是否找到文档
	Source  map[string]interface{} `json:"_source"`  // 文档源数据
}

// Response 是ES响应
type Response struct {
	Code int // 响应状态码
	ResponseItem
}

// See http://www.elasticsearch.org/guide/en/elasticsearch/guide/current/bulk.html
const (
	ActionCreate = "create" // 创建动作
	ActionUpdate = "update" // 更新动作
	ActionDelete = "delete" // 删除动作
	ActionIndex  = "index"  // 索引动作
)

// BulkRequest 用于批量发送多个请求。
type BulkRequest struct {
	Action   string // 批量操作的动作类型
	Index    string // 索引名称
	Type     string // 文档类型
	ID       string // 文档ID
	Parent   string // 父文档ID
	Pipeline string // 处理管道

	Data map[string]interface{} // 请求数据
}

// bulk 将批量请求序列化为ES格式的数据
func (r *BulkRequest) bulk(buf *bytes.Buffer) error {
	meta := make(map[string]map[string]string)
	metaData := make(map[string]string)
	if len(r.Index) > 0 {
		metaData["_index"] = r.Index
	}
	if len(r.Type) > 0 {
		metaData["_type"] = r.Type
	}

	if len(r.ID) > 0 {
		metaData["_id"] = r.ID
	}
	if len(r.Parent) > 0 {
		metaData["_parent"] = r.Parent
	}
	if len(r.Pipeline) > 0 {
		metaData["pipeline"] = r.Pipeline
	}

	meta[r.Action] = metaData

	data, err := json.Marshal(meta)
	if err != nil {
		return errors.Trace(err)
	}

	buf.Write(data)
	buf.WriteByte('\n')

	switch r.Action {
	case ActionDelete:
		// 删除操作不需要额外数据
	case ActionUpdate:
		doc := map[string]interface{}{
			"doc": r.Data,
		}
		data, err = json.Marshal(doc)
		if err != nil {
			return errors.Trace(err)
		}

		buf.Write(data)
		buf.WriteByte('\n')
	default:
		// 用于创建和索引操作
		data, err = json.Marshal(r.Data)
		if err != nil {
			return errors.Trace(err)
		}

		buf.Write(data)
		buf.WriteByte('\n')
	}

	return nil
}

// BulkResponse 是批量请求的响应。
type BulkResponse struct {
	Code   int  // 响应状态码
	Took   int  `json:"took"`   // 执行耗时（毫秒）
	Errors bool `json:"errors"` // 是否有错误

	Items []map[string]*BulkResponseItem `json:"items"` // 批量操作结果项
}

// BulkResponseItem 是批量响应中的项目。
type BulkResponseItem struct {
	Index   string          `json:"_index"`   // 索引名称
	Type    string          `json:"_type"`    // 文档类型
	ID      string          `json:"_id"`      // 文档ID
	Version int             `json:"_version"` // 版本号
	Status  int             `json:"status"`   // 状态码
	Error   json.RawMessage `json:"error"`    // 错误信息
	Found   bool            `json:"found"`    // 是否找到
}

// MappingResponse 是映射请求的响应。
type MappingResponse struct {
	Code    int     // 响应状态码
	Mapping Mapping // 映射结构
}

// Mapping 表示ES映射。
type Mapping map[string]struct {
	Mappings map[string]struct {
		Properties map[string]struct {
			Type   string      `json:"type"`   // 字段类型
			Fields interface{} `json:"fields"` // 字段定义
		} `json:"properties"` // 属性映射
	} `json:"mappings"` // 映射定义
}

// DoRequest 发送带有请求体的请求到ES。
func (c *Client) DoRequest(method string, url string, body *bytes.Buffer) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, errors.Trace(err)
	}
	if len(c.User) > 0 && len(c.Password) > 0 {
		req.SetBasicAuth(c.User, c.Password)
	}
	resp, err := c.c.Do(req)

	return resp, err
}

// Do 发送带有请求体的请求到ES。
func (c *Client) Do(method string, url string, body map[string]interface{}) (*Response, error) {
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	buf := bytes.NewBuffer(bodyData)
	if body == nil {
		buf = bytes.NewBuffer(nil)
	}

	resp, err := c.DoRequest(method, url, buf)
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer resp.Body.Close()

	ret := new(Response)
	ret.Code = resp.StatusCode

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &ret.ResponseItem)
	}

	return ret, errors.Trace(err)
}

// DoBulk 发送批量请求到ES。
func (c *Client) DoBulk(url string, items []*BulkRequest) (*BulkResponse, error) {
	var buf bytes.Buffer

	for _, item := range items {
		if err := item.bulk(&buf); err != nil {
			return nil, errors.Trace(err)
		}
	}

	resp, err := c.DoRequest("POST", url, &buf)
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer resp.Body.Close()

	ret := new(BulkResponse)
	ret.Code = resp.StatusCode

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &ret)
	}

	return ret, errors.Trace(err)
}

// CreateMapping 创建ES映射。
func (c *Client) CreateMapping(index string, docType string, mapping map[string]interface{}) error {
	reqURL := fmt.Sprintf("%s://%s/%s", c.Protocol, c.Addr,
		url.QueryEscape(index))

	r, err := c.Do("HEAD", reqURL, nil)
	if err != nil {
		return errors.Trace(err)
	}

	// 如果索引不存在，会得到404未找到错误，需要先创建索引
	if r.Code == http.StatusNotFound {
		_, err = c.Do("PUT", reqURL, nil)

		if err != nil {
			return errors.Trace(err)
		}
	} else if r.Code != http.StatusOK {
		return errors.Errorf("Error: %s, code: %d", http.StatusText(r.Code), r.Code)
	}

	reqURL = fmt.Sprintf("%s://%s/%s/%s/_mapping", c.Protocol, c.Addr,
		url.QueryEscape(index),
		url.QueryEscape(docType))

	_, err = c.Do("POST", reqURL, mapping)
	return errors.Trace(err)
}

// GetMapping 获取映射。
func (c *Client) GetMapping(index string, docType string) (*MappingResponse, error) {
	reqURL := fmt.Sprintf("%s://%s/%s/%s/_mapping", c.Protocol, c.Addr,
		url.QueryEscape(index),
		url.QueryEscape(docType))
	buf := bytes.NewBuffer(nil)
	resp, err := c.DoRequest("GET", reqURL, buf)

	if err != nil {
		return nil, errors.Trace(err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ret := new(MappingResponse)
	err = json.Unmarshal(data, &ret.Mapping)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ret.Code = resp.StatusCode
	return ret, errors.Trace(err)
}

// DeleteIndex 删除索引。
func (c *Client) DeleteIndex(index string) error {
	reqURL := fmt.Sprintf("%s://%s/%s", c.Protocol, c.Addr,
		url.QueryEscape(index))

	r, err := c.Do("DELETE", reqURL, nil)
	if err != nil {
		return errors.Trace(err)
	}

	if r.Code == http.StatusOK || r.Code == http.StatusNotFound {
		return nil
	}

	return errors.Errorf("Error: %s, code: %d", http.StatusText(r.Code), r.Code)
}

// Get 根据ID获取项目。
func (c *Client) Get(index string, docType string, id string) (*Response, error) {
	reqURL := fmt.Sprintf("%s://%s/%s/%s/%s", c.Protocol, c.Addr,
		url.QueryEscape(index),
		url.QueryEscape(docType),
		url.QueryEscape(id))

	return c.Do("GET", reqURL, nil)
}

// Update 创建或更新数据
func (c *Client) Update(index string, docType string, id string, data map[string]interface{}) error {
	reqURL := fmt.Sprintf("%s://%s/%s/%s/%s", c.Protocol, c.Addr,
		url.QueryEscape(index),
		url.QueryEscape(docType),
		url.QueryEscape(id))

	r, err := c.Do("PUT", reqURL, data)
	if err != nil {
		return errors.Trace(err)
	}

	if r.Code == http.StatusOK || r.Code == http.StatusCreated {
		return nil
	}

	return errors.Errorf("Error: %s, code: %d", http.StatusText(r.Code), r.Code)
}

// Exists 检查ID是否存在。
func (c *Client) Exists(index string, docType string, id string) (bool, error) {
	reqURL := fmt.Sprintf("%s://%s/%s/%s/%s", c.Protocol, c.Addr,
		url.QueryEscape(index),
		url.QueryEscape(docType),
		url.QueryEscape(id))

	r, err := c.Do("HEAD", reqURL, nil)
	if err != nil {
		return false, err
	}

	return r.Code == http.StatusOK, nil
}

// Delete 根据ID删除项目。
func (c *Client) Delete(index string, docType string, id string) error {
	reqURL := fmt.Sprintf("%s://%s/%s/%s/%s", c.Protocol, c.Addr,
		url.QueryEscape(index),
		url.QueryEscape(docType),
		url.QueryEscape(id))

	r, err := c.Do("DELETE", reqURL, nil)
	if err != nil {
		return errors.Trace(err)
	}

	if r.Code == http.StatusOK || r.Code == http.StatusNotFound {
		return nil
	}

	return errors.Errorf("Error: %s, code: %d", http.StatusText(r.Code), r.Code)
}

// Bulk 发送批量请求。
// 仅在'Bulk'相关API中支持父级
func (c *Client) Bulk(items []*BulkRequest) (*BulkResponse, error) {
	reqURL := fmt.Sprintf("%s://%s/_bulk", c.Protocol, c.Addr)

	return c.DoBulk(reqURL, items)
}

// IndexBulk 发送针对索引的批量请求。
func (c *Client) IndexBulk(index string, items []*BulkRequest) (*BulkResponse, error) {
	reqURL := fmt.Sprintf("%s://%s/%s/_bulk", c.Protocol, c.Addr,
		url.QueryEscape(index))

	return c.DoBulk(reqURL, items)
}

// IndexTypeBulk 发送针对索引和文档类型的批量请求。
func (c *Client) IndexTypeBulk(index string, docType string, items []*BulkRequest) (*BulkResponse, error) {
	reqURL := fmt.Sprintf("%s://%s/%s/%s/_bulk", c.Protocol, c.Addr,
		url.QueryEscape(index),
		url.QueryEscape(docType))

	return c.DoBulk(reqURL, items)
}
