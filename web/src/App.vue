<template>
  <div class="app-container">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <div class="logo">
          <img src="/favicon.svg" alt="Logo" class="logo-icon" />
          <span class="logo-text">{{ t('app.title') }}</span>
        </div>
        <p class="subtitle">{{ t('app.subtitle') }}</p>
        <!-- Language Switcher -->
        <div class="lang-switcher">
          <el-dropdown @command="changeLang">
            <el-button type="primary" plain>
              {{ currentLangLabel }} <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="zh" :class="{ active: currentLang === 'zh' }">
                  {{ t('lang.zh') }}
                </el-dropdown-item>
                <el-dropdown-item command="en" :class="{ active: currentLang === 'en' }">
                  {{ t('lang.en') }}
                </el-dropdown-item>
                <el-dropdown-item command="ja" :class="{ active: currentLang === 'ja' }">
                  {{ t('lang.ja') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <!-- Upload Section -->
      <section class="upload-section">
        <div
          class="upload-area"
          :class="{ 'drag-over': isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="handleDrop"
          @click="triggerFileInput"
        >
          <input
            ref="fileInput"
            type="file"
            multiple
            accept="image/jpeg,image/png,image/webp"
            style="display: none"
            @change="handleFileSelect"
          />
          <div class="upload-icon">
            <el-icon :size="48"><Upload /></el-icon>
          </div>
          <p class="upload-text">{{ t('upload.title') }}</p>
          <p class="upload-hint">{{ t('upload.hint') }}</p>
        </div>
      </section>

      <!-- Settings Section -->
      <section class="settings-section" v-if="pendingFiles.length > 0">
        <div class="settings-card">
          <h3>{{ t('settings.title') }}</h3>
          <div class="quality-settings">
            <div class="quality-presets">
              <el-button
                :type="quality === 90 ? 'success' : 'default'"
                @click="quality = 90"
              >
                {{ t('settings.qualityHigh') }}
              </el-button>
              <el-button
                :type="quality === 75 ? 'success' : 'default'"
                @click="quality = 75"
              >
                {{ t('settings.qualityMedium') }}
              </el-button>
              <el-button
                :type="quality === 60 ? 'success' : 'default'"
                @click="quality = 60"
              >
                {{ t('settings.qualityLow') }}
              </el-button>
            </div>
            <div class="quality-slider">
              <span>{{ t('settings.qualityCustom') }}</span>
              <el-slider
                v-model="quality"
                :min="10"
                :max="100"
                :step="5"
                show-input
                class="quality-slider-input"
              />
            </div>
          </div>
          <div class="auto-download-option">
            <el-checkbox v-model="autoDownload">{{ t('settings.autoDownload') }}</el-checkbox>
          </div>
          <div class="action-buttons">
            <el-button type="primary" size="large" @click="startCompression" :loading="isProcessing">
              {{ isBatchMode ? t('settings.batchCompress') : t('settings.compress') }} ({{ pendingFiles.length }} {{ t('settings.files') }})
            </el-button>
            <el-button size="large" @click="clearAll">{{ t('settings.clearAll') }}</el-button>
          </div>
        </div>

        <!-- Pending Files Preview -->
        <div class="pending-files">
          <div v-for="(file, index) in pendingFiles" :key="index" class="pending-file">
            <el-image
              :src="getFilePreview(file)"
              :alt="file.name"
              fit="cover"
              :preview-src-list="pendingFiles.map(f => getFilePreview(f))"
              :initial-index="index"
              class="pending-image"
            />
            <div class="pending-file-info">
              <span class="file-name">{{ file.name }}</span>
              <span class="file-size">{{ formatSize(file.size) }}</span>
            </div>
            <el-button type="danger" size="small" circle @click="removeFile(index)">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
        </div>
      </section>

      <!-- Results Section -->
      <section class="results-section" v-if="results.length > 0">
        <h2>{{ t('results.title') }}</h2>
        <div class="results-grid">
          <div v-for="(result, index) in results" :key="index" class="result-card">
            <!-- Image Comparison -->
            <div class="comparison-container">
              <div class="comparison-item">
                <span class="comparison-label">Original</span>
                <el-image
                  :src="getPreviewUrl(result.original.filename)"
                  alt="Original"
                  fit="cover"
                  :preview-src-list="[getPreviewUrl(result.original.filename), getPreviewUrl(result.compressed.filename)]"
                  :initial-index="0"
                  class="comparison-image"
                />
              </div>
              <div class="comparison-arrow">
                <el-icon :size="24"><Right /></el-icon>
              </div>
              <div class="comparison-item">
                <span class="comparison-label">Compressed</span>
                <el-image
                  :src="getPreviewUrl(result.compressed.filename)"
                  alt="Compressed"
                  fit="cover"
                  :preview-src-list="[getPreviewUrl(result.original.filename), getPreviewUrl(result.compressed.filename)]"
                  :initial-index="1"
                  class="comparison-image"
                />
              </div>
            </div>

            <!-- Stats -->
            <div class="result-stats">
              <div class="stat-row">
                <span class="stat-label">{{ t('results.dimensions') }}</span>
                <span class="stat-value">{{ result.original.width }}x{{ result.original.height }}</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">{{ t('results.originalSize') }}</span>
                <span class="stat-value">{{ formatSize(result.original.size) }}</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">{{ t('results.compressedSize') }}</span>
                <span class="stat-value compressed">{{ formatSize(result.compressed.size) }}</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">{{ t('results.compressionRatio') }}</span>
                <span class="stat-value ratio" :class="{ good: result.compression_ratio < 70 }">
                  {{ result.compression_ratio.toFixed(1) }}%
                </span>
              </div>
            </div>

            <!-- Download Button -->
            <el-button
              type="success"
              size="large"
              class="download-btn"
              @click="downloadFile(result.compressed.filename)"
            >
              <el-icon><Download /></el-icon>
              {{ t('results.download') }}
            </el-button>
          </div>
        </div>

        <!-- Batch Download -->
        <div class="batch-download" v-if="results.length > 1">
          <el-button type="success" size="large" @click="downloadAll">
            <el-icon><Download /></el-icon>
            {{ t('results.downloadAllZip', { count: results.length }) }}
          </el-button>
        </div>
      </section>

      <!-- History Section -->
      <section class="history-section" v-if="history.length > 0">
        <div class="history-header">
          <h2><el-icon><Clock /></el-icon> {{ t('history.title') }}</h2>
          <el-button type="danger" size="small" @click="clearHistory">{{ t('history.clearAll') }}</el-button>
        </div>
        <p class="history-note">{{ t('history.note') }}</p>
        <div class="history-list">
          <div v-for="(item, idx) in history" :key="item.timestamp" class="history-item">
            <div class="history-item-header">
              <span class="history-time">{{ formatTime(item.timestamp) }}</span>
              <span class="history-count">{{ item.results.length }} {{ t('history.files') }}</span>
              <el-button type="danger" size="small" circle @click="removeFromHistory(idx)">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
            <div class="history-results">
              <div v-for="(result, ridx) in item.results" :key="ridx" class="history-result-mini">
                <el-image
                  :src="getPreviewUrl(result.compressed.filename)"
                  alt="Compressed"
                  fit="cover"
                  :preview-src-list="item.results.map(r => getPreviewUrl(r.compressed.filename))"
                  :initial-index="ridx"
                  class="history-image"
                />
                <div class="mini-info">
                  <span>{{ formatSize(result.compressed.size) }}</span>
                  <span class="ratio">{{ result.compression_ratio.toFixed(1) }}%</span>
                </div>
                <el-button type="success" size="small" @click="downloadFile(result.compressed.filename)">
                  <el-icon><Download /></el-icon>
                </el-button>
              </div>
            </div>
            <div v-if="item.results.length > 1 && item.batchId" class="history-zip">
              <el-button type="primary" size="small" @click="downloadHistoryZip(item.batchId)">
                {{ t('history.downloadZip') }}
              </el-button>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="footer">
      <p>{{ t('footer', { year: copyrightYear }) }}</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Upload, Close, Right, Download, Clock, ArrowDown } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { compressImage, batchCompressImages, getDownloadUrl, getPreviewUrl, getDownloadZipUrl } from './api'
import type { CompressedResult } from './api'
import { setLocale, getLocale } from './i18n'
import type { LocaleType } from './i18n'

const { t } = useI18n()

interface HistoryItem {
  results: CompressedResult[]
  batchId: string
  timestamp: number
}

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const pendingFiles = ref<File[]>([])
const quality = ref(75)
const autoDownload = ref(false)
const isProcessing = ref(false)
const results = ref<CompressedResult[]>([])
const batchId = ref('')
const history = ref<HistoryItem[]>([])
const currentLang = ref<LocaleType>(getLocale())

const isBatchMode = computed(() => pendingFiles.value.length > 1)

const currentLangLabel = computed(() => {
  const labels: Record<LocaleType, string> = {
    zh: '中文',
    en: 'English',
    ja: '日本語'
  }
  return labels[currentLang.value] || 'English'
})

const copyrightYear = computed(() => {
  const startYear = 2026
  const currentYear = new Date().getFullYear()
  if (currentYear <= startYear) {
    return '© 2026'
  }
  return `© 2026 - ${currentYear}`
})

function changeLang(lang: LocaleType): void {
  setLocale(lang)
  currentLang.value = lang
}

const MAX_FILES = 10
const MAX_SIZE = 20 * 1024 * 1024 // 20MB
const MAX_HISTORY = 100
const STORAGE_KEY = 'image_compressor_history'

function loadHistory(): void {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      history.value = JSON.parse(stored)
    }
  } catch {
    history.value = []
  }
}

