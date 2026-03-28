import axios from 'axios'
import { ElMessage } from 'element-plus'

const http = axios.create({
  baseURL: '/api',
  timeout: 60000,
})

http.interceptors.response.use(
  (response) => {
    if (response.data.code !== 0) {
      ElMessage.error(response.data.msg || 'Request failed')
      return Promise.reject(new Error(response.data.msg))
    }
    return response.data
  },
  (error) => {
    ElMessage.error(error.message || 'Network error')
    return Promise.reject(error)
  }
)

export default http

export interface ImageInfo {
  filename: string
  original_url: string
  width: number
  height: number
  size: number
  format: string
}

export interface CompressedResult {
  original: ImageInfo
  compressed: ImageInfo
  compression_ratio: number
  quality: number
}

export interface BatchResult {
  total: number
  success: number
  results: CompressedResult[]
  batch_id: string
}

export function uploadImage(file: File): Promise<{ data: ImageInfo }> {
  const formData = new FormData()
  formData.append('file', file)
  return http.post('/upload', formData)
}

export function compressImage(file: File, quality: number): Promise<{ data: CompressedResult }> {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('quality', quality.toString())
  return http.post('/compress', formData)
}

export function batchCompressImages(files: File[], quality: number): Promise<{ data: BatchResult }> {
  const formData = new FormData()
  files.forEach((file) => {
    formData.append('files', file)
  })
  formData.append('quality', quality.toString())
  return http.post('/batch-compress', formData)
}

export function getDownloadUrl(filename: string): string {
  return `/api/download?filename=${encodeURIComponent(filename)}`
}

export function getPreviewUrl(filename: string): string {
  return `/api/preview?filename=${encodeURIComponent(filename)}`
}

export function getDownloadZipUrl(batchId: string): string {
  return `/api/download-zip?batch_id=${encodeURIComponent(batchId)}`
}