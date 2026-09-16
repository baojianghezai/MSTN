package hrc

import (
	"strconv"

	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type ResumeApi struct{}

// CreateResume 创建简历（09 §4.2 #48；6 子表随主表同事务批量写入，项目经历限 6 条，首份简历自动置默认）
// @Tags HrcResume
// @Summary 创建简历
// @Accept application/json
// @Produce json
// @Param data body ResumeRequest true "简历参数（主表可编辑字段 + educations/work/language/training/credent/projects 子表数组）"
// @Success 200 {object} Response{data=map[string]uint64}
// @Router /api/v1/personal/resumes [post]
func (a *ResumeApi) CreateResume(c *gin.Context) {
	var req ResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	id, err := hrcService.ServiceGroupApp.ResumeService.CreateResume(c.Request.Context(), uid, resumeFromRequest(req), resumeSubTablesFromRequest(req))
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, gin.H{"id": id})
}

// List 我的简历列表（09 §4.2 #49；仅本人未软删；默认简历在前，按创建时间倒序；轻量字段不返子表）
// @Tags HrcResume
// @Summary 我的简历列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} Response{data=response.PageResult{list=[]ResumeLite}}
// @Router /api/v1/personal/resumes [get]
func (a *ResumeApi) List(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	uid := middlewarehrc.GetMemberUID(c)
	list, total, err := hrcService.ServiceGroupApp.ResumeService.ListResumes(c.Request.Context(), uid, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	lite := make([]ResumeLite, 0, len(list))
	for _, r := range list {
		lite = append(lite, ResumeLite{
			ID:              r.ID,
			Title:           r.Title,
			CompletePercent: r.CompletePercent,
			Def:             r.Def,
			Display:         r.Display,
			AddTime:         r.AddTime,
			Refreshtime:     timePtrDeref(r.Refreshtime),
		})
	}
	OKWithData(c, response.PageResult{List: lite, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// GetResume 简历详情（编辑回显，09 §4.2 #50；主表 + 6 子表）
// @Tags HrcResume
// @Summary 简历详情（编辑回显）
// @Produce json
// @Param id path int true "简历 id"
// @Success 200 {object} Response{data=ResumeDetailData}
// @Router /api/v1/personal/resumes/{id} [get]
func (a *ResumeApi) GetResume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	resume, subs, err := hrcService.ServiceGroupApp.ResumeService.GetResume(c.Request.Context(), uid, id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, ResumeDetailData{
		Resume:     *resume,
		Projects:   subs.Project,
		Educations: subs.Education,
		Work:       subs.Work,
		Language:   subs.Language,
		Training:   subs.Training,
		Credent:    subs.Credent,
	})
}

// UpdateResume 编辑简历（09 §4.2 #51；6 子表全量替换，项目经历限 6 条）
// @Tags HrcResume
// @Summary 编辑简历
// @Accept application/json
// @Produce json
// @Param id path int true "简历 id"
// @Param data body ResumeRequest true "简历参数（主表可编辑字段 + 6 子表数组）"
// @Success 200 {object} Response
// @Router /api/v1/personal/resumes/{id} [put]
func (a *ResumeApi) UpdateResume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req ResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.ResumeService.UpdateResume(c.Request.Context(), uid, id, resumeFromRequest(req), resumeSubTablesFromRequest(req)); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// DeleteResume 删除简历（09 §4.2 #52；软删除：主表软删、子表保留、投递/下载记录保留；删默认简历自动升级最近一份）
// @Tags HrcResume
// @Summary 删除简历（软删除）
// @Produce json
// @Param id path int true "简历 id"
// @Success 200 {object} Response
// @Router /api/v1/personal/resumes/{id} [delete]
func (a *ResumeApi) DeleteResume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.ResumeService.DeleteResume(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// SetDisplay 公开/隐藏切换（09 §4.2 #54；1=公开 2=不公开）
// @Tags HrcResume
// @Summary 公开/隐藏切换
// @Accept application/json
// @Produce json
// @Param id path int true "简历 id"
// @Param data body ResumeDisplayRequest true "display：1=公开 2=不公开"
// @Success 200 {object} Response
// @Router /api/v1/personal/resumes/{id}/display [put]
func (a *ResumeApi) SetDisplay(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	var req ResumeDisplayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.ResumeService.SetDisplay(c.Request.Context(), uid, id, req.Display); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// SetDefault 设为默认简历（09 §4.2 #55；同 uid 内 def 互斥事务）
// @Tags HrcResume
// @Summary 设为默认简历
// @Produce json
// @Param id path int true "简历 id"
// @Success 200 {object} Response
// @Router /api/v1/personal/resumes/{id}/default [put]
func (a *ResumeApi) SetDefault(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	if err := hrcService.ServiceGroupApp.ResumeService.SetDefault(c.Request.Context(), uid, id); err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OK(c)
}

// GetCompleteness 完善度详情（09 §4.2 #57；权重 04 §2.2：基本信息 20/教育 20/工作 25/期望 15/其他 20；与主表 complete_percent 同源）
// @Tags HrcResume
// @Summary 完善度详情（缺失字段清单）
// @Produce json
// @Param id path int true "简历 id"
// @Success 200 {object} Response{data=hrcService.CompletenessResult}
// @Router /api/v1/personal/resumes/{id}/completeness [get]
func (a *ResumeApi) GetCompleteness(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := middlewarehrc.GetMemberUID(c)
	result, err := hrcService.ServiceGroupApp.ResumeService.GetCompleteness(c.Request.Context(), uid, id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, result)
}

// resumeFromRequest 请求 DTO → 简历主表模型（服务端管理字段不在此映射）
func resumeFromRequest(req ResumeRequest) *hrcModel.Resume {
	return &hrcModel.Resume{
		Title:         req.Title,
		FullName:      req.FullName,
		Sex:           req.Sex,
		SexCN:         req.SexCN,
		Birthdate:     req.Birthdate,
		Residence:     req.Residence,
		Education:     req.Education,
		EducationCN:   req.EducationCN,
		Major:         req.Major,
		MajorCN:       req.MajorCN,
		Experience:    req.Experience,
		ExperienceCN:  req.ExperienceCN,
		District:      req.District,
		DistrictCN:    req.DistrictCN,
		WageMin:       req.WageMin,
		WageMax:       req.WageMax,
		IntentionJobs: req.IntentionJobs,
		Specialty:     req.Specialty,
		Telephone:     req.Telephone,
		Email:         req.Email,
		DisplayName:   req.DisplayName,
		Current:       req.Current,
		CurrentCN:     req.CurrentCN,
		MobileAudit:   req.MobileAudit,
		Talent:        req.Talent,
		Entrust:       req.Entrust,
	}
}

// resumeSubTablesFromRequest 请求 DTO → 9 子表批量结构（子表 id/pid/uid 由服务端盖章，不信任提交值）
func resumeSubTablesFromRequest(req ResumeRequest) *hrcService.ResumeSubTables {
	subs := &hrcService.ResumeSubTables{}
	if req.Projects != nil {
		subs.Project = make([]hrcModel.ResumeProject, 0, len(req.Projects))
		for _, it := range req.Projects {
			subs.Project = append(subs.Project, hrcModel.ResumeProject{
				StartYear:   it.StartYear,
				StartMonth:  it.StartMonth,
				EndYear:     it.EndYear,
				EndMonth:    it.EndMonth,
				ToDate:      it.ToDate,
				ProjectName: it.ProjectName,
				Role:        it.Role,
				Description: it.Description,
			})
		}
	}
	if req.Educations != nil {
		subs.Education = make([]hrcModel.ResumeEducation, 0, len(req.Educations))
		for _, it := range req.Educations {
			subs.Education = append(subs.Education, hrcModel.ResumeEducation{
				StartYear:   it.StartYear,
				StartMonth:  it.StartMonth,
				EndYear:     it.EndYear,
				EndMonth:    it.EndMonth,
				ToDate:      it.ToDate,
				School:      it.School,
				Speciality:  it.Speciality,
				Education:   it.Education,
				EducationCN: it.EducationCN,
			})
		}
	}
	if req.Work != nil {
		subs.Work = make([]hrcModel.ResumeWork, 0, len(req.Work))
		for _, it := range req.Work {
			subs.Work = append(subs.Work, hrcModel.ResumeWork{
				StartYear:    it.StartYear,
				StartMonth:   it.StartMonth,
				EndYear:      it.EndYear,
				EndMonth:     it.EndMonth,
				ToDate:       it.ToDate,
				WorkType:     it.WorkType,
				CompanyName:  it.CompanyName,
				Jobs:         it.Jobs,
				Achievements: it.Achievements,
			})
		}
	}
	if req.Language != nil {
		subs.Language = make([]hrcModel.ResumeLanguage, 0, len(req.Language))
		for _, it := range req.Language {
			subs.Language = append(subs.Language, hrcModel.ResumeLanguage{
				Language:   it.Language,
				LanguageCN: it.LanguageCN,
				Level:      it.Level,
				LevelCN:    it.LevelCN,
			})
		}
	}
	if req.Training != nil {
		subs.Training = make([]hrcModel.ResumeTraining, 0, len(req.Training))
		for _, it := range req.Training {
			subs.Training = append(subs.Training, hrcModel.ResumeTraining{
				StartYear:   it.StartYear,
				StartMonth:  it.StartMonth,
				EndYear:     it.EndYear,
				EndMonth:    it.EndMonth,
				ToDate:      it.ToDate,
				Agency:      it.Agency,
				Course:      it.Course,
				Description: it.Description,
			})
		}
	}
	if req.Credent != nil {
		subs.Credent = make([]hrcModel.ResumeCredent, 0, len(req.Credent))
		for _, it := range req.Credent {
			subs.Credent = append(subs.Credent, hrcModel.ResumeCredent{
				Name:   it.Name,
				Year:   it.Year,
				Month:  it.Month,
				Images: it.Images,
			})
		}
	}
	if req.Skill != nil {
		subs.Skill = make([]hrcModel.ResumeSkill, 0, len(req.Skill))
		for _, it := range req.Skill {
			subs.Skill = append(subs.Skill, hrcModel.ResumeSkill{
				Name:  it.Name,
				Level: it.Level,
			})
		}
	}
	if req.Portfolio != nil {
		subs.Portfolio = make([]hrcModel.ResumePortfolio, 0, len(req.Portfolio))
		for _, it := range req.Portfolio {
			subs.Portfolio = append(subs.Portfolio, hrcModel.ResumePortfolio{
				Title:       it.Title,
				Description: it.Description,
				URL:         it.URL,
			})
		}
	}
	if req.StudentLeader != nil {
		subs.StudentLeader = make([]hrcModel.ResumeStudentLeader, 0, len(req.StudentLeader))
		for _, it := range req.StudentLeader {
			subs.StudentLeader = append(subs.StudentLeader, hrcModel.ResumeStudentLeader{
				Organization: it.Organization,
				Role:         it.Role,
				StartYear:    it.StartYear,
				StartMonth:   it.StartMonth,
				EndYear:      it.EndYear,
				EndMonth:     it.EndMonth,
				ToDate:       it.ToDate,
				Description:  it.Description,
			})
		}
	}
	return subs
}
