package hrc

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
)

var ErrExportEmpty = errors.New("没有符合条件的数据")

// ExportService 后台导出服务（11 §四 P0#5，v6 Export/Admin 平移）
type ExportService struct{}

// ExportCompanies 企业导出（按企业资料 id 列表生成 CSV，含 UTF-8 BOM）
func (s *ExportService) ExportCompanies(ctx context.Context, ids []uint64) ([]byte, error) {
	if len(ids) == 0 {
		return nil, ErrExportEmpty
	}
	var list []hrcModel.CompanyProfile
	if err := global.GVA_DB.WithContext(ctx).Where("id IN ?", ids).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrExportEmpty
	}

	headers := []string{
		"企业名称", "企业性质", "所属行业", "企业规模", "所在地区", "注册资金",
		"企业网址", "企业福利", "企业简介", "联系人", "联系电话", "座机", "联系邮箱", "联系地址",
	}
	rows := make([][]string, 0, len(list))
	for _, c := range list {
		rows = append(rows, []string{
			companyNameStrPtr(c.CompanyName), c.NatureCN, c.TradeCN, c.ScaleCN, c.DistrictCN, c.Registered,
			c.Website, c.Tag, c.Contents, c.Contact, c.Telephone, c.LandlineTel, c.Email, c.Address,
		})
	}
	return buildCSV(headers, rows)
}

// ExportJobs 职位导出（按 ms_jobs.id 列表，仅未删除行，CSV + UTF-8 BOM）
func (s *ExportService) ExportJobs(ctx context.Context, ids []uint64) ([]byte, error) {
	if len(ids) == 0 {
		return nil, ErrExportEmpty
	}
	var list []hrcModel.Jobs
	if err := global.GVA_DB.WithContext(ctx).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Order("id desc").
		Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrExportEmpty
	}

	headers := []string{
		"职位ID", "职位名称", "企业名称", "企业UID", "工作性质", "职位分类", "地区",
		"学历要求码", "经验要求码", "薪资", "招聘人数", "审核状态", "展示状态",
		"发布时间", "刷新时间", "点击量",
	}
	rows := make([][]string, 0, len(list))
	for _, j := range list {
		rows = append(rows, []string{
			strconv.FormatUint(j.ID, 10),
			j.JobsName,
			j.CompanyName,
			strconv.FormatUint(j.UID, 10),
			j.NatureCN,
			j.CategoryCN,
			j.DistrictCN,
			strconv.Itoa(int(j.Education)),
			strconv.Itoa(int(j.Experience)),
			formatJobWage(j),
			strconv.Itoa(int(j.Amount)),
			auditStatusCN(j.Audit),
			displayStatusCN(j.Display),
			formatTimeCSV(j.AddTime),
			formatTimeCSV(j.Refreshtime),
			strconv.FormatUint(uint64(j.Click), 10),
		})
	}
	return buildCSV(headers, rows)
}

func formatJobWage(j hrcModel.Jobs) string {
	if j.Negotiable == 1 {
		return "面议"
	}
	return fmt.Sprintf("%d-%d", j.MinWage, j.MaxWage)
}

func auditStatusCN(a int8) string {
	switch a {
	case 0:
		return "草稿"
	case 1:
		return "已通过"
	case 2:
		return "审核中"
	case 3:
		return "不通过"
	default:
		return strconv.Itoa(int(a))
	}
}

func displayStatusCN(d int8) string {
	switch d {
	case 1:
		return "展示"
	case 2:
		return "暂停"
	default:
		return strconv.Itoa(int(d))
	}
}

func formatTimeCSV(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.Format("2006-01-02 15:04:05")
}

// buildCSV 生成带 UTF-8 BOM 的 CSV 字节（Excel 打开不乱码）
func buildCSV(headers []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
