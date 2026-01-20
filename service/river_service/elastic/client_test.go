package elastic

import (
	"flag"
	"fmt"
	"testing"

	. "github.com/pingcap/check"
)

// host 定义用于测试的Elasticsearch服务器主机地址
var host = flag.String("host", "127.0.0.1", "Elasticsearch host")

// port 定义用于测试的Elasticsearch服务器端口
var port = flag.Int("port", 9200, "Elasticsearch port")

// Test 是使用gocheck框架运行测试的入口点
func Test(t *testing.T) {
	TestingT(t)
}

// elasticTestSuite 表示Elasticsearch客户端测试的测试套件
type elasticTestSuite struct {
	c *Client // Elasticsearch客户端实例
}

// 将测试套件注册到gocheck
var _ = Suite(&elasticTestSuite{})

// SetUpSuite 通过创建新的Elasticsearch客户端来初始化测试套件
func (s *elasticTestSuite) SetUpSuite(c *C) {
	cfg := new(ClientConfig)
	cfg.Addr = fmt.Sprintf("%s:%d", *host, *port)
	cfg.User = ""
	cfg.Password = ""
	s.c = NewClient(cfg)
}

// TearDownSuite 清理测试套件（当前为空）
func (s *elasticTestSuite) TearDownSuite(c *C) {

}

// makeTestData 创建一个包含name和content字段的测试数据映射
func makeTestData(arg1 string, arg2 string) map[string]interface{} {
	m := make(map[string]interface{})
	m["name"] = arg1
	m["content"] = arg2

	return m
}

// TestSimple 对Elasticsearch客户端执行基本CRUD操作测试
// 它测试文档创建、检索、存在性检查、删除和批量操作
func (s *elasticTestSuite) TestSimple(c *C) {
	// 定义测试索引和文档类型
	index := "dummy"
	docType := "blog"

	//key1 := "name"
	//key2 := "content"

	// 测试文档更新操作
	err := s.c.Update(index, docType, "1", makeTestData("abc", "hello world"))
	c.Assert(err, IsNil)

	// 测试文档存在性检查
	exists, err := s.c.Exists(index, docType, "1")
	c.Assert(err, IsNil)
	c.Assert(exists, Equals, true)

	// 测试文档检索
	r, err := s.c.Get(index, docType, "1")
	c.Assert(err, IsNil)
	c.Assert(r.Code, Equals, 200)
	c.Assert(r.ID, Equals, "1")

	// 测试文档删除
	err = s.c.Delete(index, docType, "1")
	c.Assert(err, IsNil)

	// 验证文档已被删除
	exists, err = s.c.Exists(index, docType, "1")
	c.Assert(err, IsNil)
	c.Assert(exists, Equals, false)

	// 准备批量操作请求（索引10个文档）
	items := make([]*BulkRequest, 10)

	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d", i)
		req := new(BulkRequest)
		req.Action = ActionIndex // 设置操作类型为索引
		req.ID = id
		req.Data = makeTestData(fmt.Sprintf("abc %d", i), fmt.Sprintf("hello world %d", i))
		items[i] = req
	}

	// 执行批量索引操作
	resp, err := s.c.IndexTypeBulk(index, docType, items)
	c.Assert(err, IsNil)
	c.Assert(resp.Code, Equals, 200)
	c.Assert(resp.Errors, Equals, false)

	// 准备批量操作请求（删除10个文档）
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d", i)
		req := new(BulkRequest)
		req.Action = ActionDelete // 设置操作类型为删除
		req.ID = id
		items[i] = req
	}

	// 执行批量删除操作
	resp, err = s.c.IndexTypeBulk(index, docType, items)
	c.Assert(err, IsNil)
	c.Assert(resp.Code, Equals, 200)
	c.Assert(resp.Errors, Equals, false)
}

// TestParent 测试Elasticsearch中的父子关系功能
// 这需要在映射配置中设置父级
func (s *elasticTestSuite) TestParent(c *C) {
	// 定义测试索引、文档类型和父类型
	index := "dummy"
	docType := "comment"
	ParentType := "parent"

	// 创建具有父子关系配置的映射
	mapping := map[string]interface{}{
		docType: map[string]interface{}{
			"_parent": map[string]string{"type": ParentType}, // 配置父类型
		},
	}
	err := s.c.CreateMapping(index, docType, mapping)
	c.Assert(err, IsNil)

	// 准备批量请求以索引具有父子关系的子文档
	items := make([]*BulkRequest, 10)

	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d", i)
		req := new(BulkRequest)
		req.Action = ActionIndex
		req.ID = id
		req.Data = makeTestData(fmt.Sprintf("abc %d", i), fmt.Sprintf("hello world %d", i))
		req.Parent = "1" // 设置父子关系
		items[i] = req
	}

	// 执行子文档的批量索引
	resp, err := s.c.IndexTypeBulk(index, docType, items)
	c.Assert(err, IsNil)
	c.Assert(resp.Code, Equals, 200)
	c.Assert(resp.Errors, Equals, false)

	// 准备批量请求以删除具有父子关系的子文档
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d", i)
		req := new(BulkRequest)
		req.Index = index
		req.Type = docType
		req.Action = ActionDelete
		req.ID = id
		req.Parent = "1" // 指定删除的父子关系
		items[i] = req
	}

	// 执行子文档的批量删除
	resp, err = s.c.Bulk(items)
	c.Assert(err, IsNil)
	c.Assert(resp.Code, Equals, 200)
	c.Assert(resp.Errors, Equals, false)
}
