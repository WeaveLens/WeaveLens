<script setup lang="ts">
import { useGraphStore } from '../stores/graph'
import { ref, watch } from 'vue'

const graphStore = useGraphStore()
const localQuery = ref('')

watch(() => graphStore.searchQuery, (val) => {
  localQuery.value = val
})

function onInput() {
  graphStore.setSearch(localQuery.value)
}

function onClear() {
  localQuery.value = ''
  graphStore.setSearch('')
}
</script>

<template>
  <div class="search-bar">
    <input
      v-model="localQuery"
      type="text"
      placeholder="Search resources..."
      @input="onInput"
    />
    <button v-if="localQuery" @click="onClear">Clear</button>
  </div>
</template>

<style scoped>
.search-bar {
  display: flex;
  gap: 8px;
  padding: 8px 16px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
}
.search-bar input {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.search-bar button {
  padding: 6px 12px;
  background: #f5f5f5;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
}
</style>
