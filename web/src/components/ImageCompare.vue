<template>
  <div class="image-compare" ref="containerRef" @mousedown="startDrag" @touchstart="startDrag">
    <!-- After (Compressed) - full image underneath -->
    <img
      class="image-after"
      :src="afterSrc"
      :alt="afterLabel"
      @load="onImageLoad"
      draggable="false"
    />

    <!-- Before (Original) - clipped from right side -->
    <img
      class="image-before"
      :src="beforeSrc"
      :alt="beforeLabel"
      :style="{ clipPath: `inset(0 ${100 - sliderPosition}% 0 0)` }"
      draggable="false"
    />

    <!-- Labels -->
    <span class="label-before">{{ beforeLabel }}</span>
    <span class="label-after">{{ afterLabel }}</span>

    <!-- Slider handle -->
    <div class="slider-handle" :style="{ left: sliderPosition + '%' }">
      <div class="slider-line"></div>
      <div class="slider-button">
        <span class="slider-arrows">&lt;&gt;</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    beforeSrc: string
    afterSrc: string
    beforeLabel?: string
    afterLabel?: string
    initialPosition?: number
  }>(),
  {
    beforeLabel: 'Original',
    afterLabel: 'Compressed',
    initialPosition: 50
  }
)

const containerRef = ref<HTMLDivElement | null>(null)
const sliderPosition = ref(props.initialPosition)
const isDragging = ref(false)

function onImageLoad(): void {
  // Image loaded
}

function startDrag(event: MouseEvent | TouchEvent): void {
  event.preventDefault()
  isDragging.value = true
  updateFromEvent(event)
}

function stopDrag(): void {
  isDragging.value = false
}

function updateFromEvent(event: MouseEvent | TouchEvent): void {
  const container = containerRef.value
  if (!container) return

  const rect = container.getBoundingClientRect()
  let clientX: number

  if ('touches' in event && event.touches.length > 0) {
    clientX = event.touches[0].clientX
  } else if ('clientX' in event) {
    clientX = event.clientX
  } else {
    return
  }

  const x = clientX - rect.left
  const percentage = Math.max(0, Math.min(100, (x / rect.width) * 100))
  sliderPosition.value = percentage
}

function handleGlobalMove(event: MouseEvent | TouchEvent): void {
  if (!isDragging.value) return
  updateFromEvent(event)
}

onMounted(() => {
  document.addEventListener('mousemove', handleGlobalMove)
  document.addEventListener('mouseup', stopDrag)
  document.addEventListener('touchmove', handleGlobalMove)
  document.addEventListener('touchend', stopDrag)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleGlobalMove)
  document.removeEventListener('mouseup', stopDrag)
  document.removeEventListener('touchmove', handleGlobalMove)
  document.removeEventListener('touchend', stopDrag)
})
</script>

<style scoped>
.image-compare {
  position: relative;
  width: 100%;
  overflow: hidden;
  border-radius: 8px;
  cursor: ew-resize;
  user-select: none;
  background: #1f2937;
  line-height: 0;
}

/* Both images same size, stacked */
.image-compare img {
  display: block;
  width: 100%;
  height: auto;
  max-height: 80vh;
  object-fit: contain;
  pointer-events: none;
}

.image-after {
  position: relative;
}

.image-before {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.label-before,
.label-after {
  position: absolute;
  bottom: 12px;
  padding: 6px 14px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  font-size: 13px;
  font-weight: 600;
  border-radius: 6px;
  pointer-events: none;
  white-space: nowrap;
  z-index: 20;
  line-height: 1;
}

.label-before {
  left: 12px;
}

.label-after {
  right: 12px;
}

.slider-handle {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 4px;
  z-index: 10;
  transform: translateX(-50%);
  pointer-events: none;
}

.slider-line {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 3px;
  background: white;
  transform: translateX(-50%);
  box-shadow: 0 0 8px rgba(0, 0, 0, 0.5);
}

.slider-button {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 44px;
  height: 44px;
  background: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.4);
  color: #374151;
  font-weight: bold;
}

.slider-arrows {
  font-size: 16px;
  letter-spacing: 3px;
}
</style>