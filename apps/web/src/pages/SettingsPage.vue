<template>
  <div class="settings-page">
    <div class="page-top-bar card-base">
      <div class="title-group">
        <h2 class="page-title">Cài đặt hệ thống</h2>
        <span class="page-sub">Quản lý tài khoản, cấu hình dung lượng, trí tuệ nhân tạo AI và thông số xử lý video</span>
      </div>
    </div>

    <div class="settings-grid">
      <!-- AI Gemini Configuration Card (Featured) -->
      <div class="settings-card card-base ai-card">
        <div class="card-header-with-badge">
          <div class="ai-title-wrap">
            <div class="ai-icon-box">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83"/>
              </svg>
            </div>
            <div>
              <h3 class="card-title">Cấu hình Google Gemini AI</h3>
              <span class="card-desc">Tự động dịch tiêu đề, mô tả và tạo phụ đề video từ tiếng Trung/nước ngoài sang tiếng Việt</span>
            </div>
          </div>
          <span
            class="status-pill"
            :class="isConfigured ? 'status-pill--active' : 'status-pill--inactive'"
          >
            {{ isConfigured ? '● Đã kết nối' : '○ Chưa thiết lập' }}
          </span>
        </div>

        <div class="ai-config-body">
          <div class="form-row">
            <label class="form-label">Mô hình AI sử dụng:</label>
            <div class="model-badge-box">
              <span class="model-badge">{{ currentModel || 'gemini-3.5-flash-lite' }}</span>
              <span class="model-desc">Hỗ trợ nhận diện giọng nói đa phương thức (Audio Multimodal) tốc độ cao & tối ưu chi phí</span>
            </div>
          </div>

          <div class="form-row">
            <div class="label-with-link">
              <label class="form-label">Gemini API Key:</label>
              <a
                href="https://aistudio.google.com/app/apikey"
                target="_blank"
                rel="noopener noreferrer"
                class="get-key-link"
              >
                Lấy API Key miễn phí tại Google AI Studio ↗
              </a>
            </div>

            <div class="api-key-input-wrap">
              <input
                :type="showKey ? 'text' : 'password'"
                v-model="apiKeyInput"
                :placeholder="isConfigured && maskedKey ? `Đang dùng: ${maskedKey}` : 'Dán khóa AIzaSy... vào đây'"
                class="form-input key-input"
              />
              <button
                type="button"
                class="toggle-show-btn"
                @click="showKey = !showKey"
                :title="showKey ? 'Ẩn' : 'Hiện'"
              >
                <svg v-if="showKey" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                  <line x1="1" y1="1" x2="23" y2="23"/>
                </svg>
                <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                  <circle cx="12" cy="12" r="3"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Test / Save Feedback Alert -->
          <div v-if="testResult" class="feedback-alert" :class="testResult.success ? 'feedback-alert--success' : 'feedback-alert--error'">
            <span class="alert-icon">{{ testResult.success ? '✓' : '⚠' }}</span>
            <span>{{ testResult.message }}</span>
          </div>

          <div v-if="saveSuccessMsg" class="feedback-alert feedback-alert--success">
            <span class="alert-icon">✓</span>
            <span>{{ saveSuccessMsg }}</span>
          </div>

          <!-- Actions -->
          <div class="ai-actions-row">
            <button
              class="btn-action btn-test"
              :disabled="isTesting || isSaving"
              @click="handleTestConnection"
            >
              <svg v-if="isTesting" class="spinner-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" stroke-opacity="0.25"/>
                <path d="M12 2a10 10 0 0 1 10 10"/>
              </svg>
              <span>{{ isTesting ? 'Đang kiểm tra...' : 'Kiểm tra kết nối' }}</span>
            </button>

            <button
              class="btn-action btn-save-primary"
              :disabled="isSaving || (!apiKeyInput.trim() && !isConfigured)"
              @click="handleSaveKey"
            >
              <svg v-if="isSaving" class="spinner-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" stroke-opacity="0.25"/>
                <path d="M12 2a10 10 0 0 1 10 10"/>
              </svg>
              <span>{{ isSaving ? 'Đang lưu...' : 'Lưu cấu hình API' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Profile Card -->
      <div class="settings-card card-base">
        <h3 class="card-title">Thông tin tài khoản</h3>
        <div class="profile-row">
          <img src="https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=120&q=80" alt="Avatar" class="avatar-lg" />
          <div class="profile-meta">
            <span class="profile-name">Big Hải</span>
            <span class="profile-badge">Admin</span>
            <span class="profile-email">bighai.admin@cutvideomxh.vn</span>
          </div>
        </div>
      </div>

      <!-- Storage Card -->
      <div class="settings-card card-base">
        <h3 class="card-title">Dung lượng bộ nhớ lưu trữ</h3>
        <div class="storage-details">
          <div class="storage-text-row">
            <span>Đã dùng: <strong>{{ videoStore.storageFormatted }}</strong></span>
            <span>Tổng gói: <strong>50.0 GB</strong></span>
          </div>
          <div class="storage-bar-track">
            <div class="storage-bar-fill" :style="{ width: videoStore.storagePercent + '%' }"></div>
          </div>
          <span class="storage-hint">Bạn còn trống {{ videoStore.storageFreeFormatted }} để lưu thêm video và các đoạn cắt.</span>
        </div>
      </div>

      <!-- Export Defaults -->
      <div class="settings-card card-base">
        <h3 class="card-title">Cấu hình xuất video mặc định</h3>
        <div class="config-form">
          <div class="form-row">
            <label class="form-label">Định dạng xuất:</label>
            <select class="form-select">
              <option value="mp4">MP4 (H.264 / AAC) - Khuyến nghị</option>
              <option value="webm">WebM (VP9 / Opus)</option>
            </select>
          </div>
          <div class="form-row">
            <label class="form-label">Chất lượng xuất:</label>
            <select class="form-select">
              <option value="original">Original (Giữ nguyên gốc / Stream copy - Nhanh nhất)</option>
              <option value="1080p">1080p Full HD</option>
              <option value="720p">720p HD</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Worker & Engine Status -->
      <div class="settings-card card-base">
        <h3 class="card-title">Trạng thái máy chủ xử lý</h3>
        <div class="engine-status-list">
          <div class="engine-item">
            <span class="engine-name">FFmpeg Transcoder Engine</span>
            <span class="engine-badge badge--online">● Trực tuyến</span>
          </div>
          <div class="engine-item">
            <span class="engine-name">Google Gemini 1.5 Flash</span>
            <span class="engine-badge" :class="isConfigured ? 'badge--online' : 'badge--offline'">
              {{ isConfigured ? '● Sẵn sàng' : '○ Chưa thiết lập' }}
            </span>
          </div>
          <div class="engine-item">
            <span class="engine-name">Go Worker Media Queue</span>
            <span class="engine-badge badge--online">● Hoạt động</span>
          </div>
          <div class="engine-item">
            <span class="engine-name">SQLite / PostgreSQL Database</span>
            <span class="engine-badge badge--online">● Đã kết nối</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { aiService } from '@/services/ai.service'
import { useVideoStore } from '@/stores/video.store'

const videoStore = useVideoStore()

const apiKeyInput = ref('')
const showKey = ref(false)
const isConfigured = ref(false)
const maskedKey = ref('')

const isTesting = ref(false)
const isSaving = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)
const saveSuccessMsg = ref('')

const currentModel = ref('gemini-3.8-flash')

onMounted(async () => {
  await loadAIConfig()
  if (videoStore.videos.length === 0) {
    await videoStore.fetchAllVideos()
  }
})

async function loadAIConfig() {
  try {
    const res = await aiService.getAIConfig()
    isConfigured.value = res.has_key
    maskedKey.value = res.masked_key || ''
    if (res.model) {
      currentModel.value = res.model
    }
  } catch (err) {
    console.warn('Failed to load AI config:', err)
  }
}

async function handleTestConnection() {
  isTesting.value = true
  testResult.value = null
  saveSuccessMsg.value = ''
  try {
    const res = await aiService.testAIConnection(apiKeyInput.value.trim())
    testResult.value = res
  } catch (err: any) {
    testResult.value = {
      success: false,
      message: err.message || 'Lỗi kết nối tới Google Gemini API',
    }
  } finally {
    isTesting.value = false
  }
}

async function handleSaveKey() {
  if (!apiKeyInput.value.trim() && !isConfigured.value) return
  isSaving.value = true
  testResult.value = null
  saveSuccessMsg.value = ''
  try {
    const res = await aiService.updateAIConfig(apiKeyInput.value.trim())
    if (res.success) {
      saveSuccessMsg.value = 'Đã lưu khóa API Gemini thành công!'
      apiKeyInput.value = ''
      await loadAIConfig()
      setTimeout(() => {
        saveSuccessMsg.value = ''
      }, 4000)
    }
  } catch (err: any) {
    testResult.value = {
      success: false,
      message: err.message || 'Không thể lưu khóa API',
    }
  } finally {
    isSaving.value = false
  }
}
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-top-bar {
  padding: 16px 20px;
  background: #FFFFFF;
}

.title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.page-sub {
  font-size: 13px;
  color: var(--text-muted, #64748B);
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.settings-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: #FFFFFF;
}

/* AI Featured Card */
.ai-card {
  grid-column: span 2;
  border: 1px solid rgba(139, 92, 246, 0.3);
  background: linear-gradient(180deg, #FAF8FF 0%, #FFFFFF 100%);
  box-shadow: 0 4px 14px rgba(139, 92, 246, 0.08);
}

.card-header-with-badge {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.ai-title-wrap {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.ai-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #8B5CF6 0%, #6366F1 100%);
  color: #FFFFFF;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.3);
}

.card-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary, #0F172A);
  margin: 0;
}

.card-desc {
  font-size: 13px;
  color: var(--text-secondary, #64748B);
  margin-top: 3px;
  display: block;
}

.status-pill {
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 12px;
  white-space: nowrap;
}

.status-pill--active {
  background: #DEF7EC;
  color: #03543F;
}

.status-pill--inactive {
  background: #F3F4F6;
  color: #6B7280;
}

.ai-config-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.model-badge-box {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.model-badge {
  background: rgba(139, 92, 246, 0.12);
  color: #6D28D9;
  font-family: monospace;
  font-size: 13px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 6px;
}

.model-desc {
  font-size: 12px;
  color: #64748B;
}

.label-with-link {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.get-key-link {
  font-size: 12px;
  color: #6D28D9;
  font-weight: 600;
  text-decoration: none;
}

.get-key-link:hover {
  text-decoration: underline;
}

.api-key-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.key-input {
  width: 100%;
  height: 40px;
  padding: 0 42px 0 12px;
  font-family: monospace;
  font-size: 13px;
  border: 1px solid var(--border, #CBD5E1);
  border-radius: 6px;
  outline: none;
  background: #FFFFFF;
}

.key-input:focus {
  border-color: #8B5CF6;
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.15);
}

.toggle-show-btn {
  position: absolute;
  right: 10px;
  background: transparent;
  border: none;
  color: #64748B;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.feedback-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  animation: fadeIn 0.15s ease-out;
}

.feedback-alert--success {
  background: #DEF7EC;
  color: #03543F;
  border: 1px solid #84E1BC;
}

.feedback-alert--error {
  background: #FDE8E8;
  color: #9B1C1C;
  border: 1px solid #F8B4B4;
}

.alert-icon {
  font-weight: bold;
}

.ai-actions-row {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn-action {
  height: 38px;
  padding: 0 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s;
}

.btn-test {
  background: #FFFFFF;
  border: 1px solid var(--border, #CBD5E1);
  color: #334155;
}

.btn-test:hover:not(:disabled) {
  background: #F8FAFC;
}

.btn-save-primary {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: #FFFFFF;
  border: none;
  box-shadow: 0 2px 6px rgba(124, 58, 237, 0.25);
}

.btn-save-primary:hover:not(:disabled) {
  opacity: 0.92;
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Standard Cards */
.profile-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.avatar-lg {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--primary-border, #E2E8F0);
}

.profile-meta {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.profile-name {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary, #0F172A);
}

.profile-badge {
  font-size: 11px;
  font-weight: 700;
  color: #FFFFFF;
  background: var(--primary, #1677FF);
  padding: 1px 7px;
  border-radius: 4px;
  width: fit-content;
}

.profile-email {
  font-size: 13px;
  color: var(--text-muted, #64748B);
}

.storage-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.storage-text-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-secondary, #475569);
}

.storage-bar-track {
  height: 8px;
  background: #E2E8F0;
  border-radius: 4px;
  overflow: hidden;
}

.storage-bar-fill {
  height: 100%;
  background: var(--primary, #1677FF);
  border-radius: 4px;
}

.storage-hint {
  font-size: 12px;
  color: var(--text-muted, #64748B);
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary, #475569);
}

.form-select {
  height: 38px;
  padding: 0 12px;
  font-size: 13px;
  background: var(--surface-secondary, #F8FAFC);
  border: 1px solid var(--border, #CBD5E1);
  border-radius: 6px;
}

.engine-status-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.engine-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--surface-secondary, #F8FAFC);
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
}

.engine-name {
  font-weight: 500;
  color: var(--text-primary, #0F172A);
}

.badge--online {
  color: #10B981;
  font-weight: 600;
  font-size: 12px;
}

.badge--offline {
  color: #9CA3AF;
  font-weight: 600;
  font-size: 12px;
}

@media (max-width: 900px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
  .ai-card {
    grid-column: span 1;
  }
}
</style>
