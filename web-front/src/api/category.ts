import request from '@/utils/request'
import type {ApiResponse, CategoryTreeNode} from "@/types/api"

// 首页获取职位列表
export const getJobList = () =>
    request.get<ApiResponse<CategoryTreeNode[]>>('category/jobstree')