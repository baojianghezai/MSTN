package initialize

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/stretchr/testify/require"
)

func TestSeedCategoriesIsIdempotentAndComplete(t *testing.T) {
	db := testutil.NewMemoryDBWithoutGlobal(t, &hrcModel.CategoryGroup{}, &hrcModel.Category{})

	require.NoError(t, seedCategories(db))
	var jobTitleGroup hrcModel.CategoryGroup
	require.NoError(t, db.Where("alias = ?", "jobtitle").First(&jobTitleGroup).Error)
	var existingTitle hrcModel.Category
	require.NoError(t, db.Where("group_id = ? AND name = ?", jobTitleGroup.ID, "Java开发").First(&existingTitle).Error)
	require.NoError(t, db.Model(&existingTitle).Update("display", 2).Error)
	require.NoError(t, seedCategories(db)) // 幂等：二次执行不报错、不重复、不覆盖运营配置
	require.NoError(t, db.First(&existingTitle, existingTitle.ID).Error)
	require.Equal(t, int8(2), existingTitle.Display)

	var groupCount int64
	require.NoError(t, db.Model(&hrcModel.CategoryGroup{}).Count(&groupCount).Error)
	require.Equal(t, int64(18), groupCount)
	require.Len(t, legacyJobCatalog(), 1655)
	require.Greater(t, len(legacyJobTitles()), 1000)
	flat := legacyFlatCategories()
	require.Len(t, flat["trade"], 45)
	require.Len(t, flat["jobtag"], 21)
	require.Len(t, flat["resumetag"], 21)
	require.Len(t, flat["language"], 6)
	require.Len(t, flat["languagelevel"], 3)
	require.Len(t, flat["current"], 5)
	require.Len(t, flat["age"], 5)

	// 关键分组存在且分类数量正确（value=分类 id、label=分类 name，前端下拉数据源）
	checks := map[string]int{
		"major":    23,
		"sex":      2,
		"marriage": 3,
		"nature":   7,
		"scale":    5,
		"jobtitle": 1432,
	}
	for alias, want := range checks {
		var group hrcModel.CategoryGroup
		require.NoError(t, db.Where("alias = ?", alias).First(&group).Error, "分组 %s 未 seed", alias)
		var n int64
		require.NoError(t, db.Model(&hrcModel.Category{}).Where("group_id = ?", group.ID).Count(&n).Error, alias)
		if alias == "jobtitle" {
			require.Equal(t, int64(want), n, "分组 %s 分类数量不符", alias)
		} else {
			require.GreaterOrEqual(t, n, int64(want), "分组 %s 分类数量不符", alias)
		}
	}

	// district 分组存在但分类暂为空（省市县待导入）
	var district hrcModel.CategoryGroup
	require.NoError(t, db.Where("alias = ?", "district").First(&district).Error)
	var dCount int64
	require.NoError(t, db.Model(&hrcModel.Category{}).Where("group_id = ?", district.ID).Count(&dCount).Error)
	require.Equal(t, int64(3241), dCount)
	var topLevelCount int64
	require.NoError(t, db.Model(&hrcModel.Category{}).Where("group_id = ? AND parent_id = ?", district.ID, 0).Count(&topLevelCount).Error)
	require.Equal(t, int64(34), topLevelCount)
	require.Len(t, legacyDistrictCatalog(), 3241)
}

func TestSeedJobCategoriesMigratesLegacyTreeIdempotently(t *testing.T) {
	db := testutil.NewMemoryDBWithoutGlobal(t, &hrcModel.CategoryGroup{}, &hrcModel.Category{})

	require.NoError(t, seedJobCategories(db))
	require.NoError(t, seedJobCategories(db))

	var group hrcModel.CategoryGroup
	require.NoError(t, db.Where("alias = ?", "jobcategory").First(&group).Error)

	var total int64
	require.NoError(t, db.Model(&hrcModel.Category{}).Where("group_id = ?", group.ID).Count(&total).Error)
	require.Equal(t, int64(1849), total)

	var topClass hrcModel.Category
	require.NoError(t, db.Where("group_id = ? AND parent_id = ? AND name = ?", group.ID, 0, "技术").First(&topClass).Error)
	var category hrcModel.Category
	require.NoError(t, db.Where("group_id = ? AND parent_id = ? AND name = ?", group.ID, topClass.ID, "后端开发").First(&category).Error)
	var title hrcModel.Category
	require.NoError(t, db.Where("group_id = ? AND parent_id = ? AND name = ?", group.ID, category.ID, "ERP技术/应用").First(&title).Error)
}
