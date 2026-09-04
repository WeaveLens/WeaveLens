<script setup lang="ts">
interface Props {
  name: string
  size?: number | string
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 18,
})
</script>

<template>
  <svg
    :width="props.size"
    :height="props.size"
    class="icon"
    :class="props.class"
    role="img"
    :aria-label="props.name"
  >
    <use :href="`/icons.svg#${props.name}`" />
  </svg>
</template>

<style scoped>
.icon {
  display: inline-block;
  flex-shrink: 0;
  vertical-align: middle;
}
</style>
