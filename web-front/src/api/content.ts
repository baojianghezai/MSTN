// 内容/前台公开接口封装（03-cms.md）
import request from '@/utils/request'
import type { ApiResponse, ArticleItem, Categories, CategoryItem, NavItem } from '@/types/api'

/** 前台导航列表（display=1，按 sort 升序） */
export const getNavigations = () =>
  request.get<ApiResponse<NavItem[]>>('/navigations')

/** 分类聚合（education/experience/wage/trade/district） */
export const getCategories = () =>
  request.get<ApiResponse<Categories>>('/categories')

/** 地区目录按父级逐层读取，避免公共分类接口返回完整地区树。 */
export const getDistricts = (parentId = 0) =>
  request.get<ApiResponse<CategoryItem[]>>('/categories/districts', { params: { parentId } })

/** 内容列表（#17：type 1=资讯 2=招聘会 3=帮助） */
export const getArticles = (params: { type: number; page?: number; pageSize?: number }) =>
  request.get<ApiResponse<{ list: ArticleItem[]; total: number; page: number; pageSize: number }>>(
    '/articles',
    { params }
  )

/** 内容详情（#17） */
export const getArticle = (id: number) =>
  request.get<ApiResponse<ArticleItem>>(`/articles/${id}`)