function saveHistory(): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(history.value))
  } catch {
    // Storage full or unavailable
  }
}

onMounted(() => {
  loadHistory()
})

function triggerFileInput(): void {
  fileInput.value?.click()
}

function handleFileSelect(event: Event): void {
  const target = event.target as HTMLInputElement
  if (target.files) {
    addFiles(Array.from(target.files))
  }
}

function handleDrop(event: DragEvent): void {
  isDragging.value = false
  if (event.dataTransfer?.files) {
    addFiles(Array.from(event.dataTransfer.files))
  }
}

function addFiles(files: File[]): void {
  const validFiles = files.filter((file) => {
    if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
      ElMessage.warning(`${file.name}: ${t('upload.invalidFormat')}`)
      return false
    }
    if (file.size > MAX_SIZE) {
      ElMessage.warning(`${file.name}: ${t('upload.fileTooLarge')}`)
      return false
    }
    return true
  })

  const remaining = MAX_FILES - pendingFiles.value.length
  if (validFiles.length > remaining) {
    ElMessage.warning(t('upload.maxFilesReached', { count: remaining }))
    pendingFiles.value.push(...validFiles.slice(0, remaining))
  } else {
    pendingFiles.value.push(...validFiles)
  }
}

function removeFile(index: number): void {
  pendingFiles.value.splice(index, 1)
}

