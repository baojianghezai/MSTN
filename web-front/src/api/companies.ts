import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface PublicCompany {
  id: number
  uid: number
  companyname: string
  nature: number
  natureCn: string
  trade: number
  tradeCn: string
  district: string
  districtCn: string
  scale: number
  scaleCn: string
  logo: string
  shortName: string
  shortDesc: string
  tag: string
  refreshtime: number
  jobsCount: number
}

export interface PublicCompanyJob {
  id: number
  jobsName: string
  natureCn: string
  categoryCn: string
  districtCn: string
  education: number
  experience: number
  minwage: number
  maxwage: number
  negotiable: number
  amount: number
  emergency: number
  stick: number
  addtime: number
  refreshtime: number
}

export interface PublicCompanyDetail {
  company: PublicCompany
  contents: string
  address: string
  website: string
  jobs: PublicCompanyJob[]
  jobsTotal: number
}

export interface CompanySearchParams {
  page: number
  pageSize: number
  keyword?: string
  nature?: number
  trade?: number
  scale?: number
  district?: string
}

type PageData<T> = { list: T[]; total: number; page: number; pageSize: number }

export const searchCompanies = (params: CompanySearchParams) =>
  request.get<ApiResponse<PageData<PublicCompany>>>('/companies', { params })

export const getPublicCompany = (id: number) =>
  request.get<ApiResponse<PublicCompanyDetail>>(`/companies/${id}`)

export const getPublicCompanyJobs = (id: number, params: Pick<CompanySearchParams, 'page' | 'pageSize'>) =>
  request.get<ApiResponse<PageData<PublicCompanyJob>>>(`/companies/${id}/jobs`, { params })
