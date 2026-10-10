<template>
  <div class="url-input">
    <div class="url-input__row">
      <input
        v-model="url"
        class="url-input__field"
        type="url"
        placeholder="https://youtube.com/watch?v=... or any video URL"
        :disabled="loading"
        @keydown.enter="emit('submit', url)"
      />
      <AppButton variant="primary" :loading="loading" @click="emit('submit', url)">
        Get Video
      </AppButton>
    </div>
    <div v-if="error" class="url-input__error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AppButton from '@/components/common/AppButton.vue'

defineProps<{ loading?: boolean; error?: string | null }>()
const emit = defineEmits<{ submit: [url: string] }>()
const url = ref('')
</script>

<style scoped>
.url-input { display: flex; flex-direction: column; gap: 8px; }
.url-input__row { display: flex; gap: 8px; }
.url-input__field {
  flex: 1; padding: 10px 14px;
  background: var(--surface2); border: 1px solid var(--border);
  border-radius: var(--radius); color: var(--text);
  outline: none; transition: border-color 0.15s;
}
.url-input__field:focus { border-color: var(--accent); }
.url-input__field::placeholder { color: var(--text-muted); }
.url-input__field:disabled { opacity: 0.6; }
.url-input__error { color: var(--error); font-size: 13px; }
</style>
