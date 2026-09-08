package hrc

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

type CategoryService struct{}

// CategoryTreeNode 定义返回的 VO 结构（带 Children）
type CategoryTreeNode struct {
	ID       int                `json:"id"`
	ParentID int                `json:"parentId"`
	Level    int                `json:"level"`
	Name     string             `json:"name"`
	Children []CategoryTreeNode `json:"children"`
}

func (s *CategoryService) GetJobsTree(ctx context.Context) (tree []CategoryTreeNode, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CategoryJobs{})
	// 查询所有分类（按层级和排序字段排好）
	err = db.Order("level asc, sort asc").Find(&tree).Error

	// 构建 ID -> Node 的映射表
	nodeMap := make(map[int]*CategoryTreeNode)
	for _, cat := range tree {
		nodeMap[cat.ID] = &CategoryTreeNode{
			ID:       cat.ID,
			ParentID: cat.ParentID,
			Level:    cat.Level,
			Name:     cat.Name,
			Children: []CategoryTreeNode{},
		}
	}

	// 组装树形结构（仅处理三级，高效）
	var roots []CategoryTreeNode
	for _, node := range nodeMap {
		if node.ParentID == 0 {
			roots = append(roots, *node)
		} else {
			if parent, ok := nodeMap[node.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}
	return roots, err
}
