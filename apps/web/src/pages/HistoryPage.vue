<template>
  <div class="history-page">
    <!-- Top Bar -->
    <div class="page-top-bar card-base">
      <div class="title-group">
        <h2 class="page-title">Lịch sử thao tác ({{ historyStore.items.length }})</h2>
        <span class="page-sub">Theo dõi tất cả hoạt động cắt, lưu và tải video/đoạn cắt</span>
      </div>

      <AppButton variant="secondary" size="md" @click="historyStore.fetchHistory">
        <template #icon-left>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
        </template>
        Làm mới
      </AppButton>
    </div>

    <!-- History Table Card -->
    <div class="table-card card-base">
      <table class="history-table">
        <thead>
          <tr>
            <th>Thời gian</th>
            <th>Video</th>
            <th>Hành động</th>
            <th>Đoạn cắt</th>
            <th>Trạng thái</th>
            <th>Dung lượng</th>
            <th class="text-right">Thao tác</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in historyStore.items" :key="item.id">
            <td class="time-cell">{{ formatDate(item.created_at) }}</td>
            <td class="video-cell">
              <span class="video-name" :title="item.video_title">{{ item.video_title }}</span>
            </td>
            <td>
              <span class="action-tag">{{ item.action }}</span>
            </td>
            <td class="clip-cell">
              <span class="clip-text">{{ item.clip_name || 'Toàn bộ video' }}</span>
            </td>
            <td>
              <span :class="['status-badge', `status-badge--${getStatusClass(item.status)}`]">
                <span class="status-dot"></span>
                {{ item.status }}
              </span>
            </td>
            <td class="size-cell">{{ formatBytes(item.file_size) }}</td>
            <td class="actions-cell text-right">
              <button class="row-action-btn" title="Mở phòng cắt" @click="openEditor(item.video_id)">
                Xem
              </button>
              <button class="row-action-btn btn-danger" title="Xóa log" @click="historyStore.removeItem(item.id)">
                Xóa
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useHistoryStore } from '@/stores/history.store'
import { formatDate, formatBytes } from '@/utils/format'
import AppButton from '@/components/common/AppButton.vue'

const router = useRouter()
const historyStore = useHistoryStore()

function getStatusClass(status: string): string {
  if (status === 'Đã cắt') return 'ready'
  if (status === 'Đang xử lý') return 'processing'
  if (status === 'Thất bại') return 'failed'
  if (status === 'Đã tải xuống') return 'downloaded'
  return 'neutral'
}

function openEditor(videoId: string) {
  if (videoId) {
    router.push(`/editor/${videoId}`)
  } else {
    router.push('/editor')
  }
}
</script>

<style scoped>
.history-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-top-bar {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #FFFFFF;
}

.title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.page-sub {
  font-size: 13px;
  color: var(--text-muted);
}

/* Table Card */
.table-card {
  overflow-x: auto;
  background: #FFFFFF;
}

.history-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.history-table th {
  padding: 12px 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  background: var(--surface-hover);
  border-bottom: 1px solid var(--border);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.history-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
}

.history-table tr:last-child td {
  border-bottom: none;
}

.history-table tr:hover td {
  background: var(--surface-hover);
}

.time-cell {
  color: var(--text-secondary);
  white-space: nowrap;
  font-family: var(--font-mono);
  font-size: 12px;
}

.video-cell {
  max-width: 220px;
}

.video-name {
  font-weight: 600;
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.action-tag {
  background: var(--surface-secondary);
  color: var(--text-secondary);
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.clip-cell {
  max-width: 200px;
}

.clip-text {
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

.size-cell {
  font-family: var(--font-mono);
  color: var(--text-muted);
  font-size: 12px;
}

/* Status Badges */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-badge--ready {
  background: #E6F4FF;
  color: var(--primary);
}
.status-badge--ready .status-dot { background: var(--primary); }

.status-badge--processing {
  background: #FEF3C7;
  color: #D97706;
}
.status-badge--processing .status-dot { background: #D97706; }

.status-badge--downloaded {
  background: #DCFCE7;
  color: var(--success);
}
.status-badge--downloaded .status-dot { background: var(--success); }

.status-badge--failed {
  background: #FEE2E2;
  color: var(--danger);
}
.status-badge--failed .status-dot { background: var(--danger); }

/* Actions */
.text-right {
  text-align: right;
}

.actions-cell {
  white-space: nowrap;
}

.row-action-btn {
  background: transparent;
  color: var(--primary);
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.row-action-btn:hover {
  background: var(--primary-light);
}

.btn-danger {
  color: var(--danger);
}

.btn-danger:hover {
  background: var(--danger-light);
}
</style>
