package hrc

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/go-pdf/fpdf"
)

// ErrResumePDFFontMissing 未找到可用于中文渲染的 TTF 字体
var ErrResumePDFFontMissing = errors.New("简历 PDF 字体缺失（请配置 RESUME_PDF_FONT 或放置 resource/font/*.ttf）")

// pdfStyle 简历模板样式（#22：1=经典 2=简约 3=紧凑）
type pdfStyle struct {
	accent     [3]int  // 主题色
	titleSize  float64 // 章节标题字号
	bodySize   float64 // 正文字号
	lineH      float64 // 行高（mm）
	nameSize   float64 // 姓名字号
	filledBar  bool    // 章节标题是否填充色块
	centerName bool    // 姓名是否居中
}

func pdfStyleFor(template int8) pdfStyle {
	switch template {
	case 2:
		return pdfStyle{accent: [3]int{13, 148, 136}, titleSize: 12, bodySize: 10, lineH: 6, nameSize: 22, filledBar: false, centerName: true}
	case 3:
		return pdfStyle{accent: [3]int{51, 65, 85}, titleSize: 11, bodySize: 9.5, lineH: 5.4, nameSize: 18, filledBar: true, centerName: false}
	default:
		return pdfStyle{accent: [3]int{37, 99, 235}, titleSize: 12, bodySize: 10.5, lineH: 6.2, nameSize: 24, filledBar: true, centerName: false}
	}
}

// resumePDFFontCandidates 候选中文 TTF 路径（fpdf 仅支持 TTF，不支持 ttc/otf）
func resumePDFFontCandidates() []string {
	// 显式配置优先且唯一：便于部署指定字体、也便于测试强制走 HTML 回退
	if env := strings.TrimSpace(os.Getenv("RESUME_PDF_FONT")); env != "" {
		return []string{env}
	}
	cands := make([]string, 0, 8)
	if entries, err := os.ReadDir("resource/font"); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".ttf") {
				cands = append(cands, filepath.Join("resource/font", e.Name()))
			}
		}
	}
	cands = append(cands,
		`C:\Windows\Fonts\simhei.ttf`,
		`C:\Windows\Fonts\Deng.ttf`,
		`C:\Windows\Fonts\simkai.ttf`,
		`C:\Windows\Fonts\simfang.ttf`,
		`/usr/share/fonts/truetype/arphic/ukai.ttf`,
		`/usr/share/fonts/truetype/arphic/uming.ttf`,
		`/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf`,
	)
	return cands
}

