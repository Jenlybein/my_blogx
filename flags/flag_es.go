package flags

import (
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/es_service"
)

func FlagESIndex() {
	// 初始化ES索引
	article := models.ArticleModel{}
	index := article.Index()

	fmt.Println("请输入索引设置: ")
	fmt.Println("1. 初始化文章索引设置")
	fmt.Println("2. 删除文章索引设置")

	var indexChoice int
	fmt.Scanln(&indexChoice)

	switch indexChoice {
	case 1:
		// 初始化文章索引设置
		if err := es_service.CreateIndexForce(index, article.Mapping()); err != nil {
			global.Logger.Errorf("初始化索引失败: %v", err)
			return
		}
		global.Logger.Infof("索引 %s 初始化成功", index)
	case 2:
		// 删除文章索引设置
		if err := es_service.DeleteIndex(index); err != nil {
			global.Logger.Errorf("删除索引失败: %v", err)
			return
		}
		global.Logger.Infof("索引 %s 删除成功", index)
	default:
		fmt.Println("无效的选择，不执行任何操作")
	}

	// 初始化ES pipeline
	pipelineName := article.PipelineName()

	fmt.Println("请输入pipeline设置: ")
	fmt.Println("1. 初始化文章pipeline设置")
	fmt.Println("2. 删除文章pipeline设置")

	var pipelineChoice int
	fmt.Scanln(&pipelineChoice)

	switch pipelineChoice {
	case 1:
		// 初始化文章pipeline设置
		if err := es_service.CreatePipelineForce(pipelineName, article.Pipeline()); err != nil {
			global.Logger.Errorf("初始化pipeline失败: %v", err)
			return
		}
		global.Logger.Infof("pipeline %s 初始化成功", pipelineName)
	case 2:
		// 删除文章pipeline设置
		if err := es_service.DeletePipeline(pipelineName); err != nil {
			global.Logger.Errorf("删除pipeline失败: %v", err)
			return
		}
		global.Logger.Infof("pipeline %s 删除成功", pipelineName)
	default:
		fmt.Println("无效的选择，不执行任何操作")
	}
}
