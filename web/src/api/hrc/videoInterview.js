// 视频面试接口封装（后台，05-video-interview.md）
import requestHrc from '@/utils/requestHrc'

/**
 * 视频面试列表（后台）
 * @param {Object} params { page, pageSize, keyword } keyword 跨职位名/公司名/简历姓名模糊搜索
 */
export const getVideoInterviews = (params) =>
  requestHrc({
    url: '/admin/video-interviews',
    method: 'get',
    params
  })
