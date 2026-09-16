import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface PublicResume {
  id: number
  title: string
  fullname: string
  sex: number
  sexCn: string
  birthdate: number
  education: number
  educationCn: string
  majorCn: string
  experience: number
  experienceCn: string
  district: string
  districtCn: string
  wageMin: number
  wageMax: number
  wageCn: string
  intentionJobs: string
  specialty: string
  photoImg: string
  completePercent: number
  talent: number
  refreshtime: string
}

export interface PublicResumeDetail {
  resume: PublicResume
  projects: Array<{
    id: number
    projectname: string
    role: string
    description: string
  }>
  educations: Array<{
    id: number
    school: string
    speciality: string
    educationCn: string
  }>
  work: Array<{
    id: number
    companyname: string
    jobs: string
    achievements: string
  }>
}

export interface TalentSearchParams {
  page: number
  pageSize: number
  keyword?: string
  district?: string
  education?: number
  experience?: number
  wageMin?: number
  wageMax?: number
}

export interface TalentUnlockedDetail {
  resume: {
    id: number
    uid: number
    title: string
    fullname: string
    telephone: string
    email: string
    intentionJobs: string
    districtCn: string
    educationCn: string
    experienceCn: string
    wageCn: string
    specialty: string
    wordResume: string
  }
}

export interface CompanyTalentItem {
  download: {
    id: number
    resumeId: number
    followUp: number
    downloadedAt: string
  }
  resume: TalentUnlockedDetail['resume']
}

export interface FavoriteTalentItem {
  favorite: {
    id: number
    resumeId: number
    addtime: string
  }
  resume: PublicResume
}

type PageData<T> = { list: T[]; total: number; page: number; pageSize: number }

/** #29 公开、脱敏简历搜索。 */
export const searchTalents = (params: TalentSearchParams) =>
  request.get<ApiResponse<PageData<PublicResume>>>('/resumes', { params })

/** #31 高级人才专区。 */
export const searchPremiumTalents = (params: TalentSearchParams) =>
  request.get<ApiResponse<PageData<PublicResume>>>('/resumes/talents', { params })

/** 公开脱敏简历详情。 */
export const getPublicTalent = (resumeId: number) =>
  request.get<ApiResponse<PublicResumeDetail>>(`/resumes/${resumeId}`)

/** 企业主动解锁候选人完整联系方式。 */
export const unlockTalent = (resumeId: number) =>
  request.post<ApiResponse<{ detail: TalentUnlockedDetail; newlyUnlocked: boolean }>>(
    `/company/talents/${resumeId}/unlock`
  )

/** 企业人才库。 */
export const getCompanyTalents = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<PageData<CompanyTalentItem>>>('/company/talents', { params })

/** 人才库跟进状态：0 待跟进，1 合适，2 不合适，3 待定，4 未接通。 */
export const setTalentFollowUp = (resumeId: number, followUp: number) =>
  request.put<ApiResponse>(`/company/talents/${resumeId}/follow-up`, { followUp })

/** 收藏公开候选人，不消耗下载额度。 */
export const favoriteTalent = (resumeId: number) =>
  request.post<ApiResponse>('/company/favorites', { resumeId })

/** 企业收藏的候选人列表，仍保持脱敏。 */
export const getFavoriteTalents = (params: { page: number; pageSize: number }) =>
  request.get<ApiResponse<PageData<FavoriteTalentItem>>>('/company/favorites', { params })

/** 取消企业收藏。 */
export const unfavoriteTalent = (favoriteId: number) =>
  request.delete<ApiResponse>(`/company/favorites/${favoriteId}`)