// FindResumePDFFont 返回首个可用的中文 TTF 字体路径
func FindResumePDFFont() (string, error) {
	for _, p := range resumePDFFontCandidates() {
		if !strings.HasSuffix(strings.ToLower(p), ".ttf") {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", ErrResumePDFFontMissing
}

// RenderResumePDF 将在线简历渲染为 PDF（template 决定版式；字体缺失返回 ErrResumePDFFontMissing）
func RenderResumePDF(resume *hrcModel.Resume, subs *ResumeSubTables, template int8) ([]byte, error) {
	fontPath, err := FindResumePDFFont()
	if err != nil {
		return nil, err
	}
	subs = subs.safe()
	style := pdfStyleFor(template)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 16, 18)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddUTF8Font("cjk", "", fontPath)
	pdf.AddPage()

	renderResumeHeader(pdf, resume, style)
	renderIntention(pdf, resume, style)
	renderPDFEntries(pdf, "教育经历", len(subs.Education), style, func(i int) string {
		e := subs.Education[i]
		return fmt.Sprintf("%s  %s%s", formatResumePeriod(e.StartYear, e.StartMonth, e.EndYear, e.EndMonth, e.ToDate), e.School, joinResumeDetails(e.Speciality, e.EducationCN))
	})
	renderPDFEntries(pdf, "工作/实习经历", len(subs.Work), style, func(i int) string {
		w := subs.Work[i]
		return fmt.Sprintf("[%s] %s  %s%s%s", workTypeCN(w.WorkType), formatResumePeriod(w.StartYear, w.StartMonth, w.EndYear, w.EndMonth, w.ToDate), w.CompanyName, joinResumeDetails(w.Jobs), joinResumeDetails(w.Achievements))
	})
	renderPDFEntries(pdf, "项目经历", len(subs.Project), style, func(i int) string {
		p := subs.Project[i]
		return fmt.Sprintf("%s  %s%s%s", formatResumePeriod(p.StartYear, p.StartMonth, p.EndYear, p.EndMonth, p.ToDate), p.ProjectName, joinResumeDetails(p.Role), joinResumeDetails(p.Description))
	})
	renderPDFEntries(pdf, "专业技能", len(subs.Skill), style, func(i int) string {
		s := subs.Skill[i]
		return joinResumeDetails(s.Name, skillLevelCN(s.Level))
	})
	renderPDFEntries(pdf, "语言能力", len(subs.Language), style, func(i int) string {
		l := subs.Language[i]
		return joinResumeDetails(l.LanguageCN, l.LevelCN)
	})
	renderPDFEntries(pdf, "培训经历", len(subs.Training), style, func(i int) string {
		t := subs.Training[i]
		return fmt.Sprintf("%s  %s%s%s", formatResumePeriod(t.StartYear, t.StartMonth, t.EndYear, t.EndMonth, t.ToDate), t.Agency, joinResumeDetails(t.Course), joinResumeDetails(t.Description))
	})
	renderPDFEntries(pdf, "证书", len(subs.Credent), style, func(i int) string {
		c := subs.Credent[i]
		return joinResumeDetails(fmt.Sprintf("%d-%02d", c.Year, c.Month), c.Name)
	})
	renderPDFEntries(pdf, "个人作品", len(subs.Portfolio), style, func(i int) string {
		p := subs.Portfolio[i]
		return joinResumeDetails(p.Title, p.Description, p.URL)
	})
	renderPDFEntries(pdf, "学生干部经历", len(subs.StudentLeader), style, func(i int) string {
		s := subs.StudentLeader[i]
		return fmt.Sprintf("%s  %s%s%s", formatResumePeriod(s.StartYear, s.StartMonth, s.EndYear, s.EndMonth, s.ToDate), s.Organization, joinResumeDetails(s.Role), joinResumeDetails(s.Description))
	})
	renderPDFText(pdf, "自我评价", resume.Specialty, style)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderResumeHeader(pdf *fpdf.Fpdf, resume *hrcModel.Resume, style pdfStyle) {
	pdf.SetTextColor(31, 41, 55)
	pdf.SetFont("cjk", "", style.nameSize)
	name := strings.TrimSpace(resume.FullName)
	if name == "" {
		name = "个人简历"
	}
	if style.centerName {
		pdf.CellFormat(0, style.nameSize*0.5, name, "", 1, "C", false, 0, "")
	} else {
		pdf.CellFormat(0, style.nameSize*0.5, name, "", 1, "L", false, 0, "")
	}
	// 联系信息行
	contacts := []string{}
	if resume.SexCN != "" {
		contacts = append(contacts, resume.SexCN)
	}
	if resume.Birthdate > 0 {
		contacts = append(contacts, strconv.Itoa(int(resume.Birthdate))+"年")
	}
	if resume.EducationCN != "" {
		contacts = append(contacts, resume.EducationCN)
	}
	if resume.ExperienceCN != "" {
		contacts = append(contacts, resume.ExperienceCN)
	}
	if resume.Telephone != "" {
		contacts = append(contacts, resume.Telephone)
	}
	if resume.Email != "" {
		contacts = append(contacts, resume.Email)
	}
	pdf.SetFont("cjk", "", style.bodySize)
	pdf.SetTextColor(71, 85, 105)
	align := "L"
	if style.centerName {
		align = "C"
	}
	pdf.CellFormat(0, style.lineH, strings.Join(contacts, " | "), "", 1, align, false, 0, "")
	pdf.SetDrawColor(style.accent[0], style.accent[1], style.accent[2])
	pdf.SetLineWidth(0.6)
	y := pdf.GetY() + 1
	pdf.Line(18, y, 192, y)
	pdf.SetY(y + 2)
}

func renderIntention(pdf *fpdf.Fpdf, resume *hrcModel.Resume, style pdfStyle) {
	rows := []string{}
	if resume.IntentionJobs != "" {
		rows = append(rows, "期望职位："+resume.IntentionJobs)
	}
	if resume.DistrictCN != "" {
		rows = append(rows, "期望地区："+resume.DistrictCN)
	}
	if w := formatWageRange(resume.WageMin, resume.WageMax); w != "" {
		rows = append(rows, "期望薪资："+w)
	}
	if len(rows) == 0 {
		return
	}
	renderPDFSectionTitle(pdf, "求职意向", style)
	pdf.SetFont("cjk", "", style.bodySize)
	pdf.SetTextColor(51, 65, 85)
	for _, row := range rows {
		pdf.MultiCell(0, style.lineH, row, "", "L", false)
	}
}

func renderPDFSectionTitle(pdf *fpdf.Fpdf, title string, style pdfStyle) {
	pdf.Ln(2.5)
	if style.filledBar {
		pdf.SetFillColor(style.accent[0], style.accent[1], style.accent[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("cjk", "", style.titleSize)
		pdf.CellFormat(0, style.lineH+2, "  "+title, "", 1, "L", true, 0, "")
		pdf.SetTextColor(51, 65, 85)
	} else {
		pdf.SetTextColor(style.accent[0], style.accent[1], style.accent[2])
		pdf.SetFont("cjk", "", style.titleSize)
		pdf.CellFormat(0, style.lineH+1, title, "", 1, "L", false, 0, "")
		pdf.SetDrawColor(style.accent[0], style.accent[1], style.accent[2])
		pdf.SetLineWidth(0.4)
		y := pdf.GetY()
		pdf.Line(18, y, 192, y)
		pdf.SetTextColor(51, 65, 85)
	}
	pdf.Ln(1)
}

func renderPDFEntries(pdf *fpdf.Fpdf, title string, count int, style pdfStyle, entry func(int) string) {
	if count == 0 {
		return
	}
	renderPDFSectionTitle(pdf, title, style)
	pdf.SetFont("cjk", "", style.bodySize)
	pdf.SetTextColor(51, 65, 85)
	for i := 0; i < count; i++ {
		text := strings.TrimSpace(entry(i))
		if text == "" {
			continue
		}
		pdf.MultiCell(0, style.lineH, text, "", "L", false)
	}
}

func renderPDFText(pdf *fpdf.Fpdf, title, text string, style pdfStyle) {
	if strings.TrimSpace(text) == "" {
		return
	}
	renderPDFSectionTitle(pdf, title, style)
	pdf.SetFont("cjk", "", style.bodySize)
	pdf.SetTextColor(51, 65, 85)
	pdf.MultiCell(0, style.lineH, text, "", "L", false)
}

// resumePDFFilename 生成下载文件名（形如「张三的简历.pdf」）
func resumePDFFilename(resume *hrcModel.Resume) string {
	name := strings.TrimSpace(resume.FullName)
	if name == "" {
		name = fmt.Sprintf("resume-%d", resume.ID)
	}
	return name + "的简历.pdf"
}