function clearAll(): void {
  pendingFiles.value = []
  results.value = []
}

function clearHistory(): void {
  history.value = []
  localStorage.removeItem(STORAGE_KEY)
}

function removeFromHistory(index: number): void {
  history.value.splice(index, 1)
  saveHistory()
}

function formatTime(timestamp: number): string {
  return new Date(timestamp).toLocaleString()
}

function getFilePreview(file: File): string {
  return URL.createObjectURL(file)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
}

async function startCompression(): Promise<void> {
  if (pendingFiles.value.length === 0) return

  isProcessing.value = true
  results.value = []
  batchId.value = ''

  try {
    if (pendingFiles.value.length === 1) {
      const response = await compressImage(pendingFiles.value[0], quality.value)
      results.value = [response.data]
    } else {
      const response = await batchCompressImages(pendingFiles.value, quality.value)
      results.value = response.data.results
      batchId.value = response.data.batch_id
      ElMessage.success(t('messages.success', { success: response.data.success, total: response.data.total }))
    }

    // Save to history
    if (results.value.length > 0) {
      const historyItem: HistoryItem = {
        results: results.value,
        batchId: batchId.value,
        timestamp: Date.now()
      }
      history.value.unshift(historyItem)
      if (history.value.length > MAX_HISTORY) {
        history.value = history.value.slice(0, MAX_HISTORY)
      }
      saveHistory()

      // Auto download if enabled
      if (autoDownload.value) {
        if (results.value.length > 1 && batchId.value) {
          downloadZip()
        } else if (results.value.length === 1) {
          downloadFile(results.value[0].compressed.filename)
        }
      }
    }

    pendingFiles.value = []
  } catch {
    // Error handled in interceptor
  } finally {
    isProcessing.value = false
  }
}

