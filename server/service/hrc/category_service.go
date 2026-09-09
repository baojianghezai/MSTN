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

func (s *CategoryService) GetJobsTree(ctx context.Context) ([]CategoryTreeNode, error) {
	var categories []hrcModel.CategoryJobs
	db := global.GVA_DB.WithContext(ctx).Model(&hrcModel.CategoryJobs{})
	err := db.Order("level asc, sort asc").Find(&categories).Error

	// 构建 ID -> Node 的映射表
	nodeMap := make(map[int]*CategoryTreeNode)
	for _, cat := range categories {
		nodeMap[cat.ID] = &CategoryTreeNode{
			ID:       cat.ID,
			ParentID: cat.ParentID,
			Level:    cat.Level,
			Name:     cat.Name,
			Children: []CategoryTreeNode{},
		}
	}

	// 按 level 从深到浅组装树：先处理子节点挂到父节点，再收集根节点，
	// 避免单次遍历 map 时根节点先被拷贝（Children 为空）导致子节点丢失。
	// categories 已按 level asc 排好，倒序遍历即可保证先子后父。
	for i := len(categories) - 1; i >= 0; i-- {
		cat := categories[i]
		if cat.ParentID == 0 {
			continue
		}
		child := nodeMap[cat.ID]
		parent := nodeMap[cat.ParentID]
		if parent != nil && child != nil {
			parent.Children = append(parent.Children, *child)
		}
	}

	// 收集根节点（此时其 Children 已包含完整子树）
	var roots []CategoryTreeNode
	for _, cat := range categories {
		if cat.ParentID == 0 {
			roots = append(roots, *nodeMap[cat.ID])
		}
	}

	return roots, err
}
