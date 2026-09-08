// 内容/前台公开接口封装（03-cms.md）
import request from '@/utils/request'
import type { ApiResponse, Categories, CategoryItem, NavItem } from '@/types/api'

/** 前台导航列表（display=1，按 sort 升序） */
export const getNavigations = () =>
  request.get<ApiResponse<NavItem[]>>('/navigations')

/** 分类聚合（education/experience/wage/trade/district） */
export const getCategories = () =>
  request.get<ApiResponse<Categories>>('/categories')

/** 地区目录按父级逐层读取，避免公共分类接口返回完整地区树。 */
export const getDistricts = (parentId = 0) =>
  request.get<ApiResponse<CategoryItem[]>>('/categories/districts', { params: { parentId } })