function downloadFile(filename: string): void {
  const link = document.createElement('a')
  link.href = getDownloadUrl(filename)
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function downloadZip(): void {
  const link = document.createElement('a')
  link.href = getDownloadZipUrl(batchId.value)
  link.download = `compressed_images_${batchId.value}.zip`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function downloadHistoryZip(bid: string): void {
  const link = document.createElement('a')
  link.href = getDownloadZipUrl(bid)
  link.download = `compressed_images_${bid}.zip`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

async function downloadAll(): Promise<void> {
  downloadZip()
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  min-height: 100vh;
}

.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Header */
.header {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  padding: 40px 20px;
  text-align: center;
  box-shadow: 0 4px 20px rgba(16, 185, 129, 0.3);
}

.header-content {
  max-width: 800px;
  margin: 0 auto;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 15px;
  margin-bottom: 10px;
}

.logo-icon {
  width: 48px;
  height: 48px;
}

.logo-text {
  font-size: 32px;
  font-weight: 700;
  color: white;
  letter-spacing: -0.5px;
}

.subtitle {
  color: rgba(255, 255, 255, 0.9);
  font-size: 18px;
}

.lang-switcher {
  margin-top: 15px;
}

.lang-switcher .el-button {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.5);
  color: white;
}

.lang-switcher .el-button:hover {
  background: rgba(255, 255, 255, 0.3);
  border-color: white;
}

.el-dropdown-menu .el-dropdown-item.active {
  color: #10b981;
  font-weight: 600;
}

/* Main Content */
.main-content {
  flex: 1;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 40px 20px;
}

/* Upload Section */
.upload-section {
  margin-bottom: 40px;
}

.upload-area {
  background: white;
  border: 3px dashed #10b981;
  border-radius: 16px;
  padding: 60px 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.upload-area:hover,
.upload-area.drag-over {
  background: #f0fdf4;
  border-color: #059669;
  transform: scale(1.01);
}

.upload-icon {
  color: #10b981;
  margin-bottom: 20px;
}

.upload-text {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 10px;
}

.upload-hint {
  color: #6b7280;
  font-size: 14px;
}

/* Settings Section */
.settings-section {
  margin-bottom: 40px;
}

.settings-card {
  background: white;
  border-radius: 16px;
  padding: 30px 40px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
  margin-bottom: 20px;
  max-width: 600px;
  width: 100%;
  margin-left: auto;
  margin-right: auto;
  text-align: center;
}

.settings-card h3 {
  color: #1f2937;
  margin-bottom: 25px;
  font-size: 18px;
  font-weight: 600;
}

.quality-settings {
  margin-bottom: 25px;
}

.quality-presets {
  display: flex;
  gap: 12px;
  margin-bottom: 25px;
  flex-wrap: wrap;
  justify-content: center;
}

.quality-presets .el-button {
  flex: 1;
  min-width: 120px;
  border-radius: 8px;
  font-weight: 500;
}

.quality-slider {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 15px;
  flex-wrap: wrap;
  margin-bottom: 10px;
  padding: 0 10px;
}

.quality-slider-input {
  flex: 1;
  min-width: 180px;
  max-width: 280px;
}

.auto-download-option {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #e5e7eb;
}

.auto-download-option .el-checkbox {
  color: #6b7280;
  font-size: 13px;
}

.action-buttons {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
  justify-content: center;
}

.action-buttons .el-button {
  min-width: 140px;
  border-radius: 8px;
  font-weight: 500;
}

/* Pending Files */
.pending-files {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 15px;
}

.pending-file {
  background: white;
  border-radius: 12px;
  padding: 10px;
  position: relative;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.pending-file img,
.pending-file .pending-image {
  width: 100%;
  height: 100px;
  object-fit: cover;
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
}

.pending-file-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.file-name {
  font-size: 12px;
  color: #374151;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 11px;
  color: #9ca3af;
}

.pending-file .el-button {
  position: absolute;
  top: 5px;
  right: 5px;
}

/* Results Section */
.results-section h2 {
  color: #1f2937;
  margin-bottom: 25px;
  font-size: 24px;
}

.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 25px;
}

.result-card {
  background: white;
  border-radius: 16px;
  padding: 25px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.comparison-container {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.comparison-item {
  flex: 1;
  text-align: center;
}

.comparison-label {
  display: block;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 8px;
  font-weight: 600;
}

.comparison-item img,
.comparison-item .comparison-image {
  width: 100%;
  height: 120px;
  object-fit: cover;
  border-radius: 8px;
  border: 2px solid #e5e7eb;
  cursor: pointer;
}

.comparison-arrow {
  color: #10b981;
  flex-shrink: 0;
}

.result-stats {
  background: #f9fafb;
  border-radius: 12px;
  padding: 15px;
  margin-bottom: 20px;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px solid #e5e7eb;
}

.stat-row:last-child {
  border-bottom: none;
}

.stat-label {
  color: #6b7280;
  font-size: 13px;
}

.stat-value {
  color: #1f2937;
  font-weight: 600;
  font-size: 13px;
}

.stat-value.compressed {
  color: #10b981;
}

.stat-value.ratio {
  color: #f59e0b;
}

.stat-value.ratio.good {
  color: #10b981;
}

.download-btn {
  width: 100%;
}

/* Batch Download */
.batch-download {
  margin-top: 30px;
  text-align: center;
}

/* Footer */
.footer {
  background: #1f2937;
  color: #9ca3af;
  text-align: center;
  padding: 25px;
  margin-top: auto;
}

/* History Section */
.history-section {
  margin-top: 40px;
  padding-top: 30px;
  border-top: 1px solid #e5e7eb;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.history-header h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #1f2937;
  font-size: 20px;
  margin: 0;
}

.history-note {
  color: #6b7280;
  font-size: 12px;
  margin-bottom: 20px;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.history-item {
  background: white;
  border-radius: 12px;
  padding: 15px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.history-item-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 12px;
}

.history-time {
  font-size: 13px;
  color: #6b7280;
}

.history-count {
  font-size: 12px;
  color: #10b981;
  background: #d1fae5;
  padding: 2px 8px;
  border-radius: 10px;
}

.history-results {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.history-result-mini {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #f9fafb;
  padding: 8px;
  border-radius: 8px;
}

.history-result-mini img,
.history-result-mini .history-image {
  width: 50px;
  height: 50px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
}

.mini-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12px;
}

.mini-info .ratio {
  color: #10b981;
  font-weight: 600;
}

.history-zip {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #e5e7eb;
}

/* Responsive */
@media (max-width: 768px) {
  .header {
    padding: 30px 15px;
  }

  .logo-text {
    font-size: 24px;
  }

  .subtitle {
    font-size: 14px;
  }

  .upload-area {
    padding: 40px 20px;
  }

  .upload-text {
    font-size: 16px;
  }

  .main-content {
    padding: 20px 15px;
  }

  .settings-card {
    padding: 20px 15px;
  }

  .quality-presets {
    flex-direction: column;
    gap: 8px;
    align-items: center;
  }

  .quality-presets .el-button {
    width: 100%;
    max-width: 280px;
    margin: 0;
    justify-content: center;
    flex: none;
  }

  .quality-slider {
    flex-direction: column;
    align-items: stretch;
  }

  .quality-slider-input {
    min-width: auto;
    max-width: none;
    width: 100%;
  }

  .auto-download-option {
    justify-content: flex-start;
  }

  .action-buttons {
    flex-direction: column;
    align-items: center;
  }

  .action-buttons .el-button {
    width: 100%;
    max-width: 280px;
  }

  .results-grid {
    grid-template-columns: 1fr;
  }

  .comparison-container {
    flex-direction: column;
  }

  .comparison-arrow {
    transform: rotate(90deg);
  }
}

/* Element Plus Override */
.el-button--success {
  --el-button-bg-color: #10b981;
  --el-button-border-color: #10b981;
  --el-button-hover-bg-color: #059669;
  --el-button-hover-border-color: #059669;
  --el-button-active-bg-color: #047857;
  --el-button-active-border-color: #047857;
}

.el-button--primary {
  --el-button-bg-color: #10b981;
  --el-button-border-color: #10b981;
  --el-button-hover-bg-color: #059669;
  --el-button-hover-border-color: #059669;
}

/* Quality preset buttons - default style */
.quality-presets .el-button--default {
  --el-button-bg-color: #f0fdf4;
  --el-button-border-color: #10b981;
  --el-button-text-color: #10b981;
  --el-button-hover-bg-color: #d1fae5;
  --el-button-hover-border-color: #059669;
  --el-button-hover-text-color: #059669;
}

/* Clear button - danger style */
.action-buttons .el-button--default {
  --el-button-bg-color: #fef2f2;
  --el-button-border-color: #ef4444;
  --el-button-text-color: #ef4444;
  --el-button-hover-bg-color: #fee2e2;
  --el-button-hover-border-color: #dc2626;
  --el-button-hover-text-color: #dc2626;
}
</style>